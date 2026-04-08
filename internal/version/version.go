package version

import (
	"fmt"
	"runtime"
)

var (
	// 版本号，通过 -ldflags 在编译时注入
	// 示例: go build -ldflags "-X github.com/Yyyangshenghao/goani-cli/internal/version.Version=v0.1.0"
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// Info 返回版本信息
func Info() string {
	return fmt.Sprintf("goani %s\n  Git commit: %s\n  Build date: %s\n  Go version: %s\n  Platform: %s/%s",
		Version, GitCommit, BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// Short 返回简短版本号
func Short() string {
	return Version
}
