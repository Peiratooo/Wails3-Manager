package project

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
)

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

func writeProjectIconPNGBase64(projectDir, imageBase64 string) error {
	data, err := decodePNGBase64(imageBase64)
	if err != nil {
		return err
	}

	target := filepath.Join(projectDir, DefaultAppIconRelPath)

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	return os.WriteFile(target, data, 0644)
}

func decodePNGBase64(imageBase64 string) ([]byte, error) {
	raw, err := normalizePNGBase64(imageBase64)
	if err != nil {
		return nil, err
	}

	data, err := decodeBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PNG base64: %w", err)
	}

	if !bytes.HasPrefix(data, pngSignature) {
		return nil, fmt.Errorf("only PNG images are supported")
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid PNG image: %w", err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, fmt.Errorf("invalid PNG dimensions")
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG image: %w", err)
	}

	targetSize := chooseIconSquareSize(cfg.Width, cfg.Height)

	out, err := resizePNGContainToSquare(img, targetSize)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func normalizePNGBase64(imageBase64 string) (string, error) {
	raw := strings.TrimSpace(imageBase64)
	if raw == "" {
		return "", fmt.Errorf("PNG base64 is required")
	}

	if strings.HasPrefix(strings.ToLower(raw), "data:") {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid data URL")
		}

		meta := strings.ToLower(parts[0])
		if !strings.Contains(meta, "image/png") || !strings.Contains(meta, ";base64") {
			return "", fmt.Errorf("only PNG base64 images are supported")
		}

		raw = parts[1]
	}

	raw = strings.NewReplacer(
		"\r", "",
		"\n", "",
		"\t", "",
		" ", "",
	).Replace(raw)

	if raw == "" {
		return "", fmt.Errorf("PNG base64 is required")
	}

	return raw, nil
}

func decodeBase64(raw string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var lastErr error

	for _, encoding := range encodings {
		data, err := encoding.DecodeString(raw)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, lastErr
}

func chooseIconSquareSize(width, height int) int {
	maxSide := width
	if height > maxSide {
		maxSide = height
	}

	switch {
	case maxSide <= 64:
		return 64
	case maxSide <= 128:
		return 128
	case maxSide <= 256:
		return 256
	case maxSide <= 512:
		return 512
	default:
		return 1024
	}
}

func resizePNGContainToSquare(src image.Image, size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid target size")
	}

	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 {
		return nil, fmt.Errorf("invalid PNG dimensions")
	}

	scale := math.Min(
		float64(size)/float64(srcWidth),
		float64(size)/float64(srcHeight),
	)

	dstWidth := int(math.Round(float64(srcWidth) * scale))
	dstHeight := int(math.Round(float64(srcHeight) * scale))

	if dstWidth < 1 {
		dstWidth = 1
	}
	if dstHeight < 1 {
		dstHeight = 1
	}
	if dstWidth > size {
		dstWidth = size
	}
	if dstHeight > size {
		dstHeight = size
	}

	offsetX := (size - dstWidth) / 2
	offsetY := (size - dstHeight) / 2

	dstRect := image.Rect(
		offsetX,
		offsetY,
		offsetX+dstWidth,
		offsetY+dstHeight,
	)

	// NewRGBA starts with a transparent background.
	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	xdraw.CatmullRom.Scale(
		dst,
		dstRect,
		src,
		srcBounds,
		xdraw.Over,
		nil,
	)

	var buf bytes.Buffer

	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("failed to re-encode PNG image: %w", err)
	}

	return buf.Bytes(), nil
}
