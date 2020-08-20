package show

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"

	"github.com/spf13/cobra"
)

var StateCmd = &cobra.Command{
	Use: "state",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := cluster.StateByClusterID(app.Instance.DB, app.Instance.Terra, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("cluster.StateByClusterID: %w", err))
		}

		var p []byte
		p, err = json.MarshalIndent(st, "", "\t")
		if err != nil {
			log.Fatal(fmt.Errorf("json.MarshalIndent: %w", err))
			return
		}
		log.Printf("%s \n", p)
	},
}
