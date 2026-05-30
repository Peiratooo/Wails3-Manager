package contracts

import (
	"strconv"
	"strings"
	"time"
)

// UnixTime stores timestamps as Unix seconds in JSON.
type UnixTime int64

func NowUnixTime() UnixTime { return UnixTime(time.Now().Unix()) }

func UnixTimeFrom(t time.Time) UnixTime {
	if t.IsZero() {
		return 0
	}
	return UnixTime(t.Unix())
}

func (t UnixTime) IsZero() bool { return t <= 0 }
func (t UnixTime) Time() time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return time.Unix(int64(t), 0)
}
func (t UnixTime) After(other UnixTime) bool { return int64(t) > int64(other) }

func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(t), 10)), nil
}

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*t = 0
		return nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*t = UnixTime(n)
	return nil
}

type Platform string

const (
	PlatformAuto    Platform = "auto"
	PlatformWindows Platform = "windows"
	PlatformMacOS   Platform = "darwin"
	PlatformAll     Platform = "all"
)

type PackagingConfig struct {
	SchemaVersion int              `json:"schemaVersion"`
	Project       ProjectInfo      `json:"project"`
	Build         BuildSettings    `json:"build"`
	Entry         ProgramEntry     `json:"entry"`
	Assets        []PackagingAsset `json:"assets"`
	Windows       WindowsConfig    `json:"windows"`
	MacOS         MacOSConfig      `json:"macos"`
	Artifacts     ArtifactConfig   `json:"artifacts"`
}

type ProjectInfo struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	BundleID       string `json:"bundleId"`
	Author         string `json:"author"`
	Publisher      string `json:"publisher"`
	Homepage       string `json:"homepage"`
	Copyright      string `json:"copyright"`
	Description    string `json:"description"`
	BuildOutputDir string `json:"buildOutputDir"`
	Icon           string `json:"icon"`
}

type BuildSettings struct {
	Taskfile   string   `json:"taskfile"`
	Task       string   `json:"task"`
	Command    []string `json:"command"`
	Production bool     `json:"production"`
	CGOEnabled bool     `json:"cgoEnabled"`
	AppName    string   `json:"appName"`
}

type ProgramEntry struct {
	ExecutablePath string `json:"executablePath"`
}

type PackagingAsset struct {
	Src      string `json:"src"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type WindowsConfig struct {
	Enabled               bool   `json:"enabled"`
	InnoScript            string `json:"innoScript"`
	ISCCPath              string `json:"isccPath"`
	DefaultDirName        string `json:"defaultDirName"`
	PrivilegesRequired    string `json:"privilegesRequired"`
	SetupIcon             string `json:"setupIcon"`
	OutputBaseName        string `json:"outputBaseName"`
	CreateDesktopShortcut bool   `json:"createDesktopShortcut"`
}

type MacOSConfig struct {
	Enabled       bool   `json:"enabled"`
	AppBundle     string `json:"appBundle"`
	DMGScript     string `json:"dmgScript"`
	Background    string `json:"background"`
	OutputName    string `json:"outputName"`
	CreateDMGPath string `json:"createDmgPath"`
	WindowWidth   int    `json:"windowWidth"`
	WindowHeight  int    `json:"windowHeight"`
	IconSize      int    `json:"iconSize"`
	AppX          int    `json:"appX"`
	AppY          int    `json:"appY"`
	ApplicationsX int    `json:"applicationsX"`
	ApplicationsY int    `json:"applicationsY"`
}

type ArtifactConfig struct {
	OutputRoot string `json:"outputRoot"`
}

type ScanResult struct {
	ProjectDir     string `json:"projectDir"`
	Score          int    `json:"score"`
	IsWailsProject bool   `json:"isWailsProject"`
}

type Artifact struct {
	Kind      string   `json:"kind"`
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Size      int64    `json:"size"`
	CreatedAt UnixTime `json:"createdAt"`
}

type CommandResult struct {
	OK             bool            `json:"ok"`
	Message        string          `json:"message"`
	Logs           []string        `json:"logs"`
	Artifacts      []Artifact      `json:"artifacts"`
	ArtifactChecks []ArtifactCheck `json:"artifactChecks,omitempty"`
	Warnings       []string        `json:"warnings"`
	RunID          string          `json:"runId,omitempty"`
	CacheDir       string          `json:"cacheDir,omitempty"`
	ReleaseDir     string          `json:"releaseDir,omitempty"`
	BuildArtifacts []Artifact      `json:"buildArtifacts,omitempty"`
	FinalArtifacts []Artifact      `json:"finalArtifacts,omitempty"`
}

type ArtifactCheck struct {
	Platform Platform `json:"platform"`
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Found    bool     `json:"found"`
	Message  string   `json:"message"`
}

type ToolRequirement struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Platform       Platform `json:"platform"`
	Command        string   `json:"command"`
	Required       bool     `json:"required"`
	Found          bool     `json:"found"`
	Path           string   `json:"path,omitempty"`
	ConfiguredPath string   `json:"configuredPath,omitempty"`
	Version        string   `json:"version,omitempty"`
	Message        string   `json:"message,omitempty"`
	InstallHint    string   `json:"installHint,omitempty"`
	DownloadURL    string   `json:"downloadUrl,omitempty"`
	InstallCommand []string `json:"installCommand,omitempty"`
	CanAutoInstall bool     `json:"canAutoInstall"`
	CanChoosePath  bool     `json:"canChoosePath"`
}

type ToolCheck struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Command  string   `json:"command"`
	Found    bool     `json:"found"`
	Version  string   `json:"version,omitempty"`
	Path     string   `json:"path,omitempty"`
	Required bool     `json:"required"`
	Platform Platform `json:"platform,omitempty"`
}

type EnvironmentReport struct {
	OS     string      `json:"os"`
	Arch   string      `json:"arch"`
	OK     bool        `json:"ok"`
	Checks []ToolCheck `json:"checks"`
}

type ValidationSeverity string

const (
	SeverityOK      ValidationSeverity = "ok"
	SeverityWarning ValidationSeverity = "warning"
	SeverityError   ValidationSeverity = "error"
)

type ValidationCheck struct {
	Key      string             `json:"key"`
	Title    string             `json:"title"`
	Severity ValidationSeverity `json:"severity"`
	Message  string             `json:"message"`
	Path     string             `json:"path,omitempty"`
}

type ValidationReport struct {
	OK       bool              `json:"ok"`
	Checks   []ValidationCheck `json:"checks"`
	Warnings int               `json:"warnings"`
	Errors   int               `json:"errors"`
}

type WailsProjectConfig struct {
	Info             WailsAppInfo           `json:"info"`
	Icon             string                 `json:"icon"`
	FileAssociations []WailsFileAssociation `json:"fileAssociations,omitempty"`
}

type WailsAppInfo struct {
	CompanyName       string `json:"companyName"`
	ProductName       string `json:"productName"`
	ProductIdentifier string `json:"productIdentifier"`
	Description       string `json:"description"`
	Copyright         string `json:"copyright"`
	Comments          string `json:"comments"`
	Version           string `json:"version"`
}

type WailsTaskVars struct {
	AppName    string `json:"appName"`
	Production bool   `json:"production"`
	CGOEnabled string `json:"cgoEnabled"`
}

type WailsFileAssociation struct {
	Ext         string `json:"ext"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconName    string `json:"iconName"`
	Role        string `json:"role"`
	MimeType    string `json:"mimeType,omitempty"`
}

type WailsProjectManager struct {
	ProjectDir      string             `json:"projectDir"`
	CurrentPlatform Platform           `json:"currentPlatform"`
	WailsConfig     WailsProjectConfig `json:"wailsConfig"`
	TaskVars        WailsTaskVars      `json:"taskVars"`
}

type ProjectRecord struct {
	ProjectDir   string              `json:"projectDir"`
	Project      WailsProjectManager `json:"project"`
	ImportedAt   UnixTime            `json:"importedAt"`
	LastOpenedAt UnixTime            `json:"lastOpenedAt"`
}

type UserState struct {
	Projects []ProjectRecord `json:"projects"`
}

type ManagerSettings struct {
	Theme      string `json:"theme"`
	Language   string `json:"language"`
	RecordLogs bool   `json:"recordLogs"`
}

type PackageRequest struct {
	ProjectDir string   `json:"projectDir"`
	Platform   Platform `json:"platform"`
	DryRun     bool     `json:"dryRun"`
	RunBuild   bool     `json:"runBuild"`
}

type PackageResult struct {
	OK        bool       `json:"ok"`
	Message   string     `json:"message"`
	RunID     string     `json:"runId,omitempty"`
	Artifacts []Artifact `json:"artifacts"`
	Warnings  []string   `json:"warnings,omitempty"`
}

type LogEntry struct {
	Line      string   `json:"line"`
	CreatedAt UnixTime `json:"createdAt"`
}

type LogSnapshot struct {
	Cursor int      `json:"cursor"`
	Lines  []string `json:"lines"`
}
