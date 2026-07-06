package main

import (
	"embed"
	"wails3-manager/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/install-grid.png
var installGridPNG []byte

func main() {
	desktop.Run(assets, desktop.Options{
		DMGBackgroundPNG: installGridPNG,
	})
}
