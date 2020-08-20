package cluster

import (
	"github.com/anton-dessiatov/sctf/tf/cmd/cluster/show"
	"github.com/spf13/cobra"
)

var clusterID int

var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Commands related to cluster management",
}

func init() {
	ClusterCmd.AddCommand(CreateCmd)
	ClusterCmd.AddCommand(ApplyCmd)
	ClusterCmd.AddCommand(RefreshCmd)
	ClusterCmd.AddCommand(ImportCmd)
	ClusterCmd.AddCommand(DestroyCmd)
	ClusterCmd.AddCommand(show.ShowCmd)
}
