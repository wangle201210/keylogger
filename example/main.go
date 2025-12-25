package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/wangle201210/keylogger"
)

func main() {
	// 提示用户权限问题
	fmt.Println("===========================================")
	fmt.Println("注意：此程序需要 '辅助功能' 权限。")
	fmt.Println("请前往 系统设置 > 隐私与安全性 > 辅助功能 中添加本程序。")
	fmt.Println("===========================================")
	fmt.Println()

	// 设置信号处理，用于优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 在后台goroutine中启动键盘监听
	go func() {
		// 定义事件处理函数
		keylogger.Start(func(event keylogger.KeyEvent) {
			status := "⬇️  按下"
			if !event.IsDown {
				status = "⬆️  释放"
			}

			modifiers := keylogger.GetModifierFlags(event.ModifierFlags)
			modifierStr := ""
			if len(modifiers) > 0 {
				modifierStr = " +" + strings.Join(modifiers, " ")
			}

			// 输出可读的格式
			fmt.Printf("%s %-20s %s\n", status, event.KeyName, modifierStr)
		})
	}()

	// 主Go程序等待退出信号
	fmt.Println("程序正在后台监听全局键盘输入... (按 Ctrl+C 退出)")
	<-sigChan
	fmt.Println("\n程序已退出")
}
