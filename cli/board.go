package main

import (
	"github.com/urfave/cli/v2"
)

var boardCmd = &cli.Command{
	Name:  "board",
	Usage: "gather machine board details",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		c, err := getHalConnection(log)
		if err != nil {
			return err
		}
		board, err := c.Board()
		if err != nil {
			return err
		}
		log.Infow("board", "bandtype", bandtype, "host", host, "result", board.String(), "redfish version", board.RedfishVersion, "bios", board.BiosVersion, "powermetric", board.PowerMetric, "powersupplies", board.PowerSupplies, "ledstate", board.IndicatorLED)
		bmc, err := outBandBMCConnection.BMC()
		if err != nil {
			return err
		}
		log.Infow("bmc", "result", bmc)
		return nil
	},
}