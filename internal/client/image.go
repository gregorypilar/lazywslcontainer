package client

import (
	"context"
	"fmt"
)

type Image struct {
	ID         string `json:"Id"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	Size       int64  `json:"Size"`
	Created    int64  `json:"Created"`
}

func (i Image) SizeString() string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)
	switch {
	case i.Size >= GB:
		return fmt.Sprintf("%.2fGB", float64(i.Size)/float64(GB))
	case i.Size >= MB:
		return fmt.Sprintf("%.2fMB", float64(i.Size)/float64(MB))
	case i.Size >= KB:
		return fmt.Sprintf("%.2fKB", float64(i.Size)/float64(KB))
	default:
		return fmt.Sprintf("%dB", i.Size)
	}
}

// Images calls `wslc image list --format json`.
func (w *WSLC) Images(ctx context.Context) ([]Image, error) {
	var imgs []Image
	err := w.runJSON(ctx, &imgs, "image", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	return imgs, nil
}

// ImageInspect returns raw inspect output for the given image (id or name:tag).
func (w *WSLC) ImageInspect(ctx context.Context, id string) ([]byte, error) {
	return w.run(ctx, "image", "inspect", id)
}

// ImageRemove removes an image by id or name:tag.
func (w *WSLC) ImageRemove(ctx context.Context, id string, force bool) error {
	args := []string{"image", "rm", id}
	if force {
		args = append(args, "--force")
	}
	_, err := w.run(ctx, args...)
	return err
}

// ImagePrune removes unused images.
func (w *WSLC) ImagePrune(ctx context.Context) ([]byte, error) {
	return w.run(ctx, "image", "prune")
}

// Build builds an image from a Containerfile at the given path with the
// provided tag. If tag is empty, no `-t` is passed.
func (w *WSLC) Build(ctx context.Context, path, tag string) ([]byte, error) {
	args := []string{"build"}
	if tag != "" {
		args = append(args, "-t", tag)
	}
	args = append(args, path)
	return w.run(ctx, args...)
}
