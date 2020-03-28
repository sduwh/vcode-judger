package main

type Options struct {
	Redis struct {
		Address string `json:"address" yaml:"address"`
	} `json:"redis" yaml:"redis"`
}
