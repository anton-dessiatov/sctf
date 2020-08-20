package show

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"
	"github.com/anton-dessiatov/sctf/tf/terra"
	"github.com/spf13/cobra"
)

var TerraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Prints terraform config used for given cluster",
	Run: func(cmd *cobra.Command, args []string) {
		stack, err := cluster.StackByClusterID(app.Instance.DB, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.StackByClusterID: %w", err))
		}

		st, ok := stack.Config.(terra.StackText)
		if !ok {
			log.Fatal("stack is not text-based")
		}

		log.Println(st)
	},
}
