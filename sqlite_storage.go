package keylogger

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// KeyRecord 数据库记录模型
type KeyRecord struct {
	ID            uint      `gorm:"primaryKey"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index"`
	KeyCode       int       `gorm:"index"`
	KeyName       string    `gorm:"index;size:50"`
	IsDown        bool
	ModifierFlags int
}

// TableName 指定表名
func (KeyRecord) TableName() string {
	return "key_records"
}

// SQLiteStorage SQLite存储实现
type SQLiteStorage struct {
	db *gorm.DB
	mu sync.Mutex
}

// NewSQLiteStorage 创建新的SQLite存储实例
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&KeyRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

// Save 保存键盘事件到数据库（只保存按下事件）
func (s *SQLiteStorage) Save(event KeyEvent) error {
	// 只记录按下事件，不记录释放事件
	if !event.IsDown {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record := KeyRecord{
		KeyCode:       event.KeyCode,
		KeyName:       event.KeyName,
		IsDown:        true,
		ModifierFlags: event.ModifierFlags,
	}

	return s.db.Create(&record).Error
}

// Close 关闭数据库连接
func (s *SQLiteStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return nil
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
