### About

Package `packagelookup` provides function to get list of installed packages.

Supported operating systems:

* Ubuntu
* Debian
* Red Hat Enterprise Linux
* CentOS
* Amazon Linux AMI
* openSUSE
* SUSE Linux Enterprise Server


### How to

```go
package main

import (
     "fmt"
     "github.com/tactycal/agent/packagelookup"
)

func main() {
    packages, err := packagelookup.Get("ubuntu")
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("%+v\n", packages)
    }
}
```
