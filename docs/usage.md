# 使用指南

## 先看这一段

- `goani` 默认显示帮助，不会直接进入 TUI。
- `goani tui` 是推荐的交互入口。
- `goani search` 和 `goani play` 继续保留，适合熟悉 CLI 的用户。
- 所有配置最终都写进 `config.json`。

---

## 命令概览

```bash
goani <command> [arguments]
```

| 命令 | 说明 |
|------|------|
| `tui` | 进入主 TUI |
| `search` | 搜索动漫 |
| `play` | 搜索并播放动漫 |
| `source` | 管理媒体源订阅 |
| `config` | 管理播放器和配置文件 |
| `update` | 更新程序 |
| `version` | 查看版本信息 |

如果你不确定从哪里开始，先运行 `goani tui`。

---

## TUI 流程

`goani tui` 会先进入主界面，里面可以继续进入搜索、媒体源、配置和版本信息。

当前更常用的搜索流程是：

1. 从首页进入搜索页并输入关键词。
2. 从实时搜索结果里选择一部聚合后的番剧，而不是先固定某一个片源。
3. 进入聚合选集页，按“集”查看多个片源合并后的剧集。
4. 进入线路页，查看每条线路对应的片源、解析出的直链、格式和清晰度。
5. 选择一条线路并启动播放器。
6. 播放成功后进入播放页，查看当前播放器和线路，并决定返回到哪一层；如果当前番剧还有相邻剧集，也可以直接跳到上一集或下一集。

配置页和媒体源页不在这条搜索链路里，它们是从首页单独进入的。

常用操作：

- `Esc` 返回上一页
- `Ctrl+C` 退出当前界面
- 番剧列表支持本地过滤，不会重新请求网络
- 选集页支持 `r` 切换顺序和倒序
- 选集页支持直接输入数字跳转到对应集数
- 线路解析页面会先显示加载页，整体等待时间最多 5 秒
- 播放页在有相邻剧集时支持“上一集 / 下一集”

---

## 配置播放器

播放器配置统一保存在 `config.json`。

建议先通过 TUI 配置：

1. 进入 `goani tui`
2. 打开 `配置`
3. 选择 `播放器`
4. 输入播放器路径并保存
5. 需要时再把某个播放器设为默认播放器

如果你更喜欢命令行，也可以直接使用：

### Windows

```powershell
goani config player mpv "D:\MPV播放器\mpv.exe"
goani config player vlc "C:\Program Files\VideoLAN\VLC\vlc.exe"
goani config player potplayer "C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe"
goani config player default mpv
```

### macOS

```bash
goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"
goani config player mpv "/usr/local/bin/mpv"
goani config player vlc "/Applications/VLC.app/Contents/MacOS/VLC"
goani config player default iina
```

### Linux

```bash
goani config player mpv "/usr/bin/mpv"
goani config player vlc "/usr/bin/vlc"
goani config player default mpv
```

说明：

- `goani config player <name> <path>` 只保存播放器路径，不会自动修改默认播放器。
- `goani config player default <name>` 只设置默认播放器，不会改路径。
- 如果默认播放器为空，播放时会先从已配置路径里挑一个可用播放器，并把它写回默认播放器。

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

经典搜索流程会列出片源，然后进入番剧列表、选集和播放。

### 实时搜索

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

- 这是兼容入口，功能上和 TUI 搜索相连。
- 会先把多个片源的搜索结果尽量聚合成一份番剧列表。
- 进入番剧后，会按“集”归并多个片源的候选线路。
- 在聚合番剧列表里支持本地二次过滤，方便再次缩小范围。
- 在选集页可以按顺序或倒序浏览。
- 线路解析会显示加载页，不会长时间卡在原界面。
- 线路页会标明每条线路来自哪个片源。

---

## 播放动漫

```bash
goani play <关键词>
```

示例：

```bash
goani play 葬送的芙莉莲
```

`play` 适合想快速从搜索走到播放的场景。它会复用搜索、选集和线路 fallback 逻辑。

对于 `PotPlayer + m3u8`，`goani` 可能会先启动本地 HLS 代理，再把本地地址交给播放器，这是兼容行为。

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

```bash
goani source refresh
```

### 重置为默认源

```bash
goani source reset
```

---

## 更新程序

```bash
goani update
```

---

## 查看版本

```bash
goani version
```

---

## 配置文件

配置文件位置：

- Windows: `%USERPROFILE%\.goani\config.json`
- macOS/Linux: `~/.goani/config.json`

`config.json` 里保存三类内容：

- 播放器路径
- 默认播放器
- 片源订阅

同目录下还有一个缓存文件：

- Windows: `%USERPROFILE%\.goani\sources_cache.json`
- macOS/Linux: `~/.goani/sources_cache.json`

这个缓存文件只存运行时拉取到的媒体源缓存，不是主配置。

### 直接打开配置文件

在 TUI 里进入 `配置`，再选择 `打开 config.json`，可以直接用系统默认编辑器打开配置文件。

如果你习惯直接编辑 JSON，也可以手动修改后再运行 `goani source refresh` 或重新启动程序。

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
