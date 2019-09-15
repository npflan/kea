package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type KeaRequest struct {
	Command string   `json:"command"`
	Service []string `json:"service"`
}

type KeaDHCP4 struct {
	Subnet4 []KeaSubnet4 `json:"subnet4"`
}

type KeaSubnet4 struct {
	ID     int    `json:"id"`
	Subnet string `json:"subnet"`
}

type KeaArgument struct {
	DHCP4     *KeaDHCP4     `json:"Dhcp4"`
	ResultSet *KeaResultSet `json:"result-set"`
}

type KeaResultSet struct {
	Columns []string `json:"columns"`
	Rows    []KeaRow `json:"rows"`
}

type KeaRow []int

type KeaResponse struct {
	Arguments KeaArgument `json:"arguments"`
	Result    int         `json:"result"`
}

type KeaConfig map[int]net.IPNet

type Config struct {
	PromPort  string        `required:"true" default:"9405"`
	LogLevel  zapcore.Level `required:"true" default:"info"`
	KeaNetLoc string        `required:"true" default:"http://localhost:8080"`
}

type keaCollector struct {
	keaNetLoc  string
	keaConfig  *KeaConfig
	keaConfigM sync.Mutex
}

func initKeaCollector(conf Config) *keaCollector {
	return &keaCollector{
		keaNetLoc: conf.KeaNetLoc,
	}
}

func (c *keaCollector) KeaCommand(cmd string) (*KeaArgument, error) {
	request := KeaRequest{
		Command: cmd,
		Service: []string{"dhcp4"},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return nil, err
	}
	res, err := http.Post(
		c.keaNetLoc,
		"application/json",
		&buf,
	)
	defer func() {
		_ = res.Body.Close()
	}()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		zap.L().Error("http error reply", zap.String("status", res.Status), zap.ByteString("body", body))
		return nil, fmt.Errorf("bad reply %q: %q", res.Status, body)
	}
	responses := make([]KeaResponse, 0, 1)
	if err := json.NewDecoder(res.Body).Decode(&responses); err != nil {
		return nil, err
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("%s returned empty results", cmd)
	}
	response := responses[0]
	return &response.Arguments, nil
}

func (c *keaCollector) GetConfig() (KeaConfig, error) {
	zap.L().Info("getting config from kea")
	arguments, err := c.KeaCommand("config-get")
	if err != nil {
		return nil, err
	}
	if arguments.DHCP4 == nil {
		return nil, fmt.Errorf("config returned unexpected")
	}
	dhcp4 := arguments.DHCP4

	conf := make(KeaConfig)
	for _, res := range dhcp4.Subnet4 {
		_, cidr, err := net.ParseCIDR(res.Subnet)
		if err != nil {
			return nil, err
		}
		conf[res.ID] = *cidr
	}
	return conf, nil
}

func (c *keaCollector) UpdateConfig() error {
	// Get config
	s, err := c.GetConfig()
	if err != nil {
		return err
	}
	c.keaConfigM.Lock()
	c.keaConfig = &s
	c.keaConfigM.Unlock()
	return nil
}

func (c *keaCollector) Describe(ch chan<- *prometheus.Desc) {

}

func (c *keaCollector) Collect(ch chan<- prometheus.Metric) {
	if c.keaConfig == nil {
		err := c.UpdateConfig()
		if err != nil {
			zap.L().Warn("could not get config", zap.Error(err))
			return
		}
	}
	// Copy to local variable
	c.keaConfigM.Lock()
	subnets := make(KeaConfig)
	for sid, snet := range *c.keaConfig {
		subnets[sid] = snet
	}
	c.keaConfigM.Unlock()

	arguments, err := c.KeaCommand("stat-lease4-get")
	if err != nil {
		zap.L().Warn("could not get stats", zap.Error(err))
		return
	}
	if arguments.ResultSet == nil {
		zap.L().Error("stat-lease4-get returned unexpected")
		return
	}

	results := arguments.ResultSet

	metrics := make([]*prometheus.Desc, 0, len(results.Columns)-1)

	idOffset := -1
	for i, c := range results.Columns {
		switch c {
		case "subnet-id":
			idOffset = i
		default:
			name := fmt.Sprintf("kea_lease_%s", strings.ReplaceAll(c, "-", "_"))
			d := prometheus.NewDesc(name, name, []string{"subnet_id", "subnet_desc"}, nil)
			metrics = append(metrics, d)
		}
	}
	if idOffset < 0 {
		zap.L().Error("no subnet-id in columns", zap.Strings("columns", results.Columns))
		return
	}

	for _, row := range results.Rows {
		subnetID := row[idOffset]
		subnet, subnetOK := subnets[subnetID]
		if !subnetOK {
			zap.L().Info("missing config for subnet", zap.Int("id", subnetID))
			err := c.UpdateConfig()
			if err != nil {
				zap.L().Warn("could not get config", zap.Error(err))
				return
			}
			continue
		}
		if len(row) != len(results.Columns) {
			zap.L().Error("mismatch column and row length")
			continue
		}
		mix := 0
		for i, stat := range row {
			if i == idOffset {
				continue
			}
			met, err := prometheus.NewConstMetric(
				metrics[mix], prometheus.GaugeValue, float64(stat),
				strconv.FormatInt(int64(subnetID), 10), subnet.String(),
			)
			mix++
			if err != nil {
				zap.L().Warn("failed to create metric", zap.Error(err))
			} else {
				ch <- met
			}
		}
	}
}

func do() error {
	var conf Config
	err := envconfig.Process("DHCP", &conf)
	if err != nil {
		return err
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)

	collect := initKeaCollector(conf)
	prometheus.MustRegister(collect)

	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":"+conf.PromPort, nil)
}

func main() {
	if err := do(); err != nil {
		panic(err)
	}
}
