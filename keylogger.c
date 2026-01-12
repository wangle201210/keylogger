#include "keylogger.h"
#include <stdio.h>
#include <time.h>
#include <string.h>
#include <unistd.h>
#include <stdarg.h>
#include <sys/stat.h>

// 日志文件
static FILE *logFile = NULL;

// 获取日志文件路径
static void getLogFilePath(char *path, size_t size) {
    const char *home = getenv("HOME");
    if (home) {
        snprintf(path, size, "%s/.keylogger/keylogger.log", home);
    } else {
        snprintf(path, size, "/tmp/keylogger.log");
    }
}

// 确保日志目录存在
static void ensureLogDirectory(const char *logPath) {
    char dirPath[512];
    const char *lastSlash = strrchr(logPath, '/');
    if (lastSlash) {
        size_t dirLen = lastSlash - logPath;
        strncpy(dirPath, logPath, dirLen);
        dirPath[dirLen] = '\0';
        mkdir(dirPath, 0755);
    }
}

// 写日志
static void writeLog(const char *level, const char *format, ...) {
    if (!logFile) {
        char logPath[512];
        getLogFilePath(logPath, sizeof(logPath));

        // 确保日志目录存在
        ensureLogDirectory(logPath);

        logFile = fopen(logPath, "a");
        if (!logFile) {
            return;
        }
    }

    // 获取当前时间
    time_t now = time(NULL);
    struct tm *tm_info = localtime(&now);
    char timestamp[64];
    strftime(timestamp, sizeof(timestamp), "%Y-%m-%d %H:%M:%S", tm_info);

    // 写日志
    va_list args;
    va_start(args, format);
    fprintf(logFile, "[%s] [%s] ", timestamp, level);
    vfprintf(logFile, format, args);
    fprintf(logFile, "\n");
    fflush(logFile);
    va_end(args);
}

// Go函数声明
extern void goKeyCallback(int keyCode, int isDown, int modifierFlags);

// C 回调函数
static CGEventRef eventTapCallback(
    CGEventTapProxy proxy,
    CGEventType type,
    CGEventRef event,
    void *refcon
) {
    int64_t keyCode;
    int64_t isDown;
    int64_t modifiers;

    // 处理普通键盘按下和释放事件
    if (type == kCGEventKeyDown || type == kCGEventKeyUp) {
        keyCode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        isDown = (type == kCGEventKeyDown) ? 1 : 0;
        modifiers = CGEventGetFlags(event);

        writeLog("DEBUG", "键盘事件: type=%s, keyCode=%lld, isDown=%lld, modifiers=%lld",
                 type == kCGEventKeyDown ? "KeyDown" : "KeyUp", keyCode, isDown, modifiers);

        // 调用 Go 函数处理
        goKeyCallback((int)keyCode, (int)isDown, (int)modifiers);
    }
    // 处理 Caps Lock 等修饰键状态改变事件
    else if (type == kCGEventFlagsChanged) {
        keyCode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        modifiers = CGEventGetFlags(event);

        // 检测是否是 Caps Lock (keyCode 0x39 = 57)
        // Caps Lock 是切换键，通过 flags 变化检测
        static int capsLockWasOn = 0;
        int capsLockIsOn = (modifiers & 0x10000) != 0; // 0x10000 是 Caps Lock 标志位

        // 只有当 Caps Lock 状态改变时才触发
        if (capsLockIsOn != capsLockWasOn && keyCode == 0x39) {
            isDown = capsLockIsOn ? 1 : 0;
            writeLog("DEBUG", "Caps Lock 状态改变: isDown=%lld", isDown);
            goKeyCallback((int)keyCode, (int)isDown, (int)modifiers);
            capsLockWasOn = capsLockIsOn;
        }
    }
    // 必须返回事件本身，否则会中断系统事件流
    return event;
}

// 启动监听的 C 函数实现
int startGlobalKeyListener() {
    writeLog("INFO", "========================================");
    writeLog("INFO", "开始初始化全局键盘监听");

    CGEventMask eventMask = CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(kCGEventKeyUp) | CGEventMaskBit(kCGEventFlagsChanged);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,     // 监听整个用户会话
        kCGHeadInsertEventTap,  // 在事件传递前捕获
        kCGEventTapOptionDefault,
        eventMask,
        eventTapCallback,
        NULL
    );

    if (!eventTap) {
        writeLog("ERROR", "创建事件监听失败！请检查辅助功能权限。");
        return 0;
    }

    writeLog("INFO", "事件监听器创建成功");

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(
        kCFAllocatorDefault,
        eventTap,
        0
    );
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);

    writeLog("INFO", "全局键盘监听已启动，进入事件循环");

    CFRunLoopRun(); // 阻塞，进入事件循环
    return 1;
}

// 检查是否有辅助功能权限
int checkAccessibilityPermission() {
    writeLog("INFO", "检查辅助功能权限");

    // 方法1：使用 AXIsProcessTrustedWithOptions 快速检查
    CFMutableDictionaryRef options = CFDictionaryCreateMutable(
        kCFAllocatorDefault,
        1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );

    CFDictionaryAddValue(
        options,
        kAXTrustedCheckOptionPrompt,
        kCFBooleanFalse
    );

    Boolean trusted = AXIsProcessTrustedWithOptions(options);
    CFRelease(options);

    if (!trusted) {
        writeLog("INFO", "辅助功能权限: 未授权");
        return 0;
    }

    // 方法2：尝试创建事件 tap 来真正验证权限
    // 这样可以检测签名改变导致的权限失效问题
    CGEventMask eventMask = CGEventMaskBit(kCGEventKeyDown);
    CFMachPortRef testTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        eventMask,
        eventTapCallback,
        NULL
    );

    if (!testTap) {
        writeLog("INFO", "辅助功能权限: 已失效（签名改变）");
        return 0;
    }

    // 成功创建，说明真的有权限
    CFRelease(testTap);
    writeLog("INFO", "辅助功能权限: 已授权且有效");
    return 1;
}

// 打开系统设置的辅助功能页面
void openAccessibilitySettings() {
    writeLog("INFO", "打开系统设置的辅助功能页面");

    // macOS 13 (Ventura) 及以上版本使用新的设置路径
    if (@available(macOS 13.0, *)) {
        // 打开系统设置 > 隐私与安全性 > 辅助功能
        NSURL *settingsURL = [NSURL URLWithString:@"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"];
        [[NSWorkspace sharedWorkspace] openURL:settingsURL];
    } else {
        // macOS 12 (Monterey) 及以下版本使用旧的设置路径
        NSURL *settingsURL = [NSURL URLWithString:@"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"];
        [[NSWorkspace sharedWorkspace] openURL:settingsURL];
    }
}
