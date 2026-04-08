<p align=center>
<br>
<a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg"></a>
<img src="https://img.shields.io/badge/language-Go-blue">
<img src="https://img.shields.io/badge/os-windows-yellowgreen">
<img src="https://img.shields.io/badge/os-linux-brightgreen">
<img src="https://img.shields.io/badge/os-mac-brightgreen">
<br>
<h1 align="center">
goani-cli
</h1>
</p>

<h3 align="center">
A command-line anime player for Chinese web sources, written in Go.
</h3>

## 简介

`goani-cli` 是一个用 Go 语言编写的命令行动漫播放器，专为中文动漫站点设计。

### 特性

- 🚀 单二进制文件，零依赖
- 🌐 支持 38+ 中文动漫源
- 🎮 交互式命令行界面
- 📺 支持多种播放器（mpv、vlc、potplayer、iina）

### 致谢

本项目灵感来源于：
- [pystardust/ani-cli](https://github.com/pystardust/ani-cli) - 原始 CLI 交互模型
- [MajoSissi/animeko-source](https://github.com/MajoSissi/animeko-source) - 中文站点规则参考

---

## 安装

### Windows

#### 方式一：下载二进制（推荐）

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载 `goani-windows-amd64.zip`

2. 解压得到 `goani.exe`

3. 将 `goani.exe` 移动到一个固定目录，例如 `C:\Tools\goani\`

4. **添加到环境变量 PATH**：
   
   **方法 A：通过系统设置（推荐）**
   - 右键「此电脑」→「属性」→「高级系统设置」
   - 点击「环境变量」
   - 在「系统变量」中找到 `Path`，点击「编辑」
   - 点击「新建」，添加 `goani.exe` 所在目录，例如 `C:\Tools\goani`
   - 点击「确定」保存
   
   **方法 B：通过命令行（需要管理员权限）**
   ```powershell
   # 假设 goani.exe 在 C:\Tools\goani 目录
   setx PATH "%PATH%;C:\Tools\goani"
   ```
   
   **方法 C：PowerShell（推荐）**
   ```powershell
   # 添加到用户 PATH
   $env:Path += ";C:\Tools\goani"
   [Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::User)
   ```

5. **重启终端**（重要！），然后验证：
   ```powershell
   goani version
   ```

#### 方式二：Go install

需要先安装 Go 1.22+：

```powershell
go install github.com/yshscpu/goani-cli/cmd/goani@latest
```

---

### macOS

#### 方式一：下载二进制（推荐）

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载对应版本：
   - Intel Mac：`goani-darwin-amd64.tar.gz`
   - Apple Silicon (M1/M2)：`goani-darwin-arm64.tar.gz`

2. 解压并移动到 PATH 目录：
   ```bash
   # 解压
   tar -xzf goani-darwin-arm64.tar.gz
   
   # 移动到 /usr/local/bin
   sudo mv goani /usr/local/bin/
   
   # 添加执行权限
   chmod +x /usr/local/bin/goani
   ```

3. 验证：
   ```bash
   goani version
   ```

#### 方式二：Go install

```bash
go install github.com/yshscpu/goani-cli/cmd/goani@latest
```

---

### Linux

#### 方式一：下载二进制（推荐）

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载对应版本：
   - x86_64：`goani-linux-amd64.tar.gz`
   - ARM64：`goani-linux-arm64.tar.gz`

2. 解压并安装：
   ```bash
   # 解压
   tar -xzf goani-linux-amd64.tar.gz
   
   # 移动到 /usr/local/bin
   sudo mv goani /usr/local/bin/
   
   # 添加执行权限
   chmod +x /usr/local/bin/goani
   ```

3. 验证：
   ```bash
   goani version
   ```

#### 方式二：Go install

```bash
go install github.com/yshscpu/goani-cli/cmd/goani@latest
```

---

## 快速开始

### 第一步：配置播放器

首次使用需要配置播放器路径：

**Windows：**
```powershell
# mpv
goani config player mpv "D:\MPV播放器\mpv.exe"

# VLC
goani config player vlc "C:\Program Files\VideoLAN\VLC\vlc.exe"

# PotPlayer
goani config player potplayer "C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe"
```

**macOS：**
```bash
# IINA（推荐）
goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"

# mpv
goani config player mpv "/usr/local/bin/mpv"

# VLC
goani config player vlc "/Applications/VLC.app/Contents/MacOS/VLC"
```

**Linux：**
```bash
# mpv
goani config player mpv "/usr/bin/mpv"

# VLC
goani config player vlc "/usr/bin/vlc"
```

### 第二步：搜索并播放

```bash
# 搜索动漫
goani search 葬送的芙莉莲

# 搜索并直接播放（推荐）
goani play 葬送的芙莉莲
```

### 其他命令

```bash
goani list       # 列出所有媒体源
goani version    # 显示版本
```

---

## 支持的播放器

| 播放器 | Windows | Linux | macOS | 下载地址 |
|--------|---------|-------|-------|----------|
| mpv | ✅ | ✅ | ✅ | https://mpv.io/installation/ |
| VLC | ✅ | ✅ | ✅ | https://www.videolan.org/vlc/ |
| PotPlayer | ✅ | ❌ | ❌ | https://potplayer.daum.net/ |
| IINA | ❌ | ❌ | ✅ | https://iina.io/ |

---

## 常见问题

### Q: 提示「未找到可用播放器」

需要先配置播放器路径：
```bash
goani config player mpv "/path/to/mpv"
```

### Q: Windows 下提示「goani 不是内部或外部命令」

1. 确认 `goani.exe` 所在目录已添加到 PATH
2. 重启终端或命令提示符
3. 如果是 PowerShell，可能需要重启

### Q: 如何查看当前配置？

配置文件位置：
- Windows: `%USERPROFILE%\.goani\config.json`
- macOS/Linux: `~/.goani/config.json`

### Q: 如何更换播放器？

重新运行 config 命令即可：
```bash
goani config player vlc "/path/to/vlc"
```

---

## 开发

### 构建

```bash
git clone https://github.com/yshscpu/goani-cli.git
cd goani-cli

# 构建当前平台
go build -o goani ./cmd/goani

# 或使用 Makefile
make build
```

### 运行测试

```bash
go run test/source/main.go    # 核心功能测试
go run test/player/main.go    # 播放器测试
```

---

## License

MIT
