package main

import (
	"fmt"
	"os"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/connect"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/urfave/cli/v2"
)

var (
	log logger.Logger

	bandtype string
	user     string
	password string
	host     string
	port     int

	bandtypeFlag = &cli.StringFlag{
		Name:        "bandtype",
		Value:       "outband",
		Usage:       "inband/outband",
		Destination: &bandtype,
	}
	userFlag = &cli.StringFlag{
		Name:        "user",
		Value:       "ADMIN",
		Usage:       "bmc user",
		Destination: &user,
	}
	passwordFlag = &cli.StringFlag{
		Name:        "password",
		Value:       "",
		Usage:       "bmc password",
		Destination: &password,
	}
	hostFlag = &cli.StringFlag{
		Name:        "host",
		Value:       "localhost",
		Usage:       "bmc host",
		Destination: &host,
	}
	portFlag = &cli.IntFlag{
		Name:        "port",
		Value:       623,
		Usage:       "bmc port",
		Destination: &port,
	}
	flags = []cli.Flag{
		bandtypeFlag,
		hostFlag,
		portFlag,
		userFlag,
		passwordFlag,
	}
	outBandBMCConnection api.OutBandBMCConnection
)

func main() {
	log = logger.New()

	app := &cli.App{
		Name:  "hal",
		Usage: "try bmc commands",
		Commands: []*cli.Command{
			uuidCmd,
			boardCmd,
			ledCmd,
			powerCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

	if outBandBMCConnection != nil {
		outBandBMCConnection.Close()
	}

	os.Exit(1)
}

func getHalConnection(log logger.Logger) (hal.Hal, error) {
	switch bandtype {
	case "inband":
		ib, err := connect.InBand(log)
		if err != nil {
			return nil, err
		}
		return ib, nil
	case "outband":
		ob, err := connect.OutBand(host, port, user, password, log)
		if err != nil {
			return nil, err
		}
		outBandBMCConnection = ob.BMCConnection()
		return ob, nil
	default:
		return nil, fmt.Errorf("unknown bandtype %s", bandtype)
	}
}

// func outband(log logger.Logger) {
// 	ob, err := connect.OutBand(*host, *port, *user, *password, log)
// 	if err != nil {
// 		panic(err)
// 	}

// 	uu := make(map[string]string)
// 	ee := make(map[string]error)

// 	b := ob.Board()
// 	fmt.Printf("Board:\n%#v\n", b)
// 	fmt.Printf("Power:\n%#v\n", b.PowerMetric)
// 	fmt.Printf("PowerSupplies:\n%#v\n", b.PowerSupplies)

// 	bmc, err := ob.BMCConnection().BMC()
// 	if err != nil {
// 		ee["BMCConnection.BMC"] = err
// 	}
// 	fmt.Printf("BMC:\n%#v\n", bmc)

// 	_, err = ob.UUID()
// 	if err != nil {
// 		ee["UUID"] = err
// 	}

// 	ps, err := ob.PowerState()
// 	if err != nil {
// 		ee["PowerState"] = err
// 	}
// 	if ps == hal.PowerUnknownState {
// 		uu["PowerState"] = "unexpected power state: PowerUnknownState"
// 	}
// 	err = ob.PowerOff()
// 	if err != nil {
// 		fmt.Printf("error during power off: %v\n", err)
// 	}

// 	board := ob.Board()
// 	fmt.Println("LED: " + board.IndicatorLED)

// 	err = ob.PowerCycle()
// 	if err != nil {
// 		ee["PowerCycle"] = err
// 	}

// 	board = ob.Board()
// 	fmt.Println("LED: " + board.IndicatorLED)

// 	if false {
// 		err = ob.PowerOff()
// 		if err != nil {
// 			fmt.Printf("error during power off: %v\n", err)
// 		}

// 		time.Sleep(10 * time.Second)

// 		err = ob.PowerOn()
// 		if err != nil {
// 			fmt.Printf("error during power on: %v\n", err)
// 		}

// 		// ipmitool sel
// 		err = ob.IdentifyLEDState(hal.IdentifyLEDStateOff)
// 		if err != nil {
// 			ee["IdentifyLEDState"] = err
// 		}
// 		err = ob.IdentifyLEDOn()
// 		if err != nil {
// 			ee["IdentifyLEDOn"] = err
// 		}
// 		err = ob.IdentifyLEDOff()
// 		if err != nil {
// 			ee["IdentifyLEDOff"] = err
// 		}

// 		//_, err = ob.UpdateBIOS()
// 		//if err != nil {
// 		//	ee["UpdateBIOS"] = err
// 		//}
// 		//
// 		//_, err = ob.UpdateBMC()
// 		//if err != nil {
// 		//	ee["UpdateBMC"] = err
// 		//}
// 	}

// 	if len(uu) > 0 {
// 		fmt.Println("Unexpected things:")
// 		for m, u := range uu {
// 			fmt.Printf("%s: %s\n", m, u)
// 		}
// 	}

// 	if len(ee) > 0 {
// 		fmt.Println("Failed checks:")
// 		for m, err := range ee {
// 			fmt.Printf("%s: %s\n", m, err.Error())
// 		}
// 	} else {
// 		fmt.Println("Check succeeded")
// 	}
// }
