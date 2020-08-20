package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"
	"github.com/spf13/cobra"
)

var DestroyCmd = &cobra.Command{
	Use: "destroy",
	Run: func(cmd *cobra.Command, args []string) {
		stack, err := cluster.StackByClusterID(app.Instance.DB, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.StackByClusterID: %w", err))
		}

		diags := app.Instance.Terra.Apply(cluster.StackIdentity(clusterID), stack, true)
		for _, d := range diags {
			log.Println(d)
		}

		if !diags.HasErrors() {
			log.Println("Destroyed successfully!")
		}
	},
}

func init() {
	DestroyCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	DestroyCmd.MarkFlagRequired("cluster-id")
}
