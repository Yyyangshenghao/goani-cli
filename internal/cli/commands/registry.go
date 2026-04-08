package commands

import "sync"

// 全局命令注册表
var (
	registry = make(map[string]Command)
	mu       sync.RWMutex
)

// Register 注册命令
func Register(cmd Command) {
	mu.Lock()
	defer mu.Unlock()
	registry[cmd.Name()] = cmd
}

// Get 获取命令
func Get(name string) (Command, bool) {
	mu.RLock()
	defer mu.RUnlock()
	cmd, ok := registry[name]
	return cmd, ok
}

// All 获取所有命令
func All() map[string]Command {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]Command, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}
