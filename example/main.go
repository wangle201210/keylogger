package main

import (
	"fmt"
	"log"
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
	// 创建文件存储
	dbPath := "keylogger.db"
	var storage keylogger.Storage
	var err error
	storage, err = keylogger.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Printf("Failed to create file storage: %v, using no-op storage", err)
		storage = &keylogger.NoopStorage{}
	} else {
		fmt.Printf("Storage initialized: %s\n", dbPath)
	}
	defer storage.Close()

	// 设置信号处理，用于优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// 在后台goroutine中启动键盘监听
	go func() {
		// 使用 StartWithStorage 启动键盘监听并保存到存储
		keylogger.StartWithStorage(func(event keylogger.KeyEvent) {
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
			fmt.Printf("%s %-12s %s\n", status, event.KeyName, modifierStr)
		}, storage)
	}()

	// 主Go程序等待退出信号
	fmt.Println("程序正在后台监听全局键盘输入... (按 Ctrl+C 退出)")
	fmt.Println("所有按键记录将保存到日志文件:", dbPath)
	<-sigChan
	fmt.Println("\n程序已退出")
}
