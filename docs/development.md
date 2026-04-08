# 开发指南

## 项目结构

```
goani-cli/
├── cmd/goani/          # 程序入口
├── internal/
│   ├── app/            # 应用核心
│   ├── cli/            # CLI 框架
│   │   └── commands/   # 命令实现
│   ├── config/         # 配置管理
│   ├── player/         # 播放器接口
│   ├── source/         # 媒体源接口
│   │   └── webselector/ # Web 抓取实现
│   ├── ui/             # 终端交互
│   └── version/        # 版本管理
├── mediaSourceJson/    # 媒体源配置
├── docs/               # 文档
└── test/               # 测试
```

---

## 构建

### 前置要求

- Go 1.22+

### 克隆并构建

```bash
git clone https://github.com/Yyyangshenghao/goani-cli.git
cd goani-cli

# 构建当前平台
go build -o goani ./cmd/goani
```

### 使用 Makefile

```bash
make build       # 构建当前平台
make build-all   # 构建所有平台
make test        # 运行测试
make install     # 安装到 GOPATH
```

### 版本信息注入

```bash
# 构建时注入版本信息
go build -ldflags="-s -w \
  -X github.com/Yyyangshenghao/goani-cli/internal/version.Version=v0.1.0 \
  -X github.com/Yyyangshenghao/goani-cli/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/Yyyangshenghao/goani-cli/internal/version.BuildDate=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
  -o goani ./cmd/goani
```

---

## 测试

```bash
# 核心功能测试（搜索、剧集、视频链接）
go run test/source/main.go

# 播放器测试
go run test/player/main.go
```

---

## 架构设计

### 接口设计

项目采用接口驱动设计，核心接口包括：

**媒体源接口** (`internal/source/source.go`)
```go
type Source interface {
    Name() string
    Search(keyword string) ([]Anime, error)
    GetEpisodes(animeURL string) ([]Episode, error)
    GetVideoURL(episodeURL string) (string, error)
}
```

**播放器接口** (`internal/player/player.go`)
```go
type Player interface {
    Name() string
    Play(url string) error
    IsAvailable() bool
}
```

**命令接口** (`internal/cli/commands/command.go`)
```go
type Command interface {
    Name() string
    Run(args []string)
    Usage() string
}
```

### 命令注册

命令通过 `init()` 自动注册，新增命令只需实现 `Command` 接口并在 `init()` 中调用 `Register()`：

```go
func init() {
    Register(&MyCommand{})
}
```

---

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/)：

- `feat:` 新功能
- `fix:` 修复 bug
- `docs:` 文档更新
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建/工具相关
