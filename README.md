# goani-cli

`goani-cli` 是一个面向中文动漫站点的终端工具。它保留传统 CLI，方便脚本和快速调用，也提供 `goani tui` 作为更完整的交互入口。

## 这项目能做什么

- 搜索动漫
- 进入 TUI 完成搜索、番剧筛选、选集、线路选择和播放
- 管理播放器配置、片源订阅和片源渠道开关
- 通过 `config.json` 和 `source_preferences.json` 保存配置
- 为各类播放器提供统一的 `m3u8` 本地 HLS 兼容层

## 推荐开始方式

第一次使用时，建议先完成播放器配置，然后直接进入 TUI：

```powershell
goani config player mpv "D:\Tools\mpv\mpv.exe"
goani tui
```

如果你更习惯命令行，也可以继续使用经典 CLI：

```powershell
goani search 葬送的芙莉莲
goani play 葬送的芙莉莲
goani source list
goani source doctor
```

## 文档

- [安装指南](docs/installation.md)
- [使用指南](docs/usage.md)
- [常见问题](docs/faq.md)
- [开发指南](docs/dev/development.md)

## 支持的播放器

| 播放器 | Windows | Linux | macOS |
|--------|---------|-------|-------|
| mpv | 支持 | 支持 | 支持 |
| VLC | 支持 | 支持 | 支持 |
| PotPlayer | 支持 | 不支持 | 不支持 |
| IINA | 不支持 | 不支持 | 支持 |

播放 `m3u8` 时，`goani` 会优先走自己的本地 HLS 兼容层；如果线路需要额外请求头或子 playlist 改写，这属于正常兼容行为。兼容层还会做短时分片缓存和轻量预取，用来减少小幅网络抖动带来的卡顿。

## 致谢

本项目受以下项目启发：

- [pystardust/ani-cli](https://github.com/pystardust/ani-cli)
- [MajoSissi/animeko-source](https://github.com/MajoSissi/animeko-source)

## License

[GPL v3](LICENSE)
