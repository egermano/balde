package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

// writeJSON emits v as indented JSON on the command's stdout — the single
// JSON output path for every command.
func writeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
