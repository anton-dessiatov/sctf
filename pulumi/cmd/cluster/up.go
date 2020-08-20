package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/pulumi/app"
	"github.com/anton-dessiatov/sctf/pulumi/cluster"
	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {
		err := cluster.Up(app.Instance, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.Up: %w", err))
		}
	},
}

func init() {
	UpCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	UpCmd.MarkFlagRequired("cluster-id")
}
