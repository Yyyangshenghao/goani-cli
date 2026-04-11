# 开发指南

这份文档面向维护者，记录当前项目结构、构建方式、测试方式和一些约定。

它不是用户说明书。面向用户的内容请看 [usage.md](../usage.md)。

如果想看后续待办和演进方向，可以继续看 [todo.md](./todo.md)。

## 文档分工

- `README.md`、`docs/installation.md`、`docs/usage.md`、`docs/faq.md` 记录当前对用户生效的行为和用法。
- `docs/dev/development.md`、`docs/dev/command.md` 记录当前维护者需要了解的结构、命令和约定。
- `docs/dev/todo.md` 只作为 backlog 使用，不应该当作“当前已经实现的功能清单”。
- 某项功能已经落地后，应尽快把说明同步到现状文档里，并从 `todo.md` 里移除或改写成下一步增强项。

---

## 项目结构

```text
goani-cli/
├── cmd/
│   ├── goani/          # 主程序入口
│   └── goani-debug-*/  # 手动调试工具
├── internal/
│   ├── app/            # 应用编排
│   ├── cli/            # 命令入口与注册
│   │   └── commands/   # 具体命令实现
│   ├── player/         # 播放器、配置和 HLS 代理
│   ├── source/         # 媒体源模型、订阅、归类和抓取
│   │   └── webselector/ # 站点解析实现
│   ├── ui/
│   │   ├── console/    # 传统终端交互
│   │   └── tui/        # Bubble Tea TUI 页面
│   └── version/        # 版本信息
├── internal/workflow/  # 搜索、播放、配置等跨层流程
└── docs/               # 用户文档和开发文档
```

当前比较重要的拆分点是：

- `internal/ui/console` 负责传统命令行交互。
- `internal/ui/tui` 负责 TUI 页面。
- `internal/player/hlsproxy.go` 负责 `m3u8` 本地代理兼容层。
- `internal/source/episode_group.go` 负责剧集归类和重复线路合并。
- `internal/cli/commands` 尽量只保留命令入口，跨层流程已经开始下沉到 `internal/workflow`。

---

## 构建

### 前置要求

- Go 1.22+
- 如果使用 `make`，需要 GNU Make 和 Bash 兼容环境

### 构建产物规范

- 仓库内保留型构建产物统一放在 `bin/`。
- 临时验证用的二进制建议输出到系统临时目录，例如 Windows 下用 `$env:TEMP`。
- 不建议在仓库根目录直接生成 `goani` 或 `goani.exe`，避免和源码、文档混在一起。

### 本地构建

```bash
mkdir -p bin
go build -o ./bin/goani ./cmd/goani
```

Windows PowerShell：

```powershell
New-Item -ItemType Directory -Force .\bin | Out-Null
go build -o .\bin\goani.exe .\cmd\goani
```

### 使用 Makefile

```bash
make build
make build-all
make test
make install
```

其中 `make test` 当前等价于：

```bash
go test ./...
```

### 版本信息注入

```bash
mkdir -p bin
go build -ldflags="-s -w \
  -X github.com/Yyyangshenghao/goani-cli/internal/version.Version=v0.1.0 \
  -X github.com/Yyyangshenghao/goani-cli/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/Yyyangshenghao/goani-cli/internal/version.BuildDate=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
  -o ./bin/goani ./cmd/goani
```

PowerShell 示例：

```powershell
$version = "v0.1.0"
$gitCommit = git rev-parse --short HEAD
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd_HH:mm:ss")
New-Item -ItemType Directory -Force .\bin | Out-Null

go build -ldflags "-s -w -X github.com/Yyyangshenghao/goani-cli/internal/version.Version=$version -X github.com/Yyyangshenghao/goani-cli/internal/version.GitCommit=$gitCommit -X github.com/Yyyangshenghao/goani-cli/internal/version.BuildDate=$buildDate" -o .\bin\goani.exe .\cmd\goani
```

---

## 测试

### 自动化测试

真实的 Go 单元测试应放在被测包旁边，文件名以 `*_test.go` 结尾。当前已经在这些位置放了测试：

- `internal/player/hlsproxy_test.go`
- `internal/source/episode_group_test.go`

统一运行：

```bash
go test ./...
```

基础静态检查：

```bash
go vet ./...
```

CI 会在 push / pull request 上自动执行：

- `go test ./...`
- `go vet ./...`
- `go build -o <temp> ./cmd/goani`

### 手动调试入口

手动 smoke 工具现在放在 `cmd/goani-debug-*` 下，它们是开发辅助程序，不是 `go test` 识别的测试文件。

例如：

```bash
go run .\cmd\goani-debug-source
go run .\cmd\goani-debug-player
go run .\cmd\goani-debug-potplayer
```

这类程序适合做本地验证、播放器探针和临时排查。后续如果再增加，优先保持“开发工具”定位，不要把它们当成自动测试。

---

## 架构说明

### 当前命令模型

- `goani` 默认显示帮助
- `goani tui` 进入主 TUI
- `goani search`、`goani play`、`goani source`、`goani config` 继续作为显式 CLI 入口
- `proxy-hls` 是内部隐藏命令，只给播放器兼容层使用，不对普通用户展示

### 命令与编排

命令层应该尽量薄，负责：

- 解析参数
- 调用应用编排
- 将状态交给 TUI 或 CLI 输出

复杂流程尽量放在应用层或专门的流程函数里，例如：

- 搜索后串起选番、选集、线路选择和播放
- 配置页的保存回调
- `PotPlayer + m3u8` 的本地代理播放

### 媒体源与播放器

媒体源、播放器和配置都已经开始分层：

- `internal/source` 负责媒体源、订阅、归类和解析
- `internal/player` 负责播放器发现、默认播放器和 HLS 兼容层
- `internal/app` 负责把配置、播放器和媒体源组织成一个应用对象
- `internal/workflow` 负责把 CLI、TUI、播放器和媒体源串成完整操作流程

当前比较重要的实现约定是：

- `GroupEpisodes` 优先按数字归类，再兜底到标题。
- `goani config player <name> <path>` 只保存路径，不自动设置默认播放器。
- `goani play` 会在真正播放时补默认播放器。
- `PotPlayer + m3u8` 会优先走本地代理。

---

## 命令注册

命令通过 `init()` 注册，新增命令需要实现 `Command` 接口并调用 `Register()`。

```go
func init() {
	Register(&MyCommand{})
}
```

如果是内部命令，记得实现 `HiddenCommand`，避免显示到顶层帮助里。

---

## 贡献建议

1. 改动尽量按功能块提交，不要把文档、代码和测试混成一坨。
2. 复杂逻辑优先补 `*_test.go`，不要只靠 `cmd/goani-debug-*` 里的手动工具。
3. 注释要解释边界、原因和副作用，不要重复函数名本身。
4. 改完后至少跑一次 `go build` 和 `go test ./...`。

---

## 提交规范

建议继续使用 Conventional Commits：

- `feat:` 新功能
- `fix:` 修复问题
- `docs:` 文档更新
- `refactor:` 结构调整
- `test:` 测试相关
- `chore:` 构建或工具调整
