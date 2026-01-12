//go:build darwin

package keylogger

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#include "keylogger.h"
*/
import "C"
import (
	"fmt"
	"sync/atomic"
)

// macOS 键码映射表 (基于官方 /Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/System/Library/Frameworks/Carbon.framework/Versions/A/Frameworks/HIToolbox.framework/Versions/A/Headers/Events.h)
var keyCodeMap = map[int]string{
	// US-ANSI 键盘位置键 (字符键)
	0x00: "A",
	0x01: "S",
	0x02: "D",
	0x03: "F",
	0x04: "H",
	0x05: "G",
	0x06: "Z",
	0x07: "X",
	0x08: "C",
	0x09: "V",
	0x0B: "B",
	0x0C: "Q",
	0x0D: "W",
	0x0E: "E",
	0x0F: "R",
	0x10: "Y",
	0x11: "T",
	0x12: "1",
	0x13: "2",
	0x14: "3",
	0x15: "4",
	0x16: "6",
	0x17: "5",
	0x18: "=",
	0x19: "9",
	0x1A: "7",
	0x1B: "-",
	0x1C: "8",
	0x1D: "0",
	0x1E: "]",
	0x1F: "O",
	0x20: "U",
	0x21: "[",
	0x22: "I",
	0x23: "P",
	0x25: "L",
	0x26: "J",
	0x27: "'",
	0x28: "K",
	0x29: ";",
	0x2A: "\\",
	0x2B: ",",
	0x2C: "/",
	0x2D: "N",
	0x2E: "M",
	0x2F: ".",
	0x32: "`",

	// 布局无关的功能键
	0x24: "Return", // 主键盘上的回车
	0x30: "Tab",
	0x31: "Space",
	0x33: "Backspace", // Delete (向后删除)
	0x35: "Escape",
	0x37: "⌘ Command (L)",
	0x38: "⇧ Shift (L)",
	0x39: "Caps Lock",
	0x3A: "⌥ Option (L)",
	0x3B: "⌃ Control (L)",
	0x3C: "⇧ Shift (R)",
	0x3D: "⌥ Option (R)",
	0x3E: "⌃ Control (R)",
	0x3F: "Fn",
	0x36: "⌘ Command (R)",
	0x4C: "Enter", // 可能是小键盘或扩展键盘的 Enter

	// 功能键
	0x7A: "F1",
	0x78: "F2",
	0x63: "F3",
	0x76: "F4",
	0x60: "F5",
	0x61: "F6",
	0x62: "F7",
	0x64: "F8",
	0x65: "F9",
	0x6D: "F10",
	0x67: "F11",
	0x6F: "F12",
	0x69: "F13",
	0x6B: "F14",
	0x71: "F15",
	0x6A: "F16",
	0x40: "F17",
	0x4F: "F18",
	0x50: "F19",
	0x5A: "F20",

	// 箭头键
	0x7B: "←",
	0x7C: "→",
	0x7D: "↓",
	0x7E: "↑",

	// 特殊键
	0x72: "Help",
	0x73: "Home",
	0x74: "Page Up",
	0x75: "Forward Delete",
	0x77: "End",
	0x79: "Page Down",

	// 音量和媒体键
	0x48: "Volume Up",
	0x49: "Volume Down",
	0x4A: "Mute",

	// 数字小键盘
	0x41: "Keypad .",
	0x43: "Keypad *",
	0x45: "Keypad +",
	0x47: "Keypad Clear",
	0x4B: "Keypad /",
	0x4E: "Keypad -",
	0x51: "Keypad =",
	0x52: "Keypad 0",
	0x53: "Keypad 1",
	0x54: "Keypad 2",
	0x55: "Keypad 3",
	0x56: "Keypad 4",
	0x57: "Keypad 5",
	0x58: "Keypad 6",
	0x59: "Keypad 7",
	0x5B: "Keypad 8",
	0x5C: "Keypad 9",

	// 其他键 (如果有)
	0x66: "Home",     // 可能是另一个 Home
	0x68: "Snapshot", // 可能是 Print Screen
	0xB3: "Fn",       // Function 键变体（某些键盘）
}

// KeyEvent 表示一个键盘事件
type KeyEvent struct {
	KeyCode       int    // 键码
	KeyName       string // 键名
	IsDown        bool   // true=按下, false=释放
	ModifierFlags int    // 修饰键标志
}

// KeyEventHandler 是处理键盘事件的回调函数类型
type KeyEventHandler func(event KeyEvent)

var eventHandler KeyEventHandler
var eventStorage Storage

// isRunning 用于防止重复启动键盘监听
// 使用 int32 而不是 bool 以支持原子操作
var isRunning int32 // 0 = 未运行, 1 = 运行中

// getKeyName 获取键名
func getKeyName(keyCode int) string {
	if name, ok := keyCodeMap[keyCode]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(0x%02X)", keyCode)
}

// getModifierFlags 解析修饰键标志，返回修饰键数组
func getModifierFlags(flags int) []string {
	var modifiers []string

	// 检查各种修饰键 (基于 CGEventFlags)
	// 参考: /Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/System/Library/Frameworks/CoreGraphics.framework/Versions/A/Headers/CGEvent.h
	if flags&0x20000 != 0 {
		modifiers = append(modifiers, "⇧")
	}
	if flags&0x40000 != 0 {
		modifiers = append(modifiers, "⌃")
	}
	if flags&0x80000 != 0 {
		modifiers = append(modifiers, "⌥")
	}
	if flags&0x100000 != 0 {
		modifiers = append(modifiers, "⌘")
	}
	if flags&0x10000 != 0 {
		modifiers = append(modifiers, "Caps")
	}

	return modifiers
}

// 导出给 C 调用的 Go 函数
//
//export goKeyCallback
func goKeyCallback(keyCode C.int, isDown C.int, modifierFlags C.int) {
	event := KeyEvent{
		KeyCode:       int(keyCode),
		KeyName:       getKeyName(int(keyCode)),
		IsDown:        isDown != 0,
		ModifierFlags: int(modifierFlags),
	}

	if eventHandler != nil {
		eventHandler(event)
	}

	if eventStorage != nil {
		_ = eventStorage.Save(event)
	}
}

// Start 启动键盘监听
// 需要在单独的 goroutine 中调用
func Start(handler KeyEventHandler) {
	StartWithStorage(handler, nil)
}

// StartWithStorage 启动键盘监听并保存到存储
// 需要在单独的 goroutine 中调用
func StartWithStorage(handler KeyEventHandler, storage Storage) {
	// 使用原子操作检查并设置运行状态，防止重复启动
	if !atomic.CompareAndSwapInt32(&isRunning, 0, 1) {
		// 如果已经在运行，直接返回
		return
	}

	eventHandler = handler
	eventStorage = storage
	C.startGlobalKeyListener()
}

// Stop 停止监听并关闭存储连接
func Stop() {
	if eventStorage != nil {
		_ = eventStorage.Close()
		eventStorage = nil
	}
	eventHandler = nil
	// 重置运行状态
	atomic.StoreInt32(&isRunning, 0)
}

// GetKeyName 获取指定键码的键名
func GetKeyName(keyCode int) string {
	return getKeyName(keyCode)
}

// GetModifierFlags 获取修饰键标志的数组表示
func GetModifierFlags(flags int) []string {
	return getModifierFlags(flags)
}

// CheckAccessibilityPermission 检查是否有辅助功能权限
// 返回 true 表示有权限，false 表示无权限
func CheckAccessibilityPermission() bool {
	result := C.checkAccessibilityPermission()
	return result != 0
}

// OpenAccessibilitySettings 打开系统设置的辅助功能页面
func OpenAccessibilitySettings() {
	C.openAccessibilitySettings()
}
