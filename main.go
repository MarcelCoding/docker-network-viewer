package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"net"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type network struct {
	name  string
	ipNet *net.IPNet
}

func main() {
	grp, err := user.LookupGroup("docker")
	if err != nil {
		panic(err)
	}

	// require docker group or root access
	if !CheckGroups(*grp) && strings.TrimSpace(GetProcessOwner()) != "root" {
		fmt.Println("DNV requires docker access!")
		fmt.Println("Assign yourself the docker group or execute this process as root. Exiting...")
		os.Exit(1)
	}

	// create docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// get docker networks
	networkResources, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	// parse networks
	var networks []network
	for _, obj := range networkResources {
		config := obj.IPAM.Config

		if len(config) < 1 {
			continue
		}
		_, ipNet, err := net.ParseCIDR(config[0].Subnet)
		if err != nil {
			panic(err)
		}

		networks = append(networks, network{
			name:  obj.Name,
			ipNet: ipNet,
		})
	}

	// sort networks by each octet of the ipv4 address
	sort.Slice(networks, func(i, j int) bool {
		return bytes.Compare(networks[i].ipNet.IP, networks[j].ipNet.IP) < 0
	})

	// write with padding
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 2,8, 8, '\t', 0)
	defer writer.Flush()
	for _, obj := range networks {
		_, _ = fmt.Fprintf(writer, "%s\t%s\n", obj.name, obj.ipNet)
	}
}

// Get the user that started this process
func GetProcessOwner() string {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return string(stdout)
}

// check if the user that started this process is inside a group (e.g. the docker group)
func CheckGroups(grp user.Group) bool {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	gids, err := usr.GroupIds()
	if err != nil {
		panic(err)
	}
	found := false
	for _, gid := range gids {
		if gid == grp.Gid {
			found = true
			break
		}
	}
	return found
}