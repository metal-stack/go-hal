package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/metal-stack/go-hal"

	"github.com/metal-stack/go-hal/connect"
	"github.com/metal-stack/go-hal/pkg/logger"
)

var (
	band     = flag.String("bandtype", "outband", "inband/outband")
	user     = flag.String("user", "ADMIN", "bmc username")
	password = flag.String("password", "ADMIN", "bmc password")
	host     = flag.String("host", "localhost", "bmc host")
	port     = flag.Int("port", 623, "bmc port")

	errHelp = errors.New("usage: -bandtype inband|outband")
)

func main() {
	flag.Parse()

	log := logger.New()
	switch *band {
	case "inband":
		inband(log)
	case "outband":
		outband(log)
	default:
		fmt.Printf("%s\n", errHelp)
		os.Exit(1)
	}
}

func inband(log logger.Logger) {
	ib, err := connect.InBand(log)
	if err != nil {
		panic(err)
	}
	uuid, err := ib.UUID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("UUID:%s\n", uuid)
}

func outband(log logger.Logger) {
	ob, err := connect.OutBand(*host, *port, *user, *password, log)
	if err != nil {
		panic(err)
	}

	uu := make(map[string]string)
	ee := make(map[string]error)

	b := ob.Board()
	fmt.Printf("Board:\n%#v\n", b)

	bmc, err := ob.BMCConnection().BMC()
	if err != nil {
		ee["BMCConnection.BMC"] = err
	}
	fmt.Printf("BMC:\n%#v\n", bmc)

	_, err = ob.UUID()
	if err != nil {
		ee["UUID"] = err
	}

	ps, err := ob.PowerState()
	if err != nil {
		ee["PowerState"] = err
	}
	if ps == hal.PowerUnknownState {
		uu["PowerState"] = "unexpected power state: PowerUnknownState"
	}

	// ipmitool sel
	err = ob.IdentifyLEDState(hal.IdentifyLEDStateOff)
	if err != nil {
		ee["IdentifyLEDState"] = err
	}
	err = ob.IdentifyLEDOn()
	if err != nil {
		ee["IdentifyLEDOn"] = err
	}
	err = ob.IdentifyLEDOff()
	if err != nil {
		ee["IdentifyLEDOff"] = err
	}

	//_, err = ob.UpdateBIOS()
	//if err != nil {
	//	ee["UpdateBIOS"] = err
	//}
	//
	//_, err = ob.UpdateBMC()
	//if err != nil {
	//	ee["UpdateBMC"] = err
	//}

	if len(uu) > 0 {
		fmt.Println("Unexpected things:")
		for m, u := range uu {
			fmt.Printf("%s: %s\n", m, u)
		}
	}

	if len(ee) > 0 {
		fmt.Println("Failed checks:")
		for m, err := range ee {
			fmt.Printf("%s: %s\n", m, err.Error())
		}
	} else {
		fmt.Println("Check succeeded")
	}
}
