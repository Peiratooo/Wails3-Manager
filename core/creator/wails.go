package creator

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/project"
)

const defaultWailsBinary = "wails3"

type WailsTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateWailsProjectRequest struct {
	Name                      string `json:"name"`
	Dir                       string `json:"dir"`
	Template                  string `json:"template"`
	PackageName               string `json:"packageName"`
	ProductName               string `json:"productName"`
	ProductDescription        string `json:"productDescription"`
	ProductVersion            string `json:"productVersion"`
	ProductCompany            string `json:"productCompany"`
	ProductCopyright          string `json:"productCopyright"`
	ProductComments           string `json:"productComments"`
	ProductIdentifier         string `json:"productIdentifier"`
	UseInterfaces             bool   `json:"useInterfaces"`
	Quiet                     bool   `json:"quiet"`
	SkipRemoteTemplateWarning bool   `json:"skipRemoteTemplateWarning"`
}

var ansiEscapePattern = regexp.MustCompile(
	`[\x1B\x9B][[\]()#;?]*(?:(?:(?:[a-zA-Z\d]*(?:;[a-zA-Z\d]*)*)?\x07)|(?:(?:\d{1,4}(?:;\d{0,4})*)?[\dA-PR-TZcf-nq-uy=><~]))`,
)

var projectDirectoryNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func CreateWailsProject(ctx context.Context, req CreateWailsProjectRequest, runner CommandRunner) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if runner == nil {
		return "", fmt.Errorf("command runner is required")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return "", fmt.Errorf("project name is required")
	}
	req.Dir = strings.TrimSpace(strings.Trim(req.Dir, `"`))
	req.Template = strings.TrimSpace(req.Template)
	if req.Template == "" {
		req.Template = "vanilla"
	}
	req.PackageName = strings.TrimSpace(req.PackageName)
	if req.PackageName == "" {
		req.PackageName = "main"
	}
	req.ProductName = strings.TrimSpace(req.ProductName)
	req.ProductDescription = strings.TrimSpace(req.ProductDescription)
	req.ProductVersion = strings.TrimSpace(req.ProductVersion)
	if req.ProductVersion == "" {
		return "", fmt.Errorf("product version is required")
	}
	req.ProductCompany = strings.TrimSpace(req.ProductCompany)
	req.ProductCopyright = strings.TrimSpace(req.ProductCopyright)
	req.ProductComments = strings.TrimSpace(req.ProductComments)
	req.ProductIdentifier = strings.TrimSpace(req.ProductIdentifier)

	targetDir, err := targetProjectDir(req)
	if err != nil {
		return "", err
	}
	if err := requireEmptyTargetDirectory(targetDir); err != nil {
		return "", err
	}
	initDir, err := initDirectory(req)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(initDir, 0755); err != nil {
		return "", err
	}

	command := buildInitCommand(req, initDir)
	if err := runner.Run(ctx, initDir, command); err != nil {
		return "", err
	}
	if !fsx.DirExists(targetDir) {
		return "", fmt.Errorf("wails3 init completed but project directory was not created: %s", targetDir)
	}
	if err := syncProjectInfo(ctx, runner, targetDir, req); err != nil {
		return "", err
	}
	return targetDir, nil
}

func targetProjectDir(req CreateWailsProjectRequest) (string, error) {
	name := strings.TrimSpace(req.Name)
	if err := validateProjectDirectoryName(name); err != nil {
		return "", err
	}

	parentDir := strings.TrimSpace(strings.Trim(req.Dir, `"`))
	if parentDir != "" {
		return fsx.NormalizePath(filepath.Join(parentDir, name))
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fsx.NormalizePath(filepath.Join(cwd, name))
}

func initDirectory(req CreateWailsProjectRequest) (string, error) {
	dir := strings.TrimSpace(strings.Trim(req.Dir, `"`))
	if dir != "" {
		return fsx.NormalizePath(dir)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fsx.NormalizePath(cwd)
}

func validateProjectDirectoryName(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if name == "." || name == ".." ||
		strings.HasSuffix(name, ".") ||
		!projectDirectoryNamePattern.MatchString(name) ||
		isWindowsReservedDirectoryName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}
	return nil
}

func isWindowsReservedDirectoryName(name string) bool {
	base := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	return (strings.HasPrefix(base, "COM") ||
		strings.HasPrefix(base, "LPT")) &&
		len(base) == 4 &&
		base[3] >= '1' &&
		base[3] <= '9'
}

func requireEmptyTargetDirectory(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err == nil {
		if len(entries) > 0 {
			return fmt.Errorf("target directory is not empty: %s", targetDir)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func buildInitCommand(req CreateWailsProjectRequest, targetDir string) []string {
	command := []string{
		defaultWailsBinary,
		"init",
		"-n", req.Name,
		"-d", targetDir,
		"-t", req.Template,
		"-p", req.PackageName,
		"-nocolour",
		fmt.Sprintf("-useinterfaces=%t", req.UseInterfaces),
	}

	command = appendFlag(command, "-productname", req.ProductName)
	command = appendFlag(command, "-productdescription", req.ProductDescription)
	command = appendFlag(command, "-productversion", req.ProductVersion)
	command = appendFlag(command, "-productcompany", req.ProductCompany)
	command = appendFlag(command, "-productcopyright", req.ProductCopyright)
	command = appendFlag(command, "-productcomments", req.ProductComments)
	command = appendFlag(command, "-productidentifier", req.ProductIdentifier)
	if req.Quiet {
		command = append(command, "-q")
	}
	if req.SkipRemoteTemplateWarning {
		command = append(command, "-s")
	}
	return command
}

func appendFlag(command []string, flag string, value string) []string {
	if strings.TrimSpace(value) == "" {
		return command
	}
	return append(command, flag, value)
}

func syncProjectInfo(ctx context.Context, runner CommandRunner, projectDir string, req CreateWailsProjectRequest) error {
	cfg, err := project.LoadWailsConfig(projectDir)
	if err != nil {
		return err
	}
	cfg.Info = projectInfoFromRequest(cfg.Info, req)
	if _, err := project.SaveWailsConfig(projectDir, cfg); err != nil {
		return err
	}
	if err := runner.Run(ctx, projectDir, []string{"wails3", "task", "common:update:build-assets"}); err != nil {
		return fmt.Errorf("failed to update Wails build assets: %w", err)
	}
	return nil
}

func projectInfoFromRequest(info contracts.WailsAppInfo, req CreateWailsProjectRequest) contracts.WailsAppInfo {
	info.ProductName = firstNonEmpty(req.ProductName, info.ProductName)
	info.Description = firstNonEmpty(req.ProductDescription, info.Description)
	info.Version = firstNonEmpty(req.ProductVersion, info.Version)
	info.CompanyName = firstNonEmpty(req.ProductCompany, info.CompanyName)
	info.Copyright = firstNonEmpty(req.ProductCopyright, info.Copyright)
	info.Comments = firstNonEmpty(req.ProductComments, info.Comments)
	info.ProductIdentifier = firstNonEmpty(req.ProductIdentifier, info.ProductIdentifier)
	return info
}

func firstNonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func ListWailsTemplates() ([]WailsTemplate, error) {

	ctx := context.Background()

	command := exec.CommandContext(
		ctx,
		defaultWailsBinary,
		"init",
		"-l",
		"-nocolour",
	)

	output, err := command.CombinedOutput()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf(
				"list Wails templates cancelled: %w",
				ctxErr,
			)
		}

		message := strings.TrimSpace(string(output))
		if message == "" {
			message = "no command output"
		}

		return nil, fmt.Errorf(
			"execute %q init -l failed: %w: %s",
			defaultWailsBinary,
			err,
			message,
		)
	}

	templates, err := ParseWailsTemplateList(string(output))
	if err != nil {
		return nil, fmt.Errorf(
			"parse Wails template list: %w",
			err,
		)
	}

	return templates, nil
}

func ParseWailsTemplateList(output string) ([]WailsTemplate, error) {
	output = stripANSI(output)

	scanner := bufio.NewScanner(strings.NewReader(output))

	scanner.Buffer(make([]byte, 1024), 1024*1024)

	templates := make([]WailsTemplate, 0, 16)
	seen := make(map[string]struct{})

	for scanner.Scan() {
		line := normaliseTableLine(scanner.Text())

		template, ok := parseTemplateTableRow(line)
		if !ok {
			continue
		}

		// 跳过表头。
		if strings.EqualFold(template.Name, "name") {
			continue
		}

		if _, exists := seen[template.Name]; exists {
			continue
		}

		seen[template.Name] = struct{}{}
		templates = append(templates, template)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan command output: %w", err)
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf(
			"no templates found in Wails output: %q",
			compactOutput(output),
		)
	}

	return templates, nil
}

// parseTemplateTableRow 解析一行模板表格。
// 返回值 ok=false 表示该行不是有效模板记录。
func parseTemplateTableRow(line string) (template WailsTemplate, ok bool) {
	line = strings.TrimSpace(line)

	if line == "" {
		return WailsTemplate{}, false
	}

	// 有效数据行应该以竖线开头和结尾。
	if !strings.HasPrefix(line, "|") ||
		!strings.HasSuffix(line, "|") {
		return WailsTemplate{}, false
	}

	// 移除首尾竖线：
	//
	//	| react | React + TypeScript + Vite |
	//
	// 变为：
	//
	//	react | React + TypeScript + Vite
	content := strings.TrimSpace(
		strings.TrimSuffix(
			strings.TrimPrefix(line, "|"),
			"|",
		),
	)

	// 只分割第一次出现的竖线，防止描述中未来包含竖线。
	columns := strings.SplitN(content, "|", 2)
	if len(columns) != 2 {
		return WailsTemplate{}, false
	}

	name := strings.TrimSpace(columns[0])
	description := strings.TrimSpace(columns[1])

	if name == "" || description == "" {
		return WailsTemplate{}, false
	}

	// 排除表格分隔符。
	if isTableSeparator(name) || isTableSeparator(description) {
		return WailsTemplate{}, false
	}

	return WailsTemplate{
		Name:        name,
		Description: description,
	}, true
}

// normaliseTableLine 将 Unicode 表格字符转换成普通 ASCII 竖线。
func normaliseTableLine(line string) string {
	replacer := strings.NewReplacer(
		"│", "|",
		"┃", "|",
		"║", "|",
	)

	return strings.TrimSpace(replacer.Replace(line))
}

// isTableSeparator 判断字段是否只是表格横线字符。
func isTableSeparator(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}

	for _, char := range value {
		switch char {
		case '-', '─', '━', '═', '┄', '┈':
			continue
		default:
			return false
		}
	}

	return true
}

// stripANSI 移除终端颜色和控制字符。
func stripANSI(value string) string {
	return ansiEscapePattern.ReplaceAllString(value, "")
}

// compactOutput 压缩命令输出，避免错误消息过长。
func compactOutput(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Join(strings.Fields(value), " ")

	const maxLength = 500
	if len(value) <= maxLength {
		return value
	}

	return value[:maxLength] + "..."
}
