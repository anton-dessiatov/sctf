package cluster

import (
	"fmt"

	"github.com/anton-dessiatov/sctf/pulumi/model"
)

func DefaultTemplate(cp model.CloudProvider) model.ClusterTemplate {
	result := model.ClusterTemplate{
		CloudProvider: cp,
		PrivateCIDR:   "10.0.0.0/16",
	}

	if cp == model.CloudProviderAWS {
		result.AWS.Region = awsRegion
	}
	if cp == model.CloudProviderGCP {
		result.GCP.Region = gcpRegion
	}

	for i := 0; i < nodes; i++ {
		srv := model.ServerTemplate{
			ResourceID: fmt.Sprintf("node%d", i),
			SubnetCIDR: fmt.Sprintf("10.0.%d.0/24", i),
		}
		if cp == model.CloudProviderAWS {
			srv.AWS.AvailabilityZone = awsAvailabilityZones[i%len(awsAvailabilityZones)]
		}
		if cp == model.CloudProviderGCP {
			srv.GCP.AvailabilityZone = gcpAvailabilityZones[i%len(gcpAvailabilityZones)]
		}
		result.Servers = append(result.Servers, srv)
	}

	return result
}

const (
	nodes     = 2
	awsRegion = "us-east-1"
	gcpRegion = "us-east4"
)

var awsAvailabilityZones = []string{
	"us-east-1a",
	"us-east-1b",
	"us-east-1c",
}

var gcpAvailabilityZones = []string{
	"us-east4-a",
	"us-east4-b",
	"us-east4-c",
}
