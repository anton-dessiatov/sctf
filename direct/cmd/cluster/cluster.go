package cluster

import (
	"github.com/spf13/cobra"
)

var clusterID int

var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Commands related to cluster management",
}

func init() {
	ClusterCmd.AddCommand(CreateCmd)
	ClusterCmd.AddCommand(CreateEIPCmd)
	ClusterCmd.AddCommand(DestroyCmd)
}
