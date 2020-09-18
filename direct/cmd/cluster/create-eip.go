package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/cluster"
	"github.com/spf13/cobra"
)

var CreateEIPCmd = &cobra.Command{
	Use: "create-eip",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cluster.CmdCreateEIP(app.Instance, clusterID); err != nil {
			log.Fatal(fmt.Errorf("cluster.CmdCreateEIP: %w", err))
		}
	},
}

func init() {
	CreateEIPCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	CreateEIPCmd.MarkFlagRequired("cluster-id")
}
