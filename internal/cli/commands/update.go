package commands

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/ui"
	"github.com/Yyyangshenghao/goani-cli/internal/version"
)

func init() {
	Register(&UpdateCommand{})
}

// UpdateCommand 更新命令
type UpdateCommand struct{}

// Name 返回命令名称
func (c *UpdateCommand) Name() string {
	return "update"
}

// ShortDesc 返回简短描述
func (c *UpdateCommand) ShortDesc() string {
	return "更新到最新版本"
}

// Run 执行命令
func (c *UpdateCommand) Run(args []string) {
	ui.Info("当前版本: %s", version.Short())
	ui.Info("正在检查更新...")

	// 获取最新版本
	latestVersion, err := c.getLatestVersion()
	if err != nil {
		ui.Error("检查更新失败: %v", err)
		os.Exit(1)
	}

	ui.Info("最新版本: %s", latestVersion)

	if version.Short() == latestVersion {
		ui.Success("已是最新版本")
		return
	}

	// 下载并安装
	if err := c.install(latestVersion); err != nil {
		ui.Error("更新失败: %v", err)
		os.Exit(1)
	}

	if runtime.GOOS == "windows" {
		ui.Success("更新已启动，程序退出后会完成替换")
	} else {
		ui.Success("更新完成！")
	}
	ui.Info("新版本: %s", latestVersion)
}

// getLatestVersion 获取最新版本号
func (c *UpdateCommand) getLatestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Yyyangshenghao/goani-cli/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// install 下载并安装新版本
func (c *UpdateCommand) install(ver string) error {
	// 确定下载URL
	goos := runtime.GOOS
	arch := runtime.GOARCH

	var (
		archiveName string
		binaryName  string
	)
	switch goos {
	case "windows":
		archiveName = fmt.Sprintf("goani-%s-%s.zip", goos, arch)
		binaryName = "goani.exe"
	case "darwin", "linux":
		archiveName = fmt.Sprintf("goani-%s-%s.tar.gz", goos, arch)
		binaryName = "goani"
	default:
		return fmt.Errorf("不支持的系统: %s", goos)
	}

	downloadURL := fmt.Sprintf("https://github.com/Yyyangshenghao/goani-cli/releases/download/%s/%s", ver, archiveName)
	ui.Info("下载地址: %s", downloadURL)

	tempDir, err := os.MkdirTemp("", "goani-update-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	if goos != "windows" {
		defer os.RemoveAll(tempDir)
	}

	// 下载文件
	tmpFile := filepath.Join(tempDir, archiveName)
	if err := c.downloadFile(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 解压
	ui.Info("正在解压...")
	if goos == "windows" {
		return c.installWindows(tmpFile, tempDir, binaryName)
	}

	cmd := exec.Command("tar", "-xzf", tmpFile, "-C", tempDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 备份旧版本
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("备份当前可执行文件失败: %w", err)
	}

	newBinaryPath := filepath.Join(tempDir, binaryName)
	if err := os.Rename(newBinaryPath, execPath); err != nil {
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("替换可执行文件失败: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}

func (c *UpdateCommand) installWindows(zipPath, tempDir, binaryName string) error {
	extractDir := filepath.Join(tempDir, "extract")
	if err := unzip(zipPath, extractDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	newBinaryPath := filepath.Join(extractDir, binaryName)
	if _, err := os.Stat(newBinaryPath); err != nil {
		return fmt.Errorf("未找到解压后的可执行文件: %w", err)
	}

	scriptPath := filepath.Join(tempDir, "apply-update.ps1")
	script := buildWindowsUpdateScript(execPath, newBinaryPath)
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("创建更新脚本失败: %w", err)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新脚本失败: %w", err)
	}

	return nil
}

func unzip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath := filepath.Join(dest, file.Name)
		cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
		cleanTarget := filepath.Clean(targetPath)
		if !strings.HasPrefix(cleanTarget, cleanDest) && cleanTarget != filepath.Clean(dest) {
			return fmt.Errorf("非法压缩路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		in, err := file.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			in.Close()
			return err
		}

		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			in.Close()
			return err
		}

		out.Close()
		in.Close()
	}

	return nil
}

func buildWindowsUpdateScript(execPath, newBinaryPath string) string {
	target := escapePowerShellLiteral(execPath)
	source := escapePowerShellLiteral(newBinaryPath)
	backup := escapePowerShellLiteral(execPath + ".bak")

	return fmt.Sprintf(`Start-Sleep -Milliseconds 750
for ($i = 0; $i -lt 20; $i++) {
  try {
    if (Test-Path -LiteralPath '%s') {
      Remove-Item -LiteralPath '%s' -Force -ErrorAction SilentlyContinue
    }
    Move-Item -LiteralPath '%s' -Destination '%s' -Force
    Move-Item -LiteralPath '%s' -Destination '%s' -Force
    Remove-Item -LiteralPath '%s' -Force -ErrorAction SilentlyContinue
    exit 0
  } catch {
    Start-Sleep -Milliseconds 500
  }
}
exit 1
`, backup, backup, target, backup, source, target, backup)
}

func escapePowerShellLiteral(path string) string {
	return strings.ReplaceAll(path, "'", "''")
}

// downloadFile 下载文件
func (c *UpdateCommand) downloadFile(url, filepath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Usage 返回使用说明
func (c *UpdateCommand) Usage() string {
	return "用法: goani update\n\n自动下载并安装最新版本"
}
