package image

import (
	"errors"
	"fmt"
	"strings"
)

var (
	defaultTag            = "latest"
	defaultHost           = "registry-1.docker.io"
	defaultDockerHost     = "docker.io"
	defaultSchema         = "docker"
	errInvalidImageFormat = errors.New("invalid image format")
	manifestUrlTpl        = "https://%s/v2/%s/manifests/%s"
	digestUrlTpl          = "https://%s/v2/%s/blobs/%s"
)

type ImageUrl struct {
	Schema string
	Host   string
	Name   string
	Tag    string
	Digest string
}

// NewImageUrl initialize an ImageUrl struct
// still need some rules to guard the invalid image format
func NewImageUrl(url string) (*ImageUrl, error) {
	url = strings.ToLower(url)
	i := &ImageUrl{}

	var hostPath string
	secs := strings.SplitN(url, "://", 2)
	if len(secs) == 1 {
		i.Schema = defaultSchema
		hostPath = secs[0]
	} else {
		i.Schema = secs[0]
		hostPath = secs[1]
	}

	if i.Schema != defaultSchema {
		return nil, fmt.Errorf("%w, currently only support schema docker://", errInvalidImageFormat)
	}

	// hostPath = nginx, siji/nginx, docker.io/nginx, docker.io/siji/nginx
	if !strings.Contains(hostPath, "/") {
		hostPath = "library/" + hostPath
	}

	frags := strings.Split(hostPath, "/")
	if strings.Contains(frags[0], ".") {
		if frags[0] == defaultDockerHost {
			i.Host = defaultHost
		} else {
			i.Host = frags[0]
		}
		frags = frags[1:]
	} else {
		i.Host = defaultHost
	}

	fragLen := len(frags)
	lastFrag := frags[fragLen-1]

	if strings.Contains(lastFrag, "@") {
		index := strings.Index(lastFrag, "@")
		i.Digest = lastFrag[index+1:]
		frags[fragLen-1] = lastFrag[:index]
	} else if strings.Contains(lastFrag, ":") {
		index := strings.Index(lastFrag, ":")
		i.Tag = lastFrag[index+1:]
		frags[fragLen-1] = lastFrag[:index]
	} else {
		i.Tag = defaultTag
	}

	// insert library only for Docker Hub images
	if fragLen == 1 && i.Host == defaultHost {
		i.Name = "library/" + frags[0]
	} else {
		i.Name = strings.Join(frags, "/")
	}

	return i, nil
}

func (i *ImageUrl) String() string {
	var fullName string
	if i.Tag != "" {
		fullName = i.Name + ":" + i.Tag
	}

	if i.Digest != "" {
		fullName = i.Name + "@" + i.Digest
	}

	if fullName == "" {
		fullName = i.Name + ":" + defaultTag
	}

	return i.Schema + "://" + i.Host + "/" + fullName
}

func (i *ImageUrl) RepoString() string {
	var fullName string
	if i.Tag != "" {
		fullName = i.Name + ":" + i.Tag
	}

	if i.Digest != "" {
		fullName = i.Name + "@" + i.Digest
	}

	if fullName == "" {
		fullName = i.Name + ":" + defaultTag
	}

	host := i.Host
	if host == defaultHost {
		host = defaultDockerHost
	}

	return host + "/" + fullName
}

func (i *ImageUrl) ManifestURL() string {
	ref := defaultTag
	if i.Tag != "" {
		ref = i.Tag
	}

	if i.Digest != "" {
		ref = i.Digest
	}

	return fmt.Sprintf(manifestUrlTpl, i.Host, i.Name, ref)
}

func (i *ImageUrl) DigestUrl(digest string) string {
	return fmt.Sprintf(digestUrlTpl, i.Host, i.Name, digest)
}
