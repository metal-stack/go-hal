package main

import (
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/stmcginnis/gofish"
)

func main() {
	log := slog.Default()
	ip := os.Args[1]
	pw := os.Args[2]
	user := "root"

	config := gofish.ClientConfig{
		Endpoint: "https://" + ip,
		Username: user,
		Password: pw,
		Insecure: true,
	}
	c, err := gofish.Connect(config)
	if err != nil {
		panic(err)
	}
	defer c.Logout()

	systems, err := c.Service.Systems()
	if err != nil {
		panic(err)
	}

	for _, system := range systems {
		log.Info("System", "indicator led", system.IndicatorLED, "locator led", system.LocationIndicatorActive)
		system.LocationIndicatorActive = false
		log.Info("System", "system", string(system.RawData))
	}

	systems, err = c.Service.Systems()
	if err != nil {
		panic(err)
	}
	for _, system := range systems {
		log.Info("System", "indicator led", system.IndicatorLED, "locator led", system.LocationIndicatorActive)
	}

	chassis, err := c.Service.Chassis()
	if err != nil {
		panic(err)
	}

	for _, chass := range chassis {
		log.Info("Chassis", "indicator led", chass.IndicatorLED, "locator led", chass.LocationIndicatorActive)
	}

	uid := ""
	for _, system := range systems {
		if system.UUID != "" {
			uid = system.UUID
			break
		}
	}

	u, err := uuid.Parse(uid)
	if err != nil {
		panic(err)
	}

	log.Info("BMC", "UUID", u.String())
}
