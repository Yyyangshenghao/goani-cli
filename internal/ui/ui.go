package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Select 从列表中选择一项
func Select(title string, total int, displayFn func(int) string) (int, error) {
	if total == 0 {
		return -1, fmt.Errorf("列表为空")
	}

	fmt.Printf("\n%s:\n", title)
	for i := 0; i < total && i < 10; i++ {
		fmt.Printf("  %d. %s\n", i+1, displayFn(i))
	}
	if total > 10 {
		fmt.Printf("  ... 共 %d 项\n", total)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请选择序号 (q 退出): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return -1, err
		}
		input = strings.TrimSpace(input)

		if input == "q" || input == "Q" {
			return -1, fmt.Errorf("用户取消")
		}

		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > total {
			fmt.Printf("无效输入，请输入 1-%d\n", total)
			continue
		}

		return idx - 1, nil
	}
}

// Confirm 确认操作
func Confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s (y/n): ", message)
		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("请输入 y 或 n")
		}
	}
}

// Input 获取用户输入
func Input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(input)
}
