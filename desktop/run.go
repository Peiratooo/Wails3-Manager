package desktop

import (
	"embed"
	"log"

	"wails3-manager/core/contracts"
	"wails3-manager/core/creator"
	"wails3-manager/core/environment"
	"wails3-manager/core/packaging"
	"wails3-manager/core/project"
	"wails3-manager/core/runlog"
	"wails3-manager/core/settings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var App *application.App

type Options struct {
	DMGBackgroundPNG []byte
}

func Run(assets embed.FS, options Options) {
	logSink := runlog.New()
	logSink.SetRecordLogs(settings.LoadManagerSettings().RecordLogs)
	settingsService := settings.NewService(logSink)
	projectService := project.NewService(logSink)
	packagingService := packaging.NewServiceWithOptions(logSink, packaging.ServiceOptions{
		DMGBackgroundPNG: options.DMGBackgroundPNG,
	})
	projectService.InitPackagingFn = func(projectDir string) error {
		_, err := packagingService.InitPackaging(projectDir)
		return err
	}
	settingsService.InitPackagingFn = func(projectDir string) error {
		_, err := packagingService.InitPackaging(projectDir)
		return err
	}
	application.RegisterEvent[contracts.LogLineEvent]("manager:log-line")
	App = application.New(application.Options{
		Name:        "Wails Manager",
		Description: "Universal Wails3 manager",
		Services: []application.Service{
			application.NewService(&AppService{}),
			application.NewServiceWithOptions(creator.NewService(logSink), application.ServiceOptions{Name: "CreatorService"}),
			application.NewServiceWithOptions(projectService, application.ServiceOptions{Name: "ProjectService"}),
			application.NewServiceWithOptions(environment.NewService(), application.ServiceOptions{Name: "EnvironmentService"}),
			application.NewServiceWithOptions(settingsService, application.ServiceOptions{Name: "SettingsService"}),
			application.NewServiceWithOptions(packagingService, application.ServiceOptions{Name: "PackagingService"}),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: LocalFileMiddleware,
		},
		Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	logSink.OnLine = func(line runlog.Line) {
		app := application.Get()
		if app != nil {
			app.Event.Emit("manager:log-line", contracts.LogLineEvent{
				Line:             line.Text,
				TransactionID:    line.TransactionID,
				TransactionType:  line.TransactionType,
				TransactionTitle: line.TransactionTitle,
			})
		}
	}
	App.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Wails Manager",
		Width: 1280, Height: 820, MinWidth: 980, MinHeight: 680,
		URL: "/", EnableFileDrop: true,
		Mac:     application.MacWindow{InvisibleTitleBarHeight: 44, TitleBar: application.MacTitleBarHidden},
		Windows: application.WindowsWindow{BackdropType: 2, DisableFramelessWindowDecorations: false},
	})
	if err := App.Run(); err != nil {
		log.Fatal(err)
	}
}
