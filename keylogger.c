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
    // 过滤出键盘按下和释放事件
    if (type == kCGEventKeyDown || type == kCGEventKeyUp) {
        int64_t keyCode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        int64_t isDown = (type == kCGEventKeyDown) ? 1 : 0;
        int64_t modifiers = CGEventGetFlags(event);

        // 调用 Go 函数处理
        goKeyCallback((int)keyCode, (int)isDown, (int)modifiers);
    }
    // 必须返回事件本身，否则会中断系统事件流
    return event;
}

// 启动监听的 C 函数实现
int startGlobalKeyListener() {
    CGEventMask eventMask = CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(kCGEventKeyUp);
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
