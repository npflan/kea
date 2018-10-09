from prometheus_client.core import (
    GaugeMetricFamily, REGISTRY
)
from ipaddress import IPv4Network, IPv4Address
import prometheus_client
import requests

class DhcpdCollector(object):
    def __init__(self):
        # Get config with subnet names
        self.getsubnets()

    def getsubnets(self):
        r = requests.post('http://localhost:8080/', 
            json={"command": 'config-get', "service": ["dhcp4"]}
        )
        r.raise_for_status()
        config = r.json()
        subnetlist = config[0]['arguments']['Dhcp4']['subnet4']
        self.subnets = {a['id']: IPv4Network(a['subnet']) for a in subnetlist}

    def collect(self):
        r = requests.post('http://localhost:8080/',
            json={"command": "stat-lease4-get", "service": ["dhcp4"]}
        )
        r.raise_for_status()
        stats = r.json()
        resultset = stats[0]['arguments']['result-set']
        subnetstats = (dict(zip(resultset['columns'], row)) for row in resultset['rows'])

        metrics = {
            c: GaugeMetricFamily(f'kea_lease_{c.replace("-","_")}', f'{c} in subnet', labels=['subnet_id','subnet_desc'])
            for c in resultset['columns'] if c != 'subnet-id'
        }
        for stat in subnetstats:
            subnet_id = stat['subnet-id']
            subnet_descr = self.subnets[subnet_id]
            for statname, value in stat.items():
                if statname == 'subnet-id':
                    continue
                metrics[statname].add_metric([str(subnet_id),subnet_descr.with_prefixlen], value)
        yield from metrics.values()

# For easy debugging
if __name__ == '__main__':
    # port forward kea control socket
    # kubectl port-forward dhcp-secondary-7dc4f89c94-7lw7s -n dhcp 8080:8080
    from prometheus_client.exposition import generate_latest
    print(generate_latest(REGISTRY).decode('utf-8'))

# For running with a wsgi server like gunicorn
REGISTRY.register(DhcpdCollector())
app = prometheus_client.make_wsgi_app()