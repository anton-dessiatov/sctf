package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"
	"github.com/spf13/cobra"
)

var RefreshCmd = &cobra.Command{
	Use: "refresh",
	Run: func(cmd *cobra.Command, args []string) {
		stack, err := cluster.StackByClusterID(app.Instance.DB, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.StackByClusterID: %w", err))
		}

		diags := app.Instance.Terra.Refresh(cluster.StackIdentity(clusterID), stack)
		for _, d := range diags {
			log.Println(d.Description())
		}

		if !diags.HasErrors() {
			log.Println("Refreshed successfully!")
		}
	},
}

func init() {
	RefreshCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	RefreshCmd.MarkFlagRequired("cluster-id")
}
