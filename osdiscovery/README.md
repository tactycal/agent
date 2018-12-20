### About

Package `osdiscovery` implements functions to get basic operating system
identification data for Linux operating system.
The following informations are provided:

* Distribution name
* Release version
* Architecture
* Fully qualified domain name
* Kernel release

### How to

```go
package main

import (
     "fmt"
     "github.com/tactycal/agent/osdiscovery"
)

func main() {
    osInfo, err := osdiscovery.Get()
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("%+v\n", osInfo)
    }
}
```
