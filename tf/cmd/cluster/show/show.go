package show

import "github.com/spf13/cobra"

var clusterID int

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show various cluster information",
}

func init() {
	ShowCmd.PersistentFlags().IntVar(&clusterID, "cluster-id", 0, "cluster ID")
	ShowCmd.MarkFlagRequired("cluster-id")

	ShowCmd.AddCommand(TemplateCmd)
	ShowCmd.AddCommand(TerraformCmd)
	ShowCmd.AddCommand(StateCmd)
}
