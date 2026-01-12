#include <Carbon/Carbon.h>
#include <Cocoa/Cocoa.h>

// 启动监听的 C 函数声明
int startGlobalKeyListener();

// 检查是否有辅助功能权限
int checkAccessibilityPermission();

// 打开系统设置的辅助功能页面
void openAccessibilitySettings();
