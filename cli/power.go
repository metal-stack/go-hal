package main

import (
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/metal-stack/go-hal"
	"github.com/urfave/cli/v2"
)

var powerCmd = &cli.Command{
	Name:  "power",
	Usage: "gather and modify machine power state",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		c, err := getHalConnection(log)
		if err != nil {
			return err
		}
		state, err := c.PowerState()
		if err != nil {
			return err
		}
		log.Infow("power", "state", state.String())
		return nil
	},
	Subcommands: []*cli.Command{
		{
			Name:  "on",
			Usage: "power machine on",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.PowerOn()
				if err != nil {
					return err
				}
				err = retry.Do(func() error {
					state, err := c.PowerState()
					if err != nil {
						return err
					}
					log.Infow("power", "state", state.String())
					if state != hal.PowerOnState {
						return fmt.Errorf("state is still not %s", hal.PowerOnState.String())
					}
					return nil
				})
				if err != nil {
					return err
				}
				state, err := c.PowerState()
				if err != nil {
					return err
				}
				log.Infow("power", "state", state.String())
				return nil
			},
		},
		{
			Name:  "off",
			Usage: "power machine off",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.PowerOff()
				if err != nil {
					return err
				}
				err = retry.Do(func() error {
					state, err := c.PowerState()
					if err != nil {
						return err
					}
					log.Infow("power", "state", state.String())
					if state != hal.PowerOffState {
						return fmt.Errorf("state is still not %s", hal.PowerOffState.String())
					}
					return nil
				})
				if err != nil {
					return err
				}
				state, err := c.PowerState()
				if err != nil {
					return err
				}
				log.Infow("power", "state", state.String())
				return nil
			},
		},
		{
			Name:  "cycle",
			Usage: "power cycle machine",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				state, err := c.PowerState()
				if err != nil {
					return err
				}
				log.Infow("power", "state", state.String())

				err = c.PowerCycle()
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "reset",
			Usage: "power reset machine",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				state, err := c.PowerState()
				if err != nil {
					return err
				}
				log.Infow("power", "state", state.String())

				err = c.PowerReset()
				if err != nil {
					return err
				}
				return nil
			},
		},
	},
}
