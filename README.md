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
