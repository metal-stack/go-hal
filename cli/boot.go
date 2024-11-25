package main

import (
	"github.com/metal-stack/go-hal"
	"github.com/urfave/cli/v2"
)

var bootCmd = &cli.Command{
	Name:  "boot",
	Usage: "gather and modify boot order",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		log.Warnw("getting boot order missing")
		return nil
	},
	Subcommands: []*cli.Command{
		{
			Name:  "hdd",
			Usage: "boot from hdd",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.BootFrom(hal.BootTargetDisk)
				if err != nil {
					return err
				}
				log.Infow("boot", "set to", hal.BootTargetDisk.String())
				return nil
			},
		},
		{
			Name:  "pxe",
			Usage: "boot from pxe",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.BootFrom(hal.BootTargetPXE)
				if err != nil {
					return err
				}
				log.Infow("boot", "set to", hal.BootTargetPXE.String())
				return nil
			},
		},
		{
			Name:  "bios",
			Usage: "boot to bios",
			Flags: flags,
			Action: func(ctx *cli.Context) error {
				c, err := getHalConnection(log)
				if err != nil {
					return err
				}
				err = c.BootFrom(hal.BootTargetBIOS)
				if err != nil {
					return err
				}
				log.Infow("boot", "set to", hal.BootTargetBIOS.String())
				return nil
			},
		},
	},
}
