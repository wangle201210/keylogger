# macOS Keylogger Package

一个用于监听 macOS 全局键盘事件的 Go 包。

## 功能特性

- 监听全局键盘输入（包括所有应用程序）
- 提供人类可读的键名（基于 Apple 官方键码映射）
- 支持修饰键检测（Command, Shift, Control, Option, Caps Lock, Fn）
- 简单易用的回调 API

## 系统要求

- macOS 10.12+
- Go 1.24.10+
- Xcode Command Line Tools

## 权限要求

此包需要 **辅助功能** 权限才能工作：

1. 打开"系统设置"
2. 前往"隐私与安全性" > "辅助功能"
3. 点击"+"号添加你的应用程序
4. 确保 switch 为开启状态

## 安装

```bash
go get github.com/wangle201210/keylogger
```

## 使用方法

### 基本用法

```go
package main

import (
    "fmt"
    "strings"
    "github.com/wangle201210/keylogger"
)

func main() {
    // 在 goroutine 中启动监听
    go keylogger.Start(func(event keylogger.KeyEvent) {
        if event.IsDown {
            modifiers := keylogger.GetModifierFlags(event.ModifierFlags)
            modifierStr := ""
            if len(modifiers) > 0 {
                modifierStr = " + " + strings.Join(modifiers, " ")
            }
            fmt.Printf("按下: %s%s\n", event.KeyName, modifierStr)
        }
    })

    // 保持程序运行
    select {}
}
```

### 完整示例

查看 [example](./example) 目录获取完整的使用示例。

```bash
cd example
go run main.go
```

## API 文档

### KeyEvent 结构体

```go
type KeyEvent struct {
    KeyCode       int    // 键码 (16进制)
    KeyName       string // 键名 (如 "A", "Return", "F1" 等)
    IsDown        bool   // true=按下, false=释放
    ModifierFlags int    // 修饰键标志位
}
```

### 导出的函数

#### `Start(handler KeyEventHandler)`

启动键盘监听。

**注意：** 必须在单独的 goroutine 中调用，因为此函数会阻塞。

参数：
- `handler`: 处理键盘事件的回调函数

示例：
```go
go keylogger.Start(func(event keylogger.KeyEvent) {
    fmt.Printf("键码: %d, 键名: %s\n", event.KeyCode, event.KeyName)
})
```

#### `GetKeyName(keyCode int) string`

获取指定键码的键名。

参数：
- `keyCode`: 键码

返回：
- 键名（如 "A", "Return", "F1" 等）

#### `GetModifierFlags(flags int) []string`

获取修饰键标志的数组表示。

参数：
- `flags`: 修饰键标志位

返回：
- 修饰键数组（如 `[]string{"⌘", "⇧"}`）

可能的返回值：
- `"Fn"` - Fn 键
- `"⇧"` - Shift 键
- `"⌃"` - Control 键
- `"⌥"` - Option/Alt 键
- `"⌘"` - Command 键
- `"Caps"` - Caps Lock 键

示例：
```go
modifiers := keylogger.GetModifierFlags(event.ModifierFlags)
if len(modifiers) > 0 {
    // 输出: + ⌘ ⇧
    fmt.Printf(" + %s\n", strings.Join(modifiers, " "))
}
```

## 键码映射

此包使用 Apple 官方的键码映射表，来源：
`/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/System/Library/Frameworks/Carbon.framework/Versions/A/Frameworks/HIToolbox.framework/Versions/A/Headers/Events.h`

支持的键包括：
- 所有字母和数字键
- 功能键（F1-F20）
- 箭头键、编辑键（Home, End, Page Up, Page Down）
- 修饰键（Command, Shift, Control, Option）
- 音量和媒体键
- 数字小键盘

## 技术实现

- 使用 CGEventTap API 捕获全局键盘事件
- 通过 cgo 调用 Carbon/Cocoa 框架
- 支持所有 macOS 键盘布局

## 注意事项

1. **安全性**：此包仅用于学习和开发目的。使用键盘监听功能时请遵守当地法律法规。
2. **性能**：事件处理回调应该快速执行，避免阻塞。
3. **权限**：首次使用需要授予辅助功能权限。
4. **平台限制**：仅支持 macOS，不支持其他平台。

## 许可证

MIT License

## 参考

- [Apple Technical Note TN2450](https://developer.apple.com/library/archive/technotes/tn2450/_index.html)
- [Carbon HIToolbox Events.h](/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/System/Library/Frameworks/Carbon.framework/Versions/A/Frameworks/HIToolbox.framework/Versions/A/Headers/Events.h)
