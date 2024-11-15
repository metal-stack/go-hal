package main

import (
	"github.com/urfave/cli/v2"
)

var ledCmd = &cli.Command{
	Name:  "led",
	Usage: "gather machine led state",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		// c, err := getHalConnection(log)
		// if err != nil {
		// 	return err
		// }
		// ledstate := c.IdentifyLEDState()
		// log.Infow("board", "bandtype", bandtype, "host", host, "result", board.String(), "bios", board.BIOS, "powermetric", board.PowerMetric, "powersupplies", board.PowerSupplies)
		return nil
	},
}
