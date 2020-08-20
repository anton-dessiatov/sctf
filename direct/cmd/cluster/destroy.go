package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/cluster"
	"github.com/anton-dessiatov/sctf/direct/dal"
	"github.com/spf13/cobra"
)

var DestroyCmd = &cobra.Command{
	Use: "destroy",
	Run: func(cmd *cobra.Command, args []string) {
		var c dal.Cluster
		if err := app.Instance.DB.Where("id = ?", clusterID).First(&c).Error; err != nil {
			log.Fatal(fmt.Errorf("app.Instance.DB.First: %w", err))
		}

		if err := cluster.Destroy(app.Instance, c); err != nil {
			log.Fatal(fmt.Errorf("cluster.Destroy: %w", err))
		}
	},
}

func init() {
	DestroyCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	DestroyCmd.MarkFlagRequired("cluster-id")
}
