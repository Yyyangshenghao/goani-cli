# 使用指南

## 目录

- [命令概览](#命令概览)
- [配置播放器](#配置播放器)
- [搜索动漫](#搜索动漫)
- [播放动漫](#播放动漫)
- [媒体源管理](#媒体源管理)
- [更新程序](#更新程序)
- [其他命令](#其他命令)
- [配置文件](#配置文件)

---

## 命令概览

```bash
goani <command> [arguments]
```

| 命令 | 说明 |
|------|------|
| `search` | 搜索动漫 |
| `play` | 搜索并播放动漫 |
| `source` | 管理媒体源订阅 |
| `config` | 配置播放器 |
| `update` | 更新到最新版本 |
| `version` | 显示版本信息 |

---

## 配置播放器

首次使用建议先配置播放器。现在播放器配置会统一写入 `config.json`，支持保存多个播放器路径，并指定默认播放器。

### Windows

```powershell
# mpv
goani config player mpv "D:\MPV播放器\mpv.exe"

# VLC
goani config player vlc "C:\Program Files\VideoLAN\VLC\vlc.exe"

# PotPlayer
goani config player potplayer "C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe"

# 设置默认播放器
goani config player default mpv
```

### macOS

```bash
# IINA（推荐）
goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"

# mpv
goani config player mpv "/usr/local/bin/mpv"

# VLC
goani config player vlc "/Applications/VLC.app/Contents/MacOS/VLC"

# 设置默认播放器
goani config player default iina
```

### Linux

```bash
# mpv
goani config player mpv "/usr/bin/mpv"

# VLC
goani config player vlc "/usr/bin/vlc"

# 设置默认播放器
goani config player default mpv
```

说明：
- `goani config player <name> <path>` 会保存该播放器路径，并把它设为默认播放器
- `goani config player default <name>` 只切换默认播放器，不改路径

---

## 搜索动漫

```bash
goani search <关键词>
```

示例：
```bash
goani search 葬送的芙莉莲
goani search 进击的巨人
```

搜索后会显示结果列表，选择后可查看剧集并播放。

### 实时搜索（TUI）

```bash
goani search --interactive [关键词]
goani search -i [关键词]
```

示例：

```bash
goani search --interactive 葬送的芙莉莲
goani search -i
```

说明：
- 支持交互式终端时，会进入实时搜索 TUI
- 不支持 TUI 的终端会自动回退到普通搜索模式
- 在 TUI 中可直接输入关键词，支持上下选择片源并回车确认

---

## 播放动漫

```bash
goani play <关键词>
```

示例：
```bash
goani play 葬送的芙莉莲
```

这是推荐的用法，搜索、选集、播放一气呵成。

---

## 媒体源管理

### 列出媒体源和订阅

```bash
goani source list
```

### 订阅新的媒体源

```bash
goani source sub <url> [名称]
```

示例：
```bash
goani source sub https://example.com/sources.json 我的订阅
```

### 取消订阅

```bash
goani source unsub <url>
```

### 刷新订阅

从所有订阅地址重新获取最新的媒体源：

```bash
goani source refresh
```

### 重置为默认源

```bash
goani source reset
```

---

## 更新程序

自动检查并更新到最新版本：

```bash
goani update
```

---

## 其他命令

### 查看版本

```bash
goani version
```

---

## 配置文件

配置文件位置：
- Windows: `%USERPROFILE%\.goani\config.json`
- macOS/Linux: `~/.goani/config.json`

`config.json` 里会保存播放器路径、默认播放器和片源订阅地址。

媒体源缓存：
- Windows: `%USERPROFILE%\.goani\sources_cache.json`
- macOS/Linux: `~/.goani/sources_cache.json`

`sources_cache.json` 只保存拉取下来的片源缓存数据。

### 更换播放器

重新运行 config 命令即可覆盖之前的配置：

```bash
goani config player vlc "/path/to/vlc"
```

### 配置示例

```json
{
  "player": {
    "default": "mpv",
    "paths": {
      "mpv": "D:\\MPV播放器\\mpv.exe",
      "vlc": "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
    }
  },
  "sources": {
    "subscriptions": [
      {
        "url": "https://sub.creamycake.org/v1/css1.json",
        "name": "默认源"
      }
    ]
  }
}
```
