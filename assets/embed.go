package assets

import _ "embed"

// InstallGridPNG is shared by the default DMG generator and the desktop app.
//
//go:embed install-grid.png
var InstallGridPNG []byte
