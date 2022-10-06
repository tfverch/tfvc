package source

import svchost "github.com/hashicorp/terraform-svchost"

type Registry struct {
	Hostname   string
	Namespace  string
	Name       string
	Provider   string
	Normalized string
}

type RegistryProvider struct {
	Type       string
	Namespace  string
	Hostname   svchost.Hostname
	Normalized string
}
