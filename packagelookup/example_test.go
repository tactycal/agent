package packagelookup

import "fmt"

func ExampleGet() {
	packages, err := Get(UBUNTU)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", packages)
	}
}
