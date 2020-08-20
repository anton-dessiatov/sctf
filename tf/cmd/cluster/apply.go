package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"
	"github.com/spf13/cobra"
)

var ApplyCmd = &cobra.Command{
	Use: "apply",
	Run: func(cmd *cobra.Command, args []string) {
		stack, err := cluster.StackByClusterID(app.Instance.DB, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.StackByClusterID: %w", err))
		}

		diags := app.Instance.Terra.Apply(cluster.StackIdentity(clusterID), stack, false)
		for _, d := range diags {
			log.Println(d.Description())
		}

		if !diags.HasErrors() {
			log.Println("Applied successfully!")
		}
	},
}

func init() {
	ApplyCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	ApplyCmd.MarkFlagRequired("cluster-id")
}
