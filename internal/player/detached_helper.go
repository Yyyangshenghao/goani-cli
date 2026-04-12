package player

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func startDetachedHelper(commandName string, ctx StreamRequestContext) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取当前可执行文件失败: %w", err)
	}

	encodedContext, err := EncodeStreamRequestContext(ctx)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(exe, commandName, encodedContext)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("创建后台输出管道失败: %w", err)
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("启动后台服务失败: %w", err)
	}

	type result struct {
		line string
		err  error
	}

	resultCh := make(chan result, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, err := reader.ReadString('\n')
		resultCh <- result{line: strings.TrimSpace(line), err: err}
	}()

	select {
	case res := <-resultCh:
		if res.err != nil && strings.TrimSpace(res.line) == "" {
			return "", fmt.Errorf("读取后台地址失败: %w", res.err)
		}
		if strings.TrimSpace(res.line) == "" {
			return "", fmt.Errorf("后台服务没有返回可用地址")
		}
		return strings.TrimSpace(res.line), nil
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("启动后台服务超时")
	}
}
