package cmd

import (
	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/spf13/cobra"
)

var listFormatsCmd = &cobra.Command{
	Use:     "list-formats",
	Aliases: []string{"lf"},
	Short:   "Show supported input/output format combinations",
	RunE: func(cmd *cobra.Command, args []string) error {
		format.Primary("Supported conversions by engine:")
		format.Info("  pandoc:     pdf, docx, odt, md, tex, epub, html, rst")
		format.Info("  imagemagick: jpg, png, gif, tiff, bmp, webp, svg")
		format.Info("  ffmpeg:     mp4, mkv, avi, mov, mp3, wav, webm, m4a, flac, ogg")
		format.Info("  data:       csv, tsv, json, yaml, toml")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listFormatsCmd)
}
