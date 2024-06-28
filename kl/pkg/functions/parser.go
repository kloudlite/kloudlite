package functions

import "github.com/spf13/cobra"

func ParseStringFlag(cmd *cobra.Command, flag string) string {
	v, _ := cmd.Flags().GetString(flag)
	return v
}

func ParseIntFlag(cmd *cobra.Command, flag string) int {
	v, _ := cmd.Flags().GetInt(flag)
	return v
}

func ParseBoolFlag(cmd *cobra.Command, flag string) bool {
	v, _ := cmd.Flags().GetBool(flag)
	return v
}

func WithOutputVariant(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "output format [table | json | yaml]")
}

func WithKlFile(cmd *cobra.Command) {
	cmd.Flags().StringP("klfile", "k", "", "kloudlite file")
}

func ParseKlFile(cmd *cobra.Command) string {
	if cmd.Flags().Changed("klfile") {
		v, _ := cmd.Flags().GetString("klfile")
		return v
	}

	return ""
}
