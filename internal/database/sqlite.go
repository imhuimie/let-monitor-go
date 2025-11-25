// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "SQLite Database Implementation"
//   Timestamp: "2025-11-25T13:59:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Implemented SQLite adapter matching MongoDB interface"
//   Principle_Applied: "Aether-Engineering-SOLID-S, Interface Implementation"
//   Quality_Check: "Full Database interface implementation with proper indexing"
// }}

package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

// SQLite implements the Database interface
type SQLite struct {
	db *sql.DB
}

// Ensure SQLite implements Database interface
var _ Database = (*SQLite)(nil)

// NewSQLite creates a new SQLite connection
func NewSQLite(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开 SQLite 数据库: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("无法 ping SQLite: %w", err)
	}

	s := &SQLite{db: db}

	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("无法创建表: %w", err)
	}

	log.Info("SQLite 连接成功")
	return s, nil
}

// createTables creates necessary tables and indexes
func (s *SQLite) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS threads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			category TEXT NOT NULL,
			title TEXT NOT NULL,
			link TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL,
			creator TEXT NOT NULL,
			pub_date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			last_page INTEGER DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_threads_link ON threads(link)`,
		`CREATE INDEX IF NOT EXISTS idx_threads_pub_date ON threads(pub_date DESC)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comment_id TEXT NOT NULL UNIQUE,
			thread_url TEXT NOT NULL,
			author TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			created_at_recorded DATETIME NOT NULL,
			url TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_comment_id ON comments(comment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_thread_url ON comments(thread_url, created_at DESC)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("执行查询失败: %w", err)
		}
	}

	return nil
}

// InsertThread inserts a new thread
func (s *SQLite) InsertThread(thread *Thread) error {
	query := `INSERT OR IGNORE INTO threads 
		(domain, category, title, link, description, creator, pub_date, created_at, last_page) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query,
		thread.Domain,
		thread.Category,
		thread.Title,
		thread.Link,
		thread.Description,
		thread.Creator,
		thread.PubDate,
		thread.CreatedAt,
		thread.LastPage,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil && id > 0 {
		thread.ID = id
	}

	return nil
}

// FindThread finds a thread by link
func (s *SQLite) FindThread(link string) (*Thread, error) {
	query := `SELECT id, domain, category, title, link, description, creator, 
		pub_date, created_at, last_page FROM threads WHERE link = ?`

	var thread Thread
	var pubDate, createdAt string

	err := s.db.QueryRow(query, link).Scan(
		&thread.ID,
		&thread.Domain,
		&thread.Category,
		&thread.Title,
		&thread.Link,
		&thread.Description,
		&thread.Creator,
		&pubDate,
		&createdAt,
		&thread.LastPage,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	thread.PubDate, _ = time.Parse("2006-01-02 15:04:05", pubDate)
	thread.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	return &thread, nil
}

// UpdateThreadLastPage updates the last_page field
func (s *SQLite) UpdateThreadLastPage(link string, page int) error {
	query := `UPDATE threads SET last_page = ? WHERE link = ?`
	_, err := s.db.Exec(query, page, link)
	return err
}

// InsertComment inserts a new comment
func (s *SQLite) InsertComment(comment *Comment) error {
	query := `INSERT OR IGNORE INTO comments 
		(comment_id, thread_url, author, message, created_at, created_at_recorded, url) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query,
		comment.CommentID,
		comment.ThreadURL,
		comment.Author,
		comment.Message,
		comment.CreatedAt,
		comment.CreatedAtRecorded,
		comment.URL,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil && id > 0 {
		comment.ID = id
	}

	return nil
}

// FindComment finds a comment by comment_id
func (s *SQLite) FindComment(commentID string) (*Comment, error) {
	query := `SELECT id, comment_id, thread_url, author, message, 
		created_at, created_at_recorded, url FROM comments WHERE comment_id = ?`

	var comment Comment
	var createdAt, createdAtRecorded string

	err := s.db.QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.CommentID,
		&comment.ThreadURL,
		&comment.Author,
		&comment.Message,
		&createdAt,
		&createdAtRecorded,
		&comment.URL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	comment.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	comment.CreatedAtRecorded, _ = time.Parse("2006-01-02 15:04:05", createdAtRecorded)

	return &comment, nil
}

// CommentExists checks if a comment exists
func (s *SQLite) CommentExists(commentID string) bool {
	comment, err := s.FindComment(commentID)
	return err == nil && comment != nil
}

// Disconnect closes the SQLite connection
func (s *SQLite) Disconnect() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("断开 SQLite 连接失败: %w", err)
	}

	log.Info("SQLite 连接已关闭")
	return nil
}

// Ping checks if the connection is alive
func (s *SQLite) Ping() error {
	return s.db.Ping()
}
