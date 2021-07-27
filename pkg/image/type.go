package image

import imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"

const (
	// DockerV2Schema1MediaType MIME type represents Docker manifest schema 1
	DockerV2Schema1MediaType = "application/vnd.docker.distribution.manifest.v1+json"
	// DockerV2Schema1MediaType MIME type represents Docker manifest schema 1 with a JWS signature
	DockerV2Schema1SignedMediaType = "application/vnd.docker.distribution.manifest.v1+prettyjws"
	// DockerV2Schema2MediaType MIME type represents Docker manifest schema 2
	DockerV2Schema2MediaType = "application/vnd.docker.distribution.manifest.v2+json"
	// DockerV2Schema2ConfigMediaType is the MIME type used for schema 2 config blobs.
	DockerV2Schema2ConfigMediaType = "application/vnd.docker.container.image.v1+json"
	// DockerV2Schema2LayerMediaType is the MIME type used for schema 2 layers.
	DockerV2Schema2LayerMediaType = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	// DockerV2SchemaLayerMediaTypeUncompressed is the mediaType used for uncompressed layers.
	DockerV2SchemaLayerMediaTypeUncompressed = "application/vnd.docker.image.rootfs.diff.tar"
	// DockerV2ListMediaType MIME type represents Docker manifest schema 2 list
	DockerV2ListMediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
	// DockerV2Schema2ForeignLayerMediaType is the MIME type used for schema 2 foreign layers.
	DockerV2Schema2ForeignLayerMediaType = "application/vnd.docker.image.rootfs.foreign.diff.tar"
	// DockerV2Schema2ForeignLayerMediaType is the MIME type used for gzipped schema 2 foreign layers.
	DockerV2Schema2ForeignLayerMediaTypeGzip = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
)

// DefaultRequestedManifestMIMETypes is a list of MIME types a types.ImageSource
// should request from the backend unless directed otherwise.
var DefaultRequestedManifestMIMETypes = []string{
	imgspecv1.MediaTypeImageManifest,
	imgspecv1.MediaTypeImageIndex,
	DockerV2Schema2MediaType,
	DockerV2Schema1SignedMediaType,
	DockerV2Schema1MediaType,
	DockerV2ListMediaType,
}
