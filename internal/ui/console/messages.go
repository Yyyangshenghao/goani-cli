package console

import "fmt"

// Info 打印信息
func Info(format string, args ...any) {
	fmt.Printf("\n[INFO] %s\n", fmt.Sprintf(format, args...))
}

// Error 打印错误
func Error(format string, args ...any) {
	fmt.Printf("\n[ERROR] %s\n", fmt.Sprintf(format, args...))
}

// Success 打印成功
func Success(format string, args ...any) {
	fmt.Printf("\n[OK] %s\n", fmt.Sprintf(format, args...))
}
