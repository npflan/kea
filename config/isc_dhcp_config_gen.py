import csv
import ipaddress
import sys
import os
import urllib.request
import io
import pathlib
import json


def subnet(row):
    if row['role'].casefold() not in ['access', 'wireless', 'management  netværk', 'management netværk', 'cctv']:
        return
    if row['description'].casefold() in ['wireless networks']:
        return
    ip = ipaddress.IPv4Network(row['prefix'])

    poolstart = 4
    if ip.subnet_of(ipaddress.IPv4Network('10.255.0.0/16')) or \
            ip.subnet_of(ipaddress.IPv4Network('172.20.0.0/16')):
        poolstart = 100
    elif ip.subnet_of(ipaddress.IPv4Network('10.248.0.0/16')):
        poolstart = 6

    if ip.prefixlen > 24:
        return
    return {
        "subnet": ip.with_prefixlen,
        "pools": [
          {
              "pool":  str(ip[poolstart]) + "-" + str(ip[pow(2, (32-ip.prefixlen))-6])
          }
        ],
        "relay": {
            "ip-address": str(ip[1])
        },
        "option-data": [
            {
                "name": "routers",
                "data": str(ip[1])
            }
        ]
    }

datafile = pathlib.Path(os.path.dirname(__file__), 'data.csv')
if not datafile.exists() or not datafile.is_file():
    netbox = 'https://netbox.minserver.dk/ipam/prefixes/?status=1&parent=&family=&q=&vrf=npflan&mask_length=&export'
    data = urllib.request.urlopen(netbox).read()
    with open(datafile, 'wb+') as f:
        f.write(data)
else:
    data = datafile.read_bytes()

reader = csv.DictReader(io.StringIO(data.decode()),
                        delimiter=',', quotechar='|')
print('"subnet4":')
print(
    json.dumps(
        list(filter(None, (subnet(row) for row in reader))),
        indent=True
    )
)
