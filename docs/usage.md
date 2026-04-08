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

首次使用需要配置播放器路径。

### Windows

```powershell
# mpv
goani config player mpv "D:\MPV播放器\mpv.exe"

# VLC
goani config player vlc "C:\Program Files\VideoLAN\VLC\vlc.exe"

# PotPlayer
goani config player potplayer "C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe"
```

### macOS

```bash
# IINA（推荐）
goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"

# mpv
goani config player mpv "/usr/local/bin/mpv"

# VLC
goani config player vlc "/Applications/VLC.app/Contents/MacOS/VLC"
```

### Linux

```bash
# mpv
goani config player mpv "/usr/bin/mpv"

# VLC
goani config player vlc "/usr/bin/vlc"
```

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

媒体源缓存：
- Windows: `%USERPROFILE%\.goani\sources_cache.json`
- macOS/Linux: `~/.goani/sources_cache.json`

### 更换播放器

重新运行 config 命令即可覆盖之前的配置：

```bash
goani config player vlc "/path/to/vlc"
```
