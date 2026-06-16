package project

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const onePixelPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII="

func TestWriteProjectIconPNGBase64SupportsDataURL(t *testing.T) {
	projectDir := t.TempDir()
	input := "data:image/png;base64," + onePixelPNGBase64

	if err := writeProjectIconPNGBase64(projectDir, input); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, DefaultAppIconRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, pngSignature) {
		t.Fatalf("written icon is not a PNG")
	}
}

func TestDecodePNGBase64RejectsNonPNG(t *testing.T) {
	if _, err := decodePNGBase64("bm90IGEgcG5n"); err == nil {
		t.Fatal("expected non-PNG base64 to be rejected")
	}
}
