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

## 安装

```bash
go install github.com/yshscpu/goani-cli/cmd/goani@latest
```

或从 [Releases](https://github.com/yshscpu/goani-cli/releases) 下载预编译二进制文件。

## 使用

### 配置播放器

首次使用需要配置播放器路径：

```bash
# Windows
goani config player mpv "D:\MPV播放器\mpv.exe"
goani config player vlc "C:\Program Files\VideoLAN\VLC\vlc.exe"
goani config player potplayer "C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe"

# macOS
goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"
goani config player mpv "/usr/local/bin/mpv"

# Linux
goani config player mpv "/usr/bin/mpv"
goani config player vlc "/usr/bin/vlc"
```

### 搜索动漫

```bash
goani search 葬送的芙莉莲
```

### 播放动漫

```bash
goani play 葬送的芙莉莲
```

### 其他命令

```bash
goani list       # 列出所有媒体源
goani version    # 显示版本
```

## 支持的播放器

| 播放器 | Windows | Linux | macOS |
|--------|---------|-------|-------|
| mpv | ✅ | ✅ | ✅ |
| VLC | ✅ | ✅ | ✅ |
| PotPlayer | ✅ | ❌ | ❌ |
| IINA | ❌ | ❌ | ✅ |

## 开发

### 构建

```bash
git clone https://github.com/yshscpu/goani-cli.git
cd goani-cli
go build -o goani ./cmd/goani
```

### 运行测试

```bash
go run test/source/main.go    # 核心功能测试
go run test/player/main.go    # 播放器测试
```

### 项目结构

```
goani-cli/
├── cmd/
│   └── goani/
│       └── main.go              # 入口
├── internal/
│   ├── app/
│   │   └── app.go               # 应用核心
│   ├── cli/
│   │   ├── root.go              # CLI 入口
│   │   └── commands/
│   │       └── commands.go      # 命令实现
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── player/
│   │   ├── player.go            # 播放器接口
│   │   ├── manager.go           # 播放器管理器
│   │   ├── mpv.go
│   │   ├── vlc.go
│   │   ├── potplayer.go
│   │   └── iina.go
│   ├── source/
│   │   ├── source.go            # 媒体源接口
│   │   ├── config.go            # 配置结构体
│   │   ├── config_loader.go
│   │   └── webselector/
│   │       ├── webselector.go   # 爬虫实现
│   │       ├── search.go
│   │       ├── episode.go
│   │       ├── video.go
│   │       ├── fetch.go
│   │       └── url.go
│   └── ui/
│       ├── ui.go                # UI 交互
│       └── print.go             # 打印工具
├── test/
│   ├── source/main.go           # 核心功能测试
│   └── player/main.go           # 播放器测试
├── mediaSourceJson/
│   └── css1.json                # 订阅源配置
├── go.mod
└── go.sum
```

## License

MIT
