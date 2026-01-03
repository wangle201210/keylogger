#include "keylogger.h"

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
            goKeyCallback((int)keyCode, (int)isDown, (int)modifiers);
            capsLockWasOn = capsLockIsOn;
        }
    }
    // 必须返回事件本身，否则会中断系统事件流
    return event;
}

// 启动监听的 C 函数实现
int startGlobalKeyListener() {
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
        NSLog(@"创建事件监听失败！请检查辅助功能权限。");
        return 0;
    }

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(
        kCFAllocatorDefault,
        eventTap,
        0
    );
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    NSLog(@"全局键盘监听已启动");
    CFRunLoopRun(); // 阻塞，进入事件循环
    return 1;
}
