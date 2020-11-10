package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/metal-stack/go-hal/connect"
	"github.com/metal-stack/go-hal/internal/logger"
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

	log := logger.NewLogger(logger.Configuration{ConsoleLevel: "DEBUG"}, logger.Log15Logger)
	switch *band {
	case "inband":
		inband(log)
	case "outband":
		outband(log)
	default:
		log.Infof("%s\n", errHelp)
		os.Exit(1)
	}
}

func inband(log logger.Logger) {
	inband, err := connect.InBand(log)
	if err != nil {
		panic(err)
	}
	uuid, err := inband.UUID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("UUID:%s\n", uuid)
}

func outband(log logger.Logger) {
	outband, err := connect.OutBand(*host, *port, *user, *password, log)
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
