# go-hal a hardware abstraction layer for servers

[![GoDoc](https://godoc.org/github.com/metal-stack/go-hal?status.svg)](https://pkg.go.dev/github.com/metal-stack/go-hal)

go server hardware abstraction, tries to lower the burden of supporting different server vendors.

Example usage:

```golang
package main

import (
    "fmt"
    "github.com/metal-stack/go-hal/detect"
)

func main() {
    smcInBand, err := detect.ConnectInBand()
    if err != nil {
        panic(err)
    }

    firmware, err := smcInBand.Firmware()
    if err != nil {
        panic(err)
    }
    fmt.Println(firmware)
    // UEFI

    err = smcInBand.PowerOff()
}
```

## Dell

Sample Lab Machines
172.19.100.107
User: root
Default PW: 9XW9FR9PN3FF

172.19.100.108
User: root
Default PW: K4P4NVAK9KVK

TODO:

- Console access must be switched to ssh root@<IP> console com2
  Can password access switched to pubkey ?

- Identify LED ON/OFF and State does not work

- IPMI is disabled by default, unsure if Redfish Password can be set from Inband

- Redfish Sessions must be closed, see: https://github.com/stmcginnis/gofish/blob/main/examples/query_sessions.md

- Dell iDrac Redfish Samples: https://github.com/dell/iDRAC-Redfish-Scripting