package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/cluster"
	"github.com/spf13/cobra"
)

var cloudProviderString string

var CreateCmd = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cluster.Create(app.Instance, cluster.DefaultTemplate()); err != nil {
			log.Fatal(fmt.Errorf("cluster.Create: %w", err))
		}
	},
}
