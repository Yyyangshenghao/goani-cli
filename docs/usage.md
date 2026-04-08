# 使用指南

## 命令概览

```bash
goani <command> [arguments]
```

| 命令 | 说明 |
|------|------|
| `search` | 搜索动漫 |
| `play` | 搜索并播放动漫 |
| `list` | 列出所有媒体源 |
| `config` | 配置播放器 |
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

## 其他命令

### 列出媒体源

```bash
goani list
```

### 查看版本

```bash
goani version
```

---

## 配置文件

配置文件位置：
- Windows: `%USERPROFILE%\.goani\config.json`
- macOS/Linux: `~/.goani/config.json`

### 更换播放器

重新运行 config 命令即可覆盖之前的配置：

```bash
goani config player vlc "/path/to/vlc"
```
