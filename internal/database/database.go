// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Database Interface Abstraction"
//   Timestamp: "2025-11-25T13:58:00Z"
//   Authoring_Role: "AR"
//   Analysis_Performed: "Analyzed MongoDB operations to extract common interface"
//   Principle_Applied: "Aether-Engineering-SOLID-I (Interface Segregation)"
//   Quality_Check: "Interface supports both MongoDB and SQLite implementations"
// }}

package database

import "time"

// Database defines the interface for database operations
type Database interface {
	// Thread operations
	InsertThread(thread *Thread) error
	FindThread(link string) (*Thread, error)
	UpdateThreadLastPage(link string, page int) error

	// Comment operations
	InsertComment(comment *Comment) error
	FindComment(commentID string) (*Comment, error)
	CommentExists(commentID string) bool

	// Connection management
	Disconnect() error
	Ping() error
}

// Thread represents a forum thread/post
type Thread struct {
	ID          interface{} `json:"id"` // string for SQLite, ObjectID for MongoDB
	Domain      string      `json:"domain"`
	Category    string      `json:"category"`
	Title       string      `json:"title"`
	Link        string      `json:"link"`
	Description string      `json:"description"`
	Creator     string      `json:"creator"`
	PubDate     time.Time   `json:"pub_date"`
	CreatedAt   time.Time   `json:"created_at"`
	LastPage    int         `json:"last_page"`
}

// Comment represents a comment on a thread
type Comment struct {
	ID                interface{} `json:"id"` // string for SQLite, ObjectID for MongoDB
	CommentID         string      `json:"comment_id"`
	ThreadURL         string      `json:"thread_url"`
	Author            string      `json:"author"`
	Message           string      `json:"message"`
	CreatedAt         time.Time   `json:"created_at"`
	CreatedAtRecorded time.Time   `json:"created_at_recorded"`
	URL               string      `json:"url"`
}
