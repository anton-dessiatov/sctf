package dal

import "time"

type Cluster struct {
	ID        int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type VPC struct {
	ID         int
	ClusterID  int
	ExternalID string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type IGW struct {
	ID         int
	ClusterID  int
	ExternalID string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Subnet struct {
	ID         int
	ClusterID  int
	ExternalID string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
type RTB struct {
	ID            int
	ClusterID     int
	SubnetID      *int
	ExternalID    string
	AssociationID *string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SecurityGroup struct {
	ID         int
	ClusterID  int
	Kind       string
	ExternalID string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Instance struct {
	ID         int
	ClusterID  int
	ExternalID string
	PublicIP   *string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
