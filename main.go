package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Media types
const (
	MediaTypeConfig   = "application/vnd.docker.container.image.v1+json"
	MediaTypeManifest = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeLayer    = "application/vnd.docker.image.rootfs.diff.tar.gzip"
)

// Header names
const (
	ContentType         = "Content-Type"
	DockerContentDigest = "Docker-Content-Digest"
)

func main() {
	println("Starting quine")

	registry, err := NewRegistry()
	if err != nil {
		panic(err)
	}

	println("config", registry.ConfigDigest)
	println("manifest", registry.ManifestDigest)
	println("layer", registry.LayerDigest)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/v2/quine/manifests/{tag}", func(w http.ResponseWriter, r *http.Request) {
		switch chi.URLParam(r, "tag") {
		case "latest":
			w.Header().Set(ContentType, MediaTypeManifest)
			w.Header().Add(DockerContentDigest, registry.ManifestDigest)
			w.Write(registry.Manifest)
		case registry.ManifestDigest:
			w.Header().Set(ContentType, MediaTypeManifest)
			w.Header().Add(DockerContentDigest, registry.ManifestDigest)
			w.Write(registry.Manifest)
		default:
			http.Error(w, "{}", 404)
		}
	})

	r.Get("/v2/quine/blobs/{blob}", func(w http.ResponseWriter, r *http.Request) {
		switch chi.URLParam(r, "blob") {
		case registry.ConfigDigest:
			w.Header().Set("Content-Type", MediaTypeConfig)
			w.Header().Set(DockerContentDigest, registry.ConfigDigest)
			w.Write(registry.Config)
		case registry.LayerDigest:
			w.Header().Add(DockerContentDigest, registry.LayerDigest)
			w.Write(registry.Layer)
		default:
			http.Error(w, http.StatusText(404), 404)
		}
	})

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

type Registry struct {
	Manifest       []byte
	ManifestDigest string
	Config         []byte
	ConfigDigest   string
	Layer          []byte
	LayerDigest    string
}

func NewRegistry() (*Registry, error) {
	a, err := archive()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	config := Config{
		Created:      now,
		Author:       "quine",
		Architecture: "amd64",
		OS:           "linux",
		Config: &ImageConfig{
			Cmd: []string{"/quine"},
			Env: []string{"PATH=/"},
		},
		RootFS: RootFS{
			DiffIDs: []string{digest(a)},
			Type:    "layers",
		},
		History: []History{{
			Created:   now,
			CreatedBy: "quine",
		}},
	}

	configJson, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	configDigest := digest(configJson)

	layer, err := compress(a)
	if err != nil {
		return nil, err
	}

	layerDigest := digest(layer)

	manifest := Manifest{
		SchemaVersion: 2,
		MediaType:     MediaTypeManifest,
		Config: Layer{
			MediaType: MediaTypeConfig,
			Size:      len(configJson),
			Digest:    configDigest,
		},
		Layers: []Layer{{
			MediaType: MediaTypeLayer,
			Size:      len(layer),
			Digest:    layerDigest,
		}},
	}

	manifestJson, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	r := Registry{
		Layer:          layer,
		LayerDigest:    layerDigest,
		Manifest:       manifestJson,
		ManifestDigest: digest(manifestJson),
		Config:         configJson,
		ConfigDigest:   configDigest,
	}

	return &r, nil
}

func digest(p []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(p))
}

func archive() ([]byte, error) {
	e, err := os.Executable()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(e)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	w := tar.NewWriter(&buf)
	err = w.WriteHeader(&tar.Header{
		Name: "quine", Mode: 0755, Size: info.Size(),
	})
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(w, f); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func compress(p []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(p); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// https://github.com/docker/docker/blob/master/image/spec/v1.2.md#image-json-description
type Config struct {
	Created      time.Time    `json:"created"`
	Author       string       `json:"author"`
	Architecture string       `json:"architecture"`
	OS           string       `json:"os"`
	Config       *ImageConfig `json:"config"`
	// Filesystem layers and history elements have to be in the same order
	RootFS  RootFS    `json:"rootfs"`
	History []History `json:"history"`
}

type RootFS struct {
	DiffIDs []string `json:"diff_ids"`
	Type    string   `json:"type"`
}

type History struct {
	Created   time.Time `json:"created"`
	CreatedBy string    `json:"created_by"`
}

type ImageConfig struct {
	Cmd []string
	Env []string
}

type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        Layer   `json:"config"`
	Layers        []Layer `json:"layers"`
}

type Layer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}
