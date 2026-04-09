# 常见问题

## Q: 提示「未找到可用播放器」

需要先配置播放器路径：

```bash
goani config player mpv "/path/to/mpv"
```

详见 [使用指南](usage.md#配置播放器)。

---

## Q: Windows 下提示「goani 不是内部或外部命令」

1. 确认 `goani.exe` 所在目录已添加到 PATH
2. **重启终端**（重要！）
3. 如果是 PowerShell，可能需要完全关闭后重新打开

详见 [安装指南 - Windows](installation.md#windows)。

---

## Q: 搜索不到结果

可能原因：
1. 关键词太短或不准确，尝试使用更精确的关键词
2. 网络问题，检查网络连接
3. 当前媒体源暂无该动漫，可尝试其他关键词

---

## Q: 播放失败或视频无法加载

1. 确认播放器路径配置正确
2. 确认播放器已安装且可正常运行
3. 尝试更换其他播放器

---

## Q: 如何查看当前配置？

查看配置文件：
- Windows: `%USERPROFILE%\.goani\config.json`
- macOS/Linux: `~/.goani/config.json`

---

## Q: 如何重置配置？

删除配置文件即可，程序会自动创建默认配置：

```bash
# macOS / Linux
rm ~/.goani/config.json

# Windows PowerShell
Remove-Item "$env:USERPROFILE\.goani\config.json"
```

---

## Q: 支持哪些动漫源？

运行 `goani source list` 可查看当前支持的所有媒体源和订阅列表。

---

## 其他问题

如有其他问题，欢迎在 [GitHub Issues](https://github.com/Yyyangshenghao/goani-cli/issues) 提问。
