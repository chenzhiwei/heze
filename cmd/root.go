package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "heze",
	Short: "Heze is a tool to download OCI/Docker images",
	Long:  `A very simple tool to download OCI/Docker images which can be import by podman and docker`,
}

func Execute() error {
	return rootCmd.Execute()
}
