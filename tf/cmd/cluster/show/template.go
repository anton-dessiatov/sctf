package show

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/dal"

	"github.com/spf13/cobra"
)

var TemplateCmd = &cobra.Command{
	Use: "template",
	Run: func(cmd *cobra.Command, args []string) {
		cluster, err := dal.ClusterByID(app.Instance.DB, clusterID)
		if err != nil {
			log.Fatal(fmt.Errorf("dal.ClusterByID: %w", err))
		}

		var p []byte
		p, err = json.MarshalIndent(cluster.Template, "", "\t")
		if err != nil {
			log.Fatal(fmt.Errorf("json.MarshalIndent: %w", err))
			return
		}
		log.Printf("%s \n", p)
	},
}
