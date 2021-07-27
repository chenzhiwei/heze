package fetch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	glog "github.com/goduang/glog"
	gohttp "github.com/goduang/http"

	"github.com/chenzhiwei/heze/pkg/image"
)

type ImageFetcher struct {
	Username   string
	Password   string
	isAuthed   bool
	authTokens map[string]map[string]string
}

func (i *ImageFetcher) FetchManifest(ctx context.Context, img *image.ImageUrl) ([]byte, error) {
	glog.V(1).Infof("Image Manifest URL: %s\n", img.ManifestURL())

	header := http.Header{
		"Accept": image.DefaultRequestedManifestMIMETypes,
	}

	authToken, ok := i.authTokens[img.Host]
	if ok {
		token, ok := authToken[img.Name]
		if ok {
			header.Set("Authorization", "Bearer "+token)
		}
	}

	data := &gohttp.HttpRequest{
		Url:    img.ManifestURL(),
		Client: http.DefaultClient,
		Header: header,
	}

	res, err := gohttp.MakeRequest(ctx, data)
	if err != nil {
		return nil, err
	}

	if res.Code == http.StatusOK {
		return res.Body, nil
	} else if res.Code == http.StatusUnauthorized && i.isAuthed == false {
		i.isAuthed = true
		authHead := res.Header.Get("www-authenticate")
		glog.V(1).Infof("Auth Header www-authenticate: %s\n", authHead)
		if err := i.setupAuthTokens(ctx, img, authHead); err != nil {
			return nil, err
		}
		return i.FetchManifest(ctx, img)
	} else {
		glog.V(1).Infof("Res Header: %v, Body: %s\n", res.Header, res.Body)
		return nil, fmt.Errorf("Failed to fetch manifest, response code %d", res.Code)
	}
}

func (i *ImageFetcher) setupAuthTokens(ctx context.Context, img *image.ImageUrl, authHead string) error {
	authHead = strings.ToLower(authHead)
	tokens := strings.Split(authHead, ",")
	if len(tokens) != 3 || !strings.HasPrefix(strings.ToLower(tokens[0]), "bearer realm") {
		return fmt.Errorf("could not parse www-authenticate header: %s", authHead)
	}

	var realm, service, scope string
	for _, token := range tokens {
		if strings.HasPrefix(token, "bearer realm") {
			realm = strings.Trim(token[len("bearer realm="):], "\"")
		}
		if strings.HasPrefix(token, "service") {
			service = strings.Trim(token[len("service="):], "\"")
		}
		if strings.HasPrefix(token, "scope") {
			scope = strings.Trim(token[len("scope="):], "\"")
		}
	}

	if realm == "" {
		return fmt.Errorf("missing realm in bearer auth challenge")
	}
	if service == "" {
		return fmt.Errorf("missing service in bearer auth challenge")
	}
	if scope == "" {
		return fmt.Errorf("missing scope in bearer auth challenge")
	}

	glog.V(2).Infof("bearer realm: %s, service: %s, scope: %s\n", realm, service, scope)

	params := url.Values{
		"service": {service},
		"scope":   {scope},
	}
	reqUrl := realm + "?" + params.Encode()

	header := http.Header{}
	if i.Username != "" && i.Password != "" {
		header.Set("Authorization", "Basic "+basicAuth(i.Username, i.Password))
	}

	data := &gohttp.HttpRequest{
		Url:    reqUrl,
		Client: http.DefaultClient,
		Header: header,
	}

	res, err := gohttp.MakeRequest(ctx, data)
	if err != nil {
		return err
	}

	if res.Code == http.StatusOK {
		tokenStruct := struct {
			Token string `json:"token"`
		}{}

		err = json.Unmarshal(res.Body, &tokenStruct)
		if err != nil {
			return err
		}

		if i.authTokens == nil {
			i.authTokens = make(map[string]map[string]string)
		}

		repo := strings.Split(scope, ":")[1]
		authToken := map[string]string{
			repo: tokenStruct.Token,
		}

		i.authTokens[img.Host] = authToken
	} else {
		return fmt.Errorf("Failed to get authToken, response code %d", res.Code)
	}

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
