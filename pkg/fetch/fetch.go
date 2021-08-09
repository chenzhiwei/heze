package fetch

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	stdUrl "net/url"
	"os"
	"strings"

	glog "github.com/goduang/glog"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/chenzhiwei/heze/pkg/image"
)

type ImageFetcher struct {
	Username  string
	Password  string
	isAuthed  bool
	authToken map[string]string
}

func (i *ImageFetcher) Fetch(ctx context.Context, img *image.ImageUrl, output string) error {
	manifestBytes, err := i.FetchManifest(ctx, img)
	if err != nil {
		return err
	}
	glog.V(2).Infof("Image Manifest: %s\n", manifestBytes)

	manifest := &imgspecv1.Manifest{}
	if err := json.Unmarshal(manifestBytes, manifest); err != nil {
		return err
	}

	if manifest.Config.Digest == "" {
		return errors.New("Do not support fat image")
	}

	savedBytes, err := savedManifest(img, manifest)
	if err != nil {
		return err
	}

	tarfile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	var fileWriter io.WriteCloser = tarfile
	tarfileWriter := tar.NewWriter(fileWriter)
	defer tarfileWriter.Close()

	// write manifest.json
	manifestHeader := &tar.Header{
		Name: "manifest.json",
		Size: int64(len(savedBytes)),
		Mode: 0644,
	}
	err = tarfileWriter.WriteHeader(manifestHeader)
	if err != nil {
		return err
	}
	_, err = io.Copy(tarfileWriter, bytes.NewReader(savedBytes))
	if err != nil {
		return err
	}

	// write config+layers to save image tar file
	layers := append(manifest.Layers, manifest.Config)
	for _, layer := range layers {
		glog.V(1).Infof("Layer digest: %s, size: %d\n", layer.Digest, layer.Size)
		layerUrl := img.DigestUrl(layer.Digest.String())
		layerHeader := &tar.Header{
			Name: layer.Digest.String(),
			Size: layer.Size,
			Mode: 0644,
		}
		err = tarfileWriter.WriteHeader(layerHeader)
		if err != nil {
			return err
		}

		res, err := i.makeRequest(ctx, layerUrl)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			_, err = io.Copy(tarfileWriter, res.Body)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error response")
		}
	}

	return nil
}

func (i *ImageFetcher) FetchManifest(ctx context.Context, img *image.ImageUrl) ([]byte, error) {
	manifestUrl := img.ManifestURL()
	glog.V(1).Infof("Image Manifest URL: %s\n", manifestUrl)

	return i.fetchContent(ctx, manifestUrl)
}

func (i *ImageFetcher) FetchConfig(ctx context.Context, img *image.ImageUrl, digest string) ([]byte, error) {
	configUrl := img.DigestUrl(digest)
	glog.V(1).Infof("Image Config URL: %s\n", configUrl)

	return i.fetchContent(ctx, configUrl)
}

func (i *ImageFetcher) fetchFile(ctx context.Context, url, filename string) error {
	res, err := i.makeRequest(ctx, url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		out, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, res.Body)
		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.New("Error response")
	}
}

func (i *ImageFetcher) fetchContent(ctx context.Context, url string) ([]byte, error) {
	res, err := i.makeRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return body, nil
	} else {
		return nil, errors.New("Error response")
	}
}

func (i *ImageFetcher) makeRequest(ctx context.Context, url string) (*http.Response, error) {
	header := http.Header{
		"Accept": image.DefaultRequestedManifestMIMETypes,
	}

	host := strings.Split(url, "/")[2]
	token, ok := i.authToken[host]
	if ok {
		header.Set("Authorization", "Bearer "+token)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header = header
	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized && i.isAuthed == false {
		i.isAuthed = true
		authHead := response.Header.Get("www-authenticate")
		glog.V(2).Infof("Auth Header www-authenticate: %s\n", authHead)
		if err := i.requestAuthTokens(ctx, host, authHead); err != nil {
			return nil, err
		}
		return i.makeRequest(ctx, url)
	} else {
		return response, nil
	}
}

func (i *ImageFetcher) requestAuthTokens(ctx context.Context, host, authHead string) error {
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
		return errors.New("missing realm in bearer auth challenge")
	}
	if service == "" {
		return errors.New("missing service in bearer auth challenge")
	}
	if scope == "" {
		return errors.New("missing scope in bearer auth challenge")
	}

	glog.V(2).Infof("bearer realm: %s, service: %s, scope: %s\n", realm, service, scope)

	params := stdUrl.Values{
		"service": {service},
		"scope":   {scope},
	}
	reqUrl := realm + "?" + params.Encode()

	header := http.Header{}
	if i.Username != "" && i.Password != "" {
		header.Set("Authorization", "Basic "+basicAuth(i.Username, i.Password))
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return err
	}

	request.Header = header
	client := http.DefaultClient
	res, err := client.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		tokenStruct := struct {
			Token string `json:"token"`
		}{}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &tokenStruct)
		if err != nil {
			return err
		}

		if i.authToken == nil {
			i.authToken = make(map[string]string)
		}

		i.authToken[host] = tokenStruct.Token

		return nil
	} else {
		return fmt.Errorf("Failed to get authToken, response code %d", res.StatusCode)
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func hostFromUrl(url string) string {
	return strings.Split(url, "/")[2]
}

func savedManifest(img *image.ImageUrl, manifest *imgspecv1.Manifest) ([]byte, error) {
	var layers []string
	for _, layer := range manifest.Layers {
		layers = append(layers, layer.Digest.String())
	}

	savedManifests := make(image.SavedManifests, 1)
	savedManifests[0] = image.SavedManifest{
		Config:   manifest.Config.Digest.String(),
		RepoTags: []string{img.RepoString()},
		Layers:   layers,
	}

	return json.Marshal(savedManifests)
}
