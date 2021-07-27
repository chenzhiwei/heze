package cmd

import (
	"context"

	"github.com/goduang/glog"
	"github.com/spf13/cobra"

	"github.com/chenzhiwei/heze/pkg/fetch"
	"github.com/chenzhiwei/heze/pkg/image"
)

var (
	username  string
	password  string
	outputdir string

	fetchCmd = &cobra.Command{
		Use:   "fetch [image]",
		Short: "Fetch the OCI/Docker image from remote image registry",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := runFetch(args); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	fetchCmd.Flags().StringVarP(&username, "username", "u", "", "username of remote image registry")
	fetchCmd.Flags().StringVarP(&password, "password", "p", "", "password of remote image registry")
	fetchCmd.Flags().StringVar(&outputdir, "outputdir", ".", "outputdir of the fetched image")
	fetchCmd.Flags().SortFlags = false
}

func runFetch(args []string) error {
	url := args[0]
	img, err := image.NewImageUrl(url)
	if err != nil {
		return err
	}

	glog.V(1).Infof("Image URL: %s\n", img)

	ctx := context.TODO()

	fc := &fetch.ImageFetcher{
		Username: username,
		Password: password,
	}

	return fc.Fetch(ctx, img)
}
