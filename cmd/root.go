package cmd

import (
	"github.com/goduang/glog"
	"github.com/spf13/cobra"
)

var (
	verbosity int
	rootCmd   = &cobra.Command{
		Use:          "heze",
		Short:        "Heze is a tool to download OCI/Docker images",
		Long:         `A very simple tool to download OCI/Docker images which can be import by podman and docker`,
		SilenceUsage: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			glog.InitLogs(verbosity)
			go glog.Flush()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 0, "the log verbosity")

	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
