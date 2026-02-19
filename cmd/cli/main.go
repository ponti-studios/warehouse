package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"gogogo/cmd/cli/commands/browser"
	"gogogo/cmd/cli/commands/finance"
	"gogogo/cmd/cli/commands/flatten"
	"gogogo/cmd/cli/commands/importcmd/music"
	"gogogo/cmd/cli/commands/notes"
	"gogogo/cmd/cli/commands/random"
	"gogogo/cmd/cli/commands/server"
	"gogogo/cmd/cli/commands/typingmind"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gogogo",
		Short: "CLI utilities and tools",
	}

	rootCmd.AddCommand(browserCmd())
	rootCmd.AddCommand(flattenCmd())
	rootCmd.AddCommand(typingmindCmd())
	rootCmd.AddCommand(finance.Command())
	rootCmd.AddCommand(notesCmd())
	rootCmd.AddCommand(randomCmd())
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(importCmd())

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func browserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "browser",
		Short: "Browser automation tools",
		Run: func(cmd *cobra.Command, args []string) {
			if err := browser.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func flattenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatten",
		Short: "Flatten directory structure",
		Run: func(cmd *cobra.Command, args []string) {
			if err := flatten.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func typingmindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "typingmind",
		Short: "Convert TypingMind chat data",
		Run: func(cmd *cobra.Command, args []string) {
			if err := typingmind.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func notesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notes",
		Short: "Parse markdown files for notes",
		Run: func(cmd *cobra.Command, args []string) {
			if err := notes.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func randomCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "random",
		Short: "Go example code snippets",
		Run: func(cmd *cobra.Command, args []string) {
			if err := random.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func serverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start REST API server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := server.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func importCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import data from various sources",
	}

	cmd.AddCommand(importAmazonCmd())
	cmd.AddCommand(importAppleCmd())
	cmd.AddCommand(importHealthCmd())
	cmd.AddCommand(importMusicCmd())
	cmd.AddCommand(importTrackingCmd())
	cmd.AddCommand(importSocialCmd())

	return cmd
}

func importAmazonCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "amazon",
		Short: "Import Amazon orders",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: gogogo import amazon <args>")
		},
	}
}

func importAppleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apple",
		Short: "Import Apple receipts",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: gogogo import apple <args>")
		},
	}
}

func importHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Import health data",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: gogogo import health <args>")
		},
	}
}

func importMusicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "music",
		Short: "Import music data (Spotify, Apple Music)",
	}

	cmd.AddCommand(importSpotifyCmd())
	cmd.AddCommand(importAppleMusicCmd())

	return cmd
}

func importSpotifyCmd() *cobra.Command {
	var db, source string
	var dryRun, force bool

	cmd := &cobra.Command{
		Use:   "spotify",
		Short: "Import Spotify data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return music.HandleSpotifyImport(
				context.Background(),
				db,
				source,
				dryRun,
				force,
			)
		},
	}

	cmd.Flags().StringVar(&db, "db", "", "Path to SQLite database")
	cmd.Flags().StringVar(&source, "source", "", "Source directory containing Spotify export")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate without importing")
	cmd.Flags().BoolVar(&force, "force", false, "Skip duplicate checking")

	return cmd
}

func importAppleMusicCmd() *cobra.Command {
	var db, source string
	var dryRun, force bool

	cmd := &cobra.Command{
		Use:   "apple",
		Short: "Import Apple Music data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return music.HandleAppleMusicImport(
				context.Background(),
				db,
				source,
				dryRun,
				force,
			)
		},
	}

	cmd.Flags().StringVar(&db, "db", "", "Path to SQLite database")
	cmd.Flags().StringVar(&source, "source", "", "Source directory containing Apple Music export")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate without importing")
	cmd.Flags().BoolVar(&force, "force", false, "Skip duplicate checking")

	return cmd
}

func importTrackingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tracking",
		Short: "Import tracking data",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: gogogo import tracking <args>")
		},
	}
}

func importSocialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "social",
		Short: "Import social media data",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: gogogo import social <args>")
		},
	}
}
