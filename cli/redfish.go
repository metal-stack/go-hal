package main

import (
	"fmt"
	"io"

	"github.com/stmcginnis/gofish"
	"github.com/urfave/cli/v2"
)

var redfishCmd = &cli.Command{
	Name:        "redfish",
	Usage:       "raw redfish usage",
	Description: "for example use --redfish-path /redfish/v1/Chassis/System.Embedded.1",
	Flags:       append(flags, redfishPathFlag),
	Action: func(ctx *cli.Context) error {

		config := gofish.ClientConfig{
			Endpoint: "https://" + host,
			Username: user,
			Password: password,
			Insecure: true,
		}
		c, err := gofish.Connect(config)
		if err != nil {
			return err
		}

		resp, err := c.Get(redfishPath)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Println(string(body))
		return nil
	},
}
