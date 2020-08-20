package cluster

import (
	"fmt"

	"github.com/anton-dessiatov/sctf/direct/model"
)

func DefaultTemplate() model.ClusterTemplate {
	result := model.ClusterTemplate{
		PrivateCIDR: "10.0.0.0/16",
	}

	for i := 0; i < nodes; i++ {
		srv := model.ServerTemplate{
			SubnetCIDR:       fmt.Sprintf("10.0.%d.0/24", i),
			AvailabilityZone: awsAvailabilityZones[i%len(awsAvailabilityZones)],
		}
		result.Servers = append(result.Servers, srv)
	}

	return result
}

const nodes = 2

var awsAvailabilityZones = []string{
	"us-east-1a",
	"us-east-1b",
	"us-east-1c",
}
