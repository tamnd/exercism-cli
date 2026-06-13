package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) exercisesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exercises <slug>",
		Short: "List exercises in a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			a.progressf("fetching exercises for track %q...", slug)
			exercises, err := a.client.Exercises(cmd.Context(), slug)
			if err != nil {
				return mapFetchErr(err)
			}
			n := a.effectiveLimit(0)
			if n > 0 && n < len(exercises) {
				exercises = exercises[:n]
			}
			return a.renderOrEmpty(exercises, len(exercises))
		},
	}
}
