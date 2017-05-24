package osDiscovery

import "fmt"

func ExampleGetKernel() {
	kernel, err := GetKernel()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(kernel)
	}
}

func ExampleGetDistributionRelease() {
	distribution, release, err := GetDistributionRelease()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(distribution, release)
	}
}

func ExampleGetArchitecture() {
	architecture, err := GetArchitecture()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(architecture)
	}
}

func ExampleGetFqdn() {
	fqdn, err := GetFqdn()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(fqdn)
	}
}

func ExampleGet() {
	osInfo, err := Get()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", osInfo)
	}
}
