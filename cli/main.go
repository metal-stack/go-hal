package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/connect"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/urfave/cli/v2"
)

var (
	log logger.Logger

	bandtype    string
	user        string
	password    string
	host        string
	port        int
	redfishPath string

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
	redfishPathFlag = &cli.StringFlag{
		Name:        "redfish-path",
		Value:       "",
		Usage:       "redfish raw path",
		Destination: &redfishPath,
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
	log = logger.NewSlog(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	app := &cli.App{
		Name:  "hal",
		Usage: "try bmc commands",
		Commands: []*cli.Command{
			uuidCmd,
			boardCmd,
			ledCmd,
			powerCmd,
			redfishCmd,
		},
		Flags: flags,
	}

	log.Infow("go hal cli", "host", host, "port", port, "password", password, "bandtype", bandtype)

	if err := app.Run(os.Args); err != nil {
		log.Errorw("execution failed", "error", err)
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
