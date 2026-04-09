<p align=center>
<br>
<a href="http://makeapullrequest.com"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg"></a>
<img src="https://img.shields.io/badge/language-Go-blue">
<img src="https://img.shields.io/badge/os-windows-yellowgreen">
<img src="https://img.shields.io/badge/os-linux-brightgreen">
<img src="https://img.shields.io/badge/os-mac-brightgreen">
<br>
<h1 align="center">goani-cli</h1>
</p>

<h3 align="center">命令行动漫播放器，专为中文动漫站点设计</h3>

## 简介

`goani-cli` 是一个用 Go 语言编写的命令行动漫播放器。

### 特性

- 🚀 程序本体为单二进制文件
- 🌐 支持 38+ 中文动漫源
- 🎮 交互式命令行界面
- 📺 支持多种播放器（mpv、vlc、potplayer、iina）

### 致谢

本项目灵感来源于：
- [pystardust/ani-cli](https://github.com/pystardust/ani-cli) - 原始 CLI 交互模型
- [MajoSissi/animeko-source](https://github.com/MajoSissi/animeko-source) - 中文站点规则参考

---

## 快速开始

首次播放前，请先确保系统已安装 `mpv`、`vlc`、`potplayer` 或 `iina` 之一；如果程序未自动识别到播放器，可按[使用指南](docs/usage.md#配置播放器)手动配置路径。

```bash
# 配置播放器（会同时设为默认播放器）
goani config player mpv "/path/to/mpv"

# 搜索动漫
goani search 葬送的芙莉莲

# 搜索并播放
goani play 葬送的芙莉莲
```

---

## 文档

| 文档 | 说明 |
|------|------|
| [安装指南](docs/installation.md) | Windows / macOS / Linux 安装方法 |
| [使用指南](docs/usage.md) | 命令详解与播放器配置 |
| [常见问题](docs/faq.md) | FAQ 与故障排除 |
| [开发指南](docs/development.md) | 构建、测试与贡献 |

---

## 支持的播放器

| 播放器 | Windows | Linux | macOS |
|--------|---------|-------|-------|
| mpv | ✅ | ✅ | ✅ |
| VLC | ✅ | ✅ | ✅ |
| PotPlayer | ✅ | ❌ | ❌ |
| IINA | ❌ | ❌ | ✅ |

---

## License

[GPL v3](LICENSE)
