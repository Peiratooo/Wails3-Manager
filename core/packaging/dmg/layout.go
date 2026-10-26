package dmg

// FinderTitleBarHeight is the height of Finder's title bar with its toolbar
// hidden, as configured by the DMG library's .DS_Store window settings.
const FinderTitleBarHeight = 32

// FinderWindowHeight converts the configured background/content height to the
// outer window height expected by both the DMG library and wails3 tool package.
func FinderWindowHeight(contentHeight int) int {
	return contentHeight + FinderTitleBarHeight
}
