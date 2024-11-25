package main

import (
	"github.com/urfave/cli/v2"
)

var uuidCmd = &cli.Command{
	Name:  "uuid",
	Usage: "gather machine uuid",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		c, err := getHalConnection(log)
		if err != nil {
			return err
		}
		uid, err := c.UUID()
		if err != nil {
			return err
		}
		log.Infow("uuid", "bandtype", bandtype, "host", host, "result", uid.String())
		return nil
	},
}
