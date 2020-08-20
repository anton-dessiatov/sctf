package cluster

import (
	"context"

	"github.com/anton-dessiatov/sctf/pulumi/app"
	"github.com/pulumi/pulumi/sdk/go/common/workspace"
)

// ClusterWorkspace is a Pulumi workspace for a cluster
type ClusterWorkspace struct {
	App       *app.App
	ClusterID int

	project *workspace.Project
}

func (w *ClusterWorkspace) ProjectSettings(context.Context) (*workspace.Project, error) {
	return nil, nil
}

func SaveProjectSettings(context.Context, *workspace.Project) error {
	return nil
}
