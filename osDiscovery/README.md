### About

Package `osDiscovery` implements functions to get basic operating system 
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
     "github.com/tactycal/osDiscovery"
)

func main() {
    osInfo := osDiscovery.Get()
    fmt.Printf("%+v\n", osInfo)
}
```
