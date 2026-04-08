package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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

	ui.Success("更新完成！")
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

	var archiveName string
	switch goos {
	case "windows":
		archiveName = fmt.Sprintf("goani-%s-%s.zip", goos, arch)
	case "darwin", "linux":
		archiveName = fmt.Sprintf("goani-%s-%s.tar.gz", goos, arch)
	default:
		return fmt.Errorf("不支持的系统: %s", goos)
	}

	downloadURL := fmt.Sprintf("https://github.com/Yyyangshenghao/goani-cli/releases/download/%s/%s", ver, archiveName)
	ui.Info("下载地址: %s", downloadURL)

	// 下载文件
	tmpFile := archiveName
	if err := c.downloadFile(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer os.Remove(tmpFile)

	// 解压
	ui.Info("正在解压...")
	var cmd *exec.Cmd
	if goos == "windows" {
		// Windows 使用 PowerShell 解压
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Expand-Archive -Path %s -DestinationPath . -Force", tmpFile))
	} else {
		cmd = exec.Command("tar", "-xzf", tmpFile)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 替换可执行文件
	binaryName := "goani"
	if goos == "windows" {
		binaryName = "goani.exe"
	}

	// 备份旧版本
	backupPath := execPath + ".bak"
	os.Rename(execPath, backupPath)

	// 移动新版本
	if err := os.Rename(binaryName, execPath); err != nil {
		// 恢复备份
		os.Rename(backupPath, execPath)
		return fmt.Errorf("替换可执行文件失败: %w", err)
	}

	// 删除备份
	os.Remove(backupPath)

	return nil
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
