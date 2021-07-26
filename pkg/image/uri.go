package image

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	defaultTag            = "latest"
	defaultHost           = "registry-1.docker.io"
	defaultSchema         = "docker"
	errInvalidImageFormat = fmt.Errorf("invalid image format")
)

type ImageURI struct {
	HttpSchema string
	Schema     string
	Host       string
	Port       int
	Name       string
	Tag        string
	Digest     string
}

func NewImageURI(url string) (*ImageURI, error) {
	i := &ImageURI{}

	var fullPath string
	secs := strings.Split(url, "://")
	if len(secs) == 1 {
		i.Schema = defaultSchema
		fullPath = secs[0]
	} else if len(secs) == 2 {
		i.Schema = strings.ToLower(secs[0])
		fullPath = secs[1]
	} else {
		return nil, errInvalidImageFormat
	}

	fields := strings.Split(fullPath, "/")
	name := fields[len(fields)-1]
	if strings.Contains(name, "@") {
		index := strings.Index(name, "@")
		i.Digest = name[index+1:]
		fields[len(fields)-1] = name[:index]
	} else if strings.Contains(name, ":") {
		index := strings.Index(name, ":")
		i.Tag = name[index+1:]
		fields[len(fields)-1] = name[:index]
	} else {
		i.Tag = defaultTag
	}

	// image format: username/image/haha
	if !strings.Contains(fields[0], ".") && len(fields) > 2 {
		return nil, errInvalidImageFormat
	}

	// image format: quay.io
	if strings.Contains(fields[0], ".") && len(fields) == 1 {
		return nil, errInvalidImageFormat
	}

	var fullName string
	if !strings.Contains(fields[0], ".") {
		i.Host = defaultHost
		if len(fields) == 1 {
			fullName = "library/" + fields[0]
		} else {
			fullName = strings.Join(fields, "/")
		}
	} else {
		fullName = strings.Join(fields[1:], "/")
		ss := strings.Split(fields[0], ":")
		if len(ss) == 1 {
			i.Host = fields[0]
		} else if len(ss) == 2 {
			i.Host = ss[0]
			port, err := strconv.Atoi(ss[1])
			if err != nil {
				return nil, errInvalidImageFormat
			}
			i.Port = port
		} else {
			return nil, errInvalidImageFormat
		}

	}

	i.Name = fullName

	return i, nil
}

func (i *ImageURI) String() string {
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

	fullHost := i.Host
	// Using 443 on HTTP and 80  on HTTPS is ignored
	if i.Port != 0 && i.Port != 443 && i.Port != 80 {
		fullHost = fullHost + ":" + strconv.Itoa(i.Port)
	}

	return i.Schema + "://" + fullHost + "/" + fullName
}

func (i *ImageURI) ManifestURL() string {
	return ""
}
