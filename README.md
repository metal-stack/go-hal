# go-hal a hardware abstraction layer for servers

go server hardware abstraction, tries to lower the burden of supporting different server vendors.

Example usage:

```golang
package main

import (
    "fmt"
    smc "github.com/metal-stack/go-hal/supermicro"
)

func main() {
    smcInBand,err := smc.InBand()
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
