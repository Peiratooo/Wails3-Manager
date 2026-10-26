package packaging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
	"howett.net/plist"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
	"wails3-manager/core/packaging/dmg"
	"wails3-manager/core/runlog"
)

func ensureProjectBundleDestination(projectDir, destination string) error {
	relative, err := filepath.Rel(projectDir, destination)
	if err != nil || !filepath.IsLocal(relative) || relative == "." {
		return fmt.Errorf("app bundle output must be inside the project: %s", destination)
	}
	return ensureBundleWritePath(projectDir, destination)
}

// Resolve existing ancestors before any write, including writes through symlinks.
func ensureBundleWritePath(root, destination string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	parent := destination
	for {
		_, err := os.Lstat(parent)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return err
		}
		next := filepath.Dir(parent)
		if next == parent {
			return fmt.Errorf("invalid output path: %s", destination)
		}
		parent = next
	}
	resolved, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil || (relative != "." && !filepath.IsLocal(relative)) {
		return fmt.Errorf("bundle write escapes its staging directory: %s", destination)
	}
	return nil
}

// Only discover an explicitly included Taskfile; never replace a user's tasks.
func darwinBundleTask(projectDir string, cfg contracts.PackagingConfig) bool {
	data, err := os.ReadFile(fsx.ResolveProjectFile(projectDir, cfg.Build.Taskfile))
	if err != nil {
		return false
	}
	var root struct{ Includes map[string]any }
	if yaml.Unmarshal(data, &root) != nil {
		return false
	}
	var path string
	switch value := root.Includes["darwin"].(type) {
	case string:
		path = value
	case map[string]any:
		path, _ = value["taskfile"].(string)
	}
	if path == "" {
		return false
	}
	path = filepath.Join(filepath.Dir(fsx.Resolve(projectDir, cfg.Build.Taskfile)), path)
	data, err = os.ReadFile(path)
	if err != nil {
		return false
	}
	var tasks struct{ Tasks map[string]any }
	if yaml.Unmarshal(data, &tasks) != nil {
		return false
	}
	_, found := tasks.Tasks["create:app:bundle"]
	return found
}

func (s *Service) stageMacOSBundle(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, tx runlog.Transaction, runBuild bool) (string, func(), error) {
	source := fsx.Resolve(projectDir, config.MacOSAppBundlePath(cfg, projectConfig))
	selected := config.RenderPlaceholders(cfg.Entry.ExecutablePath, cfg, projectConfig)
	selectedBundle := isMacOSAppBundlePath(selected)
	if selectedBundle {
		source = fsx.Resolve(projectDir, selected)
		if !fsx.DirExists(source) {
			return "", nil, fmt.Errorf("selected application bundle does not exist: %s", source)
		}
	}
	if !isMacOSAppBundlePath(source) {
		return "", nil, fmt.Errorf("macOS app bundle path must end with .app: %s", source)
	}
	officialBundle := false
	// Standard build outputs can use the project's official bundle task after
	// its binary build. An explicitly selected external bundle stays untouched.
	useBuiltBinary := runBuild && !selectedBundle && config.ResolveMacOSAppBundlePath(cfg, projectConfig).UsingDefaultExecutable
	if useBuiltBinary && darwinBundleTask(projectDir, cfg) {
		command := []string{"wails3", "task", "-taskfile", cfg.Build.Taskfile, "darwin:create:app:bundle"}
		command = append(command, buildTaskVars(cfg)...)
		if err := (runlog.Runner{Log: s.Log, Transaction: tx}).Run(ctx, projectDir, command); err != nil {
			return "", nil, err
		}
		source = fsx.Resolve(projectDir, "bin/"+config.AppName(cfg)+".app")
		officialBundle = true
	}
	stage, err := os.MkdirTemp("", "wails-manager-bundle-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(stage) }
	bundle := filepath.Join(stage, config.MacOSAppBundleName(projectConfig)+".app")
	fail := func(err error) (string, func(), error) { cleanup(); return "", nil, err }
	modified := false
	if fsx.DirExists(source) {
		if !fsx.FileExists(filepath.Join(source, "Contents", "Info.plist")) {
			return fail(fmt.Errorf("invalid application bundle: %s", source))
		}
		if runtime.GOOS == "darwin" {
			if err := (runlog.Runner{Log: s.Log, Transaction: tx}).Run(ctx, projectDir, []string{"ditto", source, bundle}); err != nil {
				return fail(err)
			}
		} else if err := fsx.CopyDir(source, bundle, true); err != nil {
			return fail(err)
		}
		customLaunch := !config.SameAssetPath(macOSLaunchExecutablePath(cfg, projectConfig), config.DefaultMacOSBinaryPath(cfg), contracts.PlatformMacOS)
		// A freshly built bundle can still contain the template's old executable
		// name or version. Refresh its metadata even when the official task ran.
		modified = len(cfg.Assets) > 0 || customLaunch || useBuiltBinary
		if modified {
			if err := filepath.WalkDir(bundle, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.Type()&os.ModeSymlink != 0 {
					return ensureBundleWritePath(stage, path)
				}
				return nil
			}); err != nil {
				return fail(err)
			}
			contents := filepath.Join(bundle, "Contents")
			for _, path := range []string{contents, filepath.Join(contents, "MacOS"), filepath.Join(contents, "Info.plist"), filepath.Join(contents, "_CodeSignature")} {
				if err := ensureBundleWritePath(stage, path); err != nil {
					return fail(err)
				}
			}
			if err := os.RemoveAll(filepath.Join(contents, "_CodeSignature")); err != nil {
				return fail(err)
			}
			if useBuiltBinary && !officialBundle {
				if err := ensureBundleWritePath(stage, filepath.Join(contents, "MacOS", config.AppName(cfg))); err != nil {
					return fail(err)
				}
				if err := copyRequiredFile(fsx.Resolve(projectDir, config.DefaultMacOSBinaryPath(cfg)), filepath.Join(contents, "MacOS", config.AppName(cfg)), "macOS executable", 0755); err != nil {
					return fail(err)
				}
			}
			if err := copyMacOSPayloads(projectDir, contents, cfg, projectConfig); err != nil {
				return fail(err)
			}
			if customLaunch {
				if err := copyMacOSLaunchExecutable(projectDir, filepath.Join(contents, "MacOS"), macOSLaunchExecutablePath(cfg, projectConfig)); err != nil {
					return fail(err)
				}
			}
			if err := updateStagedBundlePlist(bundle, cfg, projectConfig, useBuiltBinary || customLaunch); err != nil {
				return fail(err)
			}
		}
	} else {
		if !config.ResolveMacOSAppBundlePath(cfg, projectConfig).UsingDefaultExecutable {
			return fail(fmt.Errorf("configured application bundle does not exist: %s", source))
		}
		if err := populateMacOSAppBundle(projectDir, cfg, projectConfig, bundle); err != nil {
			return fail(err)
		}
		modified = true
	}
	if modified && runtime.GOOS == "darwin" {
		runner := runlog.Runner{Log: s.Log, Transaction: tx}
		for _, command := range [][]string{
			{"xattr", "-cr", bundle},
			{"codesign", "--force", "--deep", "--sign", "-", bundle},
			{"codesign", "--verify", "--deep", "--strict", bundle},
		} {
			if err := runner.Run(ctx, projectDir, command); err != nil {
				return fail(err)
			}
		}
		if s.Log != nil {
			s.Log.PrintlnWithTransaction(tx, "Staged application was ad-hoc signed; Developer ID signing and notarization are required for public distribution.")
		}
	}
	return bundle, cleanup, nil
}

func updateStagedBundlePlist(bundle string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, changeExecutable bool) error {
	path := filepath.Join(bundle, "Contents", "Info.plist")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var values map[string]any
	if _, err := plist.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("invalid app Info.plist: %w", err)
	}
	if changeExecutable {
		values["CFBundleExecutable"] = macOSBundleExecutableName(cfg, projectConfig)
	}
	values["CFBundleName"] = config.ProjectName(projectConfig)
	values["CFBundleDisplayName"] = config.ProjectName(projectConfig)
	values["CFBundleVersion"] = config.ProjectVersion(projectConfig)
	values["CFBundleShortVersionString"] = config.ProjectVersion(projectConfig)
	if id := strings.TrimSpace(config.ProjectBundleID(projectConfig)); id != "" {
		values["CFBundleIdentifier"] = id
	}
	data, err = plist.Marshal(values, plist.XMLFormat)
	if err != nil {
		return err
	}
	return fsx.WriteFileAtomic(path, data, 0644)
}

func (s *Service) runOfficialDMG(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, bundle, background string, tx runlog.Transaction) (string, error) {
	stage := filepath.Dir(bundle)
	name := strings.TrimSuffix(filepath.Base(bundle), ".app")
	command := []string{"wails3", "tool", "package", "--format", "dmg", "--name", name, "--out", stage,
		"--window-width", fmt.Sprint(cfg.MacOS.WindowWidth), "--window-height", fmt.Sprint(dmg.FinderWindowHeight(cfg.MacOS.WindowHeight))}
	if background != "" {
		command = append(command, "--background", background)
	}
	for _, icon := range []struct{ flag, path string }{{"--volume-icon", cfg.MacOS.VolumeIcon}, {"--file-icon", cfg.MacOS.FileIcon}} {
		if icon.path != "" {
			command = append(command, icon.flag, fsx.Resolve(projectDir, config.RenderPlaceholders(icon.path, cfg, projectConfig)))
		}
	}
	if err := (runlog.Runner{Log: s.Log, Transaction: tx}).Run(ctx, projectDir, command); err != nil {
		return "", err
	}
	output := filepath.Join(packageOutputDir(projectDir, cfg, projectConfig, contracts.PlatformMacOS), config.InstallerOutputName(cfg, projectConfig, contracts.PlatformMacOS)+".dmg")
	if err := fsx.PublishFile(filepath.Join(stage, name+".dmg"), output); err != nil {
		return "", err
	}
	return output, nil
}
