package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) conceptsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "concepts <slug>",
		Short: "List concepts in a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			a.progressf("fetching concepts for track %q...", slug)
			concepts, err := a.client.Concepts(cmd.Context(), slug)
			if err != nil {
				return mapFetchErr(err)
			}
			n := a.effectiveLimit(0)
			if n > 0 && n < len(concepts) {
				concepts = concepts[:n]
			}
			return a.renderOrEmpty(concepts, len(concepts))
		},
	}
}
