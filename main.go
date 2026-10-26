package main

import (
	"embed"
	appassets "wails3-manager/assets"
	"wails3-manager/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	desktop.Run(assets, desktop.Options{
		DMGBackgroundPNG: appassets.InstallGridPNG,
	})
}
