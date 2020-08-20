package model

type ClusterTemplate struct {
	PrivateCIDR string
	Servers     []ServerTemplate
}

type ServerTemplate struct {
	AvailabilityZone string
	SubnetCIDR       string
}
