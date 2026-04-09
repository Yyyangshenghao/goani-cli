# 安装指南

## 先准备什么

`goani` 是终端工具，安装完成后通常还需要一个可用播放器。

- Windows: `mpv`、`VLC`、`PotPlayer`
- macOS: `mpv`、`VLC`、`IINA`
- Linux: `mpv`、`VLC`

如果你不确定先从哪里开始，建议先安装 `mpv`，再运行 `goani tui`。

## Windows

### 方式一：下载二进制，最省事

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载 `goani-windows-amd64.zip`
2. 解压得到 `goani.exe`
3. 把 `goani.exe` 放到一个固定目录，例如 `C:\Tools\goani\`
4. 把该目录加入 `PATH`
5. 重新打开终端，执行：

```powershell
goani version
```

### 方式二：使用 Go 安装

需要先安装 Go 1.22+：

```powershell
go install github.com/Yyyangshenghao/goani-cli/cmd/goani@latest
```

## macOS

### 方式一：下载二进制

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载对应架构的压缩包
2. 解压后把 `goani` 放到 `PATH` 中，例如：

```bash
tar -xzf goani-darwin-arm64.tar.gz
sudo mv goani /usr/local/bin/
chmod +x /usr/local/bin/goani
```

3. 验证安装：

```bash
goani version
```

### 方式二：使用 Go 安装

```bash
go install github.com/Yyyangshenghao/goani-cli/cmd/goani@latest
```

## Linux

### 方式一：下载二进制

1. 从 [Releases](https://github.com/Yyyangshenghao/goani-cli/releases) 下载对应架构的压缩包
2. 解压后把 `goani` 放到 `PATH` 中，例如：

```bash
tar -xzf goani-linux-amd64.tar.gz
sudo mv goani /usr/local/bin/
chmod +x /usr/local/bin/goani
```

3. 验证安装：

```bash
goani version
```

### 方式二：使用 Go 安装

```bash
go install github.com/Yyyangshenghao/goani-cli/cmd/goani@latest
```

## 安装后做什么

1. 先运行 `goani version` 确认安装成功。
2. 配置一个播放器路径。
3. 直接进入 `goani tui`，或使用 `goani search` / `goani play`。

如果你想先做最短路径，推荐：

```powershell
goani config player mpv "D:\Tools\mpv\mpv.exe"
goani tui
```
