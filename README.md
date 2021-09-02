# Docker Network Viewer

Liste docker networks and according subnet.

## Setup

1. Setup my Debian Repository: [deb.m4rc3l.de](https://deb.m4rc3l.de/)
2. `sudo apt install docker-network-viewer`
3. Run `dnv` or `docker-network-viewer`

## Credits

I have been inspired by @felbinger who created this python script, he also made a go script, but decided to delete his repository:
```python
#!/usr/bin/python3.8

from docker import from_env as docker_env
from prettytable import PrettyTable
from os import geteuid


def get_networks():
    networks = list()
    for network in docker_env().networks.list():
        if network:
            config = network.attrs.get('IPAM').get('Config')
            subnet = config[0].get('Subnet') if len(config) else None
            if subnet:
                networks.append([
                    network.attrs.get('Name'),
                    subnet
                ])
    return networks


if __name__ == "__main__":
    if geteuid() != 0:
        print("Please run the script as root!")
        exit(1)

    networks = sorted(get_networks(), key=lambda ip: list(map(int, ip[1].split("/")[0].split("."))))

    table = PrettyTable()
    table.field_names = ["Name", "Subnet"]
    for row in networks:
        table.add_row(row)

    table.align['Name'] = 'l'
    table.align['Subnet'] = 'l'
    print(table)
```
