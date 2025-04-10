package models

type IPAM struct {
	Description string
	Subnets     map[string]Subnets
}

type Subnets struct {
	Description string
	Tags        []string
	Subnets     map[string]Subnets
}
