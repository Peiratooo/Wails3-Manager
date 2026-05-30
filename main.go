package main

import (
	"embed"
	"wails3-manager/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	desktop.Run(assets)
}
