package main

import (
	"github.com/metal-stack/go-hal"
	"github.com/urfave/cli/v2"
)

var ledCmd = &cli.Command{
	Name:  "led",
	Usage: "gather machine led state",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		c, err := getHalConnection(log)
		if err != nil {
			return err
		}
		ledstate, err := c.GetIdentifyLED()
		if err != nil {
			return err
		}
		log.Infow("led", "state", ledstate.String())
		return nil
	},
	Subcommands: []*cli.Command{
		{
			Name:  "on",
			Usage: "identify led on",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.IdentifyLEDState(hal.IdentifyLEDStateOn)
				if err != nil {
					return err
				}
				log.Infow("led state set to on")
				ledstate, err := c.GetIdentifyLED()
				if err != nil {
					return err
				}
				log.Infow("led", "state", ledstate.String())
				return nil
			},
		},
		{
			Name:  "off",
			Usage: "identify led off",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.IdentifyLEDState(hal.IdentifyLEDStateOff)
				if err != nil {
					return err
				}
				log.Infow("led state set to off")
				ledstate, err := c.GetIdentifyLED()
				if err != nil {
					return err
				}
				log.Infow("led", "state", ledstate.String())
				return nil
			},
		},
	},
}
