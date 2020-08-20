package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type CloudProvider string

const (
	CloudProviderAWS CloudProvider = "aws"
	CloudProviderGCP CloudProvider = "gcp"
)

func (cp CloudProvider) Validate() error {
	if cp == CloudProviderAWS {
		return nil
	}
	if cp == CloudProviderGCP {
		return nil
	}

	return fmt.Errorf("unexpected cloud provider: %q", string(cp))
}

type ClusterIdentity int

type ClusterTemplate struct {
	AWS           ClusterTemplateAWS
	GCP           ClusterTemplateGCP
	CloudProvider CloudProvider
	PrivateCIDR   string
	Servers       []ServerTemplate
}

type ClusterTemplateAWS struct {
	Region string
}

type ClusterTemplateGCP struct {
	Region string
}

func (c *ClusterTemplate) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %s", value)
	}

	var result ClusterTemplate
	err := json.Unmarshal(bytes, &result)
	*c = result
	return err
}

func (c ClusterTemplate) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type ClusterState struct {
	Template ClusterTemplate
	Servers  []ServerState
}

type ServerTemplate struct {
	AWS        ServerTemplateAWS
	GCP        ServerTemplateGCP
	SubnetCIDR string
	// ResourceID is similar to ExternalID, but instead of being AWS ID, it
	// is Terraform resource name
	ResourceID string
}

type ServerTemplateAWS struct {
	AvailabilityZone string
}

type ServerTemplateGCP struct {
	AvailabilityZone string
}

type ServerState struct {
	Template ServerTemplate
	AWS      ServerStateAWS
	GCP      ServerStateGCP
	PublicIP string
}

type ServerStateAWS struct {
	ID string
}

type ServerStateGCP struct {
	SelfLink   string
	InstanceID string
}
