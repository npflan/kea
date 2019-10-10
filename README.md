# kea
Kea DHCP server

### Update config

```bash
mv config/data.csv config/data.csv.bak # force download
python3 config/isc_dhcp_config_gen.py > subnet.conf
kubectl create configmap keasubnet -n dhcp --from-file=subnet.conf --dry-run=true -o yaml > subnet.yaml
```

### Apply config

Log into each pod and run. Remember to verify config map update has synced to the pod.

```bash
kill -HUP 1
```

