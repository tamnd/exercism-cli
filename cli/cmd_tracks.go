package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) tracksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tracks",
		Short: "List all programming language tracks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching tracks...")
			tracks, err := a.client.Tracks(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			n := a.effectiveLimit(0)
			if n > 0 && n < len(tracks) {
				tracks = tracks[:n]
			}
			return a.renderOrEmpty(tracks, len(tracks))
		},
	}
}

func (a *App) trackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "track <slug>",
		Short: "Show detail for a single track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			a.progressf("fetching track %q...", slug)
			tr, err := a.client.GetTrack(cmd.Context(), slug)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(tr)
		},
	}
}
