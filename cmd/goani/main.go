package main

import (
    "fmt"
    "log"
    "os"

    "github.com/yshscpu/goani-cli/internal/config"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    cmd := os.Args[1]

    switch cmd {
    case "config":
        handleConfig(os.Args[2:])
    case "search":
        handleSearch(os.Args[2:])
    default:
        fmt.Printf("未知命令: %s\n", cmd)
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("goani - 命令行动漫播放器")
    fmt.Println()
    fmt.Println("用法:")
    fmt.Println("  goani config player <name> <path>  配置播放器")
    fmt.Println("  goani search <keyword>             搜索动漫")
    fmt.Println()
    fmt.Println("示例:")
    fmt.Println("  goani config player mpv \"D:\\MPV播放器\\mpv.exe\"")
    fmt.Println("  goani search 葬送的芙莉莲")
}

func handleConfig(args []string) {
    if len(args) < 3 || args[0] != "player" {
        fmt.Println("用法: goani config player <name> <path>")
        fmt.Println("示例: goani config player mpv \"D:\\MPV播放器\\mpv.exe\"")
        os.Exit(1)
    }

    name := args[1]
    path := args[2]

    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    cfg.SetPlayer(name, path)
    if err := cfg.Save(); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("已配置播放器: %s -> %s\n", name, path)
}

func handleSearch(args []string) {
    if len(args) < 1 {
        fmt.Println("用法: goani search <keyword>")
        os.Exit(1)
    }

    keyword := args[0]
    fmt.Printf("搜索: %s\n", keyword)

    // TODO: 实现搜索逻辑
    fmt.Println("搜索功能开发中...")
}
