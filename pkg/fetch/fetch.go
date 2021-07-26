package fetch

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/chenzhiwei/heze/pkg/image"
	gohttp "github.com/goduang/http"
)

type ImageFetcher struct {
	Username   string
	Password   string
	isAuthed   bool
	authTokens map[string]map[string]string
}

func (i *ImageFetcher) FetchManifest(ctx context.Context, img *image.ImageUrl) ([]byte, error) {
	data := &gohttp.HttpRequest{
		Url:    img.ManifestURL(),
		Client: http.DefaultClient,
	}
	res, err := gohttp.MakeRequest(ctx, data)
	if err != nil {
		return nil, err
	}

	if res.Code == http.StatusOK {
		return res.Body, nil
	} else if res.Code != http.StatusUnauthorized {
		return nil, fmt.Errorf("Failed to fetch manifest, response code %d", res.Code)
	}

	if res.Code == http.StatusUnauthorized && i.isAuthed == false {
		authHead := res.Header.Get("www-authenticate")
		fmt.Println(authHead)
		if err := i.setupAuthTokens(ctx, img, authHead); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (i *ImageFetcher) setupAuthTokens(ctx context.Context, img *image.ImageUrl, authHead string) error {
	tokens := strings.Split(authHead, ",")
	if len(tokens) != 3 || !strings.HasPrefix(strings.ToLower(tokens[0]), "bearer realm") {
		return fmt.Errorf("could not parse www-authenticate header: %s", authHead)
	}
	return nil
}
