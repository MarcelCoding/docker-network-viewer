package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Network struct {
	name  string
	ipNet []net.IPNet
}

func main() {
	networks := GetNetworks()

	// sort networks by each octet of the ipv4 address
	sort.Slice(networks, func(i, j int) bool {
		return bytes.Compare(networks[i].ipNet[0].IP, networks[j].ipNet[0].IP) < 0
	})

	PrintNetworks(networks)
}

func GetNetworks() []Network {
	docker, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	networkList, err := docker.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var networks []Network

	for _, network := range networkList {
		config := network.IPAM.Config

		if len(config) < 1 {
			continue
		}

		var ipNets []net.IPNet

		for _, conf := range config {
			_, ipNet, err := net.ParseCIDR(conf.Subnet)
			if err != nil {
				panic(err)
			}

			ipNets = append(ipNets, *ipNet)
		}

		networks = append(
			networks,
			Network{
				name:  network.Name,
				ipNet: ipNets,
			},
		)
	}

	return networks
}

func PrintNetworks(networks []Network) {
	writer := new(tabwriter.Writer)

	writer.Init(os.Stdout, 2, 8, 8, '\t', 0)
	defer writer.Flush()

	for _, network := range networks {
		fmt.Fprintf(writer, "%s\t%s\n", network.name, PrintIpNets(network.ipNet))
	}
}

func PrintIpNets(nets []net.IPNet) string {
	res := make([]string, len(nets))
	for i := 0; i < len(nets); i++ {
		res[i] = nets[i].String()
	}
	return strings.Join(res, ", ")
}
