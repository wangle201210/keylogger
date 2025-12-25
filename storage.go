package keylogger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Storage 定义键盘事件存储接口
type Storage interface {
	// Save 保存键盘事件
	Save(event KeyEvent) error
	// Close 关闭存储连接
	Close() error
}

// NoopStorage 空存储实现，不保��任何数据
type NoopStorage struct{}

// Save 不保存任何数据
func (n *NoopStorage) Save(event KeyEvent) error {
	return nil
}

// Close 关闭空存储
func (n *NoopStorage) Close() error {
	return nil
}

// FileStorage 文件日志存储实现
type FileStorage struct {
	file *os.File
	mu   sync.Mutex
}

// NewFileStorage 创建新的文件存储实例
func NewFileStorage(filePath string) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileStorage{
		file: file,
	}, nil
}

// Save 保存键盘事件到日志文件
func (f *FileStorage) Save(event KeyEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	status := "⬇️  按下"
	if !event.IsDown {
		status = "⬆️  释放"
	}

	logLine := fmt.Sprintf("[%s] %s KeyCode=%d KeyName=%s ModifierFlags=%s\n",
		timestamp, status, event.KeyCode, event.KeyName, GetModifierFlags(event.ModifierFlags))

	if _, err := f.file.WriteString(logLine); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	return nil
}

// Close 关闭文件
func (f *FileStorage) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
