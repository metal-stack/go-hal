package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/metal-stack/go-hal/connect"
)

var (
	band     = flag.String("bandtype", "inband", "inband/outband")
	user     = flag.String("user", "ADMIN", "bmc username")
	password = flag.String("password", "ADMIN", "bmc password")
	host     = flag.String("host", "localhost", "bmc host")
	port     = flag.Int("port", 623, "bmc port")

	errHelp = errors.New("usage: -bandtype inband|outband")
)

func main() {
	flag.Parse()
	switch *band {
	case "inband":
		fmt.Printf("inband test\n")
		inband()
	case "outband":
		fmt.Printf("outband test\n")
		outband()
	default:
		fmt.Printf("%s\n", errHelp)
		os.Exit(1)
	}
}

func inband() {
	inband, err := connect.InBand()
	if err != nil {
		panic(err)
	}
	uuid, err := inband.UUID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("UUID:%s\n", uuid)
}

func outband() {
	outband, err := connect.OutBand(*host, *port, *user, *password)
	if err != nil {
		panic(err)
	}
	uuid, err := outband.UUID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("UUID:%s\n", uuid)
	ps, err := outband.PowerState()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Powerstate:%s\n", ps)

	bmc, err := outband.BMCConnection().BMC()
	if err != nil {
		panic(err)
	}
	fmt.Printf("BMC:%s\n", bmc)
}
