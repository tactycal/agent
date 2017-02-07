package main

type Package struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Source       string `json:"source"`
	Architecture string `json:"architecture"`
	Maintainer   string `json:"-"`
	Official     bool   `json:"official"`
}
