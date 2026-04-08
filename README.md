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
- 🌐 支持多个中文动漫源
- 🔍 并行搜索多个源
- 📺 支持多种播放器

### 致谢

本项目灵感来源于：
- [pystardust/ani-cli](https://github.com/pystardust/ani-cli) - 原始 CLI 交互模型
- [MajoSissi/animeko-source](https://github.com/MajoSissi/animeko-source) - 中文站点规则参考

## 安装

```bash
go install github.com/yshscpu/goani-cli/cmd/kanfan@latest
```

或从 [Releases](https://github.com/yshscpu/goani-cli/releases) 下载预编译二进制文件。

## 使用

```bash
kanfan search "葬送的芙莉莲"
kanfan play "葬送的芙莉莲" --episode 1
```

## 开发

### 构建

```bash
git clone https://github.com/yshscpu/goani-cli.git
cd goani-cli
go build -o kanfan ./cmd/kanfan
```

### 项目结构

```
goani-cli/
├── cmd/
│   └── kanfan/           # 入口
├── internal/
│   └── source/           # 媒体源模块
│       ├── source.go     # 接口定义
│       ├── config.go     # 配置结构体
│       ├── config_loader.go
│       └── web_selector.go
├── mediaSourceJson/      # 订阅源配置
├── go.mod
└── go.sum
```

## License

MIT
