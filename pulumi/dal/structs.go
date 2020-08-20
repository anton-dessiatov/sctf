package dal

import (
	"time"

	"github.com/anton-dessiatov/sctf/pulumi/model"
)

type Cluster struct {
	ID        int
	Template  model.ClusterTemplate
	CreatedAt time.Time
	UpdatedAt time.Time
}
