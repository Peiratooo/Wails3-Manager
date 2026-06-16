import * as CreatorService from '../../bindings/wails3-manager/core/creator/creatorservice'
import * as EnvironmentService from '../../bindings/wails3-manager/core/environment/environmentservice'
import * as PackagingService from '../../bindings/wails3-manager/core/packaging/packagingservice'
import * as ProjectService from '../../bindings/wails3-manager/core/project/projectservice'
import * as SettingsService from '../../bindings/wails3-manager/core/settings/settingsservice'

export const appService = {
	creator: CreatorService,
	project: ProjectService,
	settings: SettingsService,
	environment: EnvironmentService,
	packaging: PackagingService,

	CreateWailsProject: CreatorService.CreateWailsProject,

	ScanProject: ProjectService.ScanProject,
	ImportProject: ProjectService.ImportProject,
	SaveProject: ProjectService.SaveProject,
	ReplaceProjectIcon: ProjectService.ReplaceProjectIcon,

	OpenProject: SettingsService.OpenProject,
	ListProjects: SettingsService.ListProjects,
	RemoveProject: SettingsService.RemoveProject,
	GetSettings: SettingsService.GetSettings,
	SaveSettings: SettingsService.SaveSettings,
	ClearLogs: SettingsService.ClearLogs,

	CheckEnvironment: EnvironmentService.CheckEnvironment,

	InitPackaging: PackagingService.InitPackaging,
	LoadPackagingConfig: PackagingService.LoadPackagingConfig,
	SavePackagingConfig: PackagingService.SavePackagingConfig,
	GetPackagingRuntimeInfo: PackagingService.GetPackagingRuntimeInfo,
	Package: PackagingService.Package,
	Artifacts: PackagingService.Artifacts,
}
