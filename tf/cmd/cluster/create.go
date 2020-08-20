package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/tf/app"
	"github.com/anton-dessiatov/sctf/tf/cluster"
	"github.com/anton-dessiatov/sctf/tf/dal"
	"github.com/anton-dessiatov/sctf/tf/model"
	"github.com/spf13/cobra"
)

var cloudProviderString string

var CreateCmd = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		cloudProvider := model.CloudProvider(cloudProviderString)
		if err := cloudProvider.Validate(); err != nil {
			log.Fatal(fmt.Errorf("cloudProvider.Validate: %w", err))
		}

		tmpl := cluster.DefaultTemplate(cloudProvider)
		row := dal.Cluster{Template: tmpl}

		if err := app.Instance.DB.Create(&row).Error; err != nil {
			log.Fatal(fmt.Errorf("app.Instance.DB.Create: %w", err))
		}

		log.Printf("Created cluster ID %d", row.ID)
	},
}

func init() {
	CreateCmd.Flags().StringVarP(&cloudProviderString, "cloud-provider", "p", "", "cloud provider")
	CreateCmd.MarkFlagRequired("cloud-provider")
}
