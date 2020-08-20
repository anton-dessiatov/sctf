package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"
	"github.com/spf13/cobra"
)

var importAddress string
var importID string

var ImportCmd = &cobra.Command{
	Use: "import",
	Run: func(cmd *cobra.Command, args []string) {
		stack, err := cluster.StackByClusterID(app.Instance.DB, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.StackByClusterID: %w", err))
		}

		diags := app.Instance.Terra.Import(cluster.StackIdentity(clusterID), stack,
			importAddress, importID)
		for _, d := range diags {
			log.Println(d.Description())
		}

		if !diags.HasErrors() {
			log.Println("Imported the resource successfully!")
		}
	},
}

func init() {
	ImportCmd.Flags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	ImportCmd.Flags().StringVar(&importAddress, "address", "", "imported resource address at config")
	ImportCmd.Flags().StringVar(&importID, "id", "", "imported resource provider-specific id")
	ImportCmd.MarkFlagRequired("cluster-id")
	ImportCmd.MarkFlagRequired("address")
	ImportCmd.MarkFlagRequired("id")
}
