package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
    Use:   "convert",
    Short: "Convert files between formats",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Convert command not yet implemented")
    },
}

func init() {
    rootCmd.AddCommand(convertCmd)
}
