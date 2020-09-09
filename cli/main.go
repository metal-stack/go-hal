package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/metal-stack/go-hal/connect"
)

var (
	band    = flag.String("bandtype", "inband", "inband/outband")
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
	outband, err := connect.OutBand("10.5.2.93", 623, "ADMIN", "ADMIN")
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
}
