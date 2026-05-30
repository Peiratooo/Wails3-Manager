package desktop

import (
	"embed"
	"log"

	"wails3-manager/core/environment"
	"wails3-manager/core/packaging"
	"wails3-manager/core/project"
	"wails3-manager/core/runlog"
	"wails3-manager/core/settings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func Run(assets embed.FS) {
	logSink := runlog.New()
	settingsService := settings.NewService(logSink)
	projectService := project.NewService(logSink)
	app := application.New(application.Options{
		Name:        "Wails Manager",
		Description: "Universal Wails3 manager",
		Services: []application.Service{
			application.NewServiceWithOptions(projectService, application.ServiceOptions{Name: "ProjectService"}),
			application.NewServiceWithOptions(environment.NewService(), application.ServiceOptions{Name: "EnvironmentService"}),
			application.NewServiceWithOptions(settingsService, application.ServiceOptions{Name: "SettingsService"}),
			application.NewServiceWithOptions(packaging.NewService(logSink), application.ServiceOptions{Name: "PackagingService"}),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: LocalFileMiddleware,
		},
		Mac:    application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Wails Manager",
		Width: 1280, Height: 820, MinWidth: 980, MinHeight: 680,
		URL: "/", EnableFileDrop: true,
		Mac:     application.MacWindow{InvisibleTitleBarHeight: 44, TitleBar: application.MacTitleBarHidden},
		Windows: application.WindowsWindow{BackdropType: 2, DisableFramelessWindowDecorations: false},
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
