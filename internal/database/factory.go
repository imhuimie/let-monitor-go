// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Database Factory Pattern"
//   Timestamp: "2025-11-25T14:00:00Z"
//   Authoring_Role: "AR"
//   Analysis_Performed: "Implemented Factory pattern for database selection"
//   Principle_Applied: "Aether-Engineering-SOLID-O (Open/Closed Principle)"
//   Quality_Check: "Supports easy addition of new database types"
// }}

package database

import (
	"fmt"
	"strings"
)

// DatabaseType represents the type of database to use
type DatabaseType string

const (
	TypeMongoDB DatabaseType = "mongodb"
	TypeSQLite  DatabaseType = "sqlite"
)

// NewDatabase creates a new database instance based on the provided type and connection string
func NewDatabase(dbType DatabaseType, connectionString string) (Database, error) {
	switch strings.ToLower(string(dbType)) {
	case string(TypeMongoDB):
		return NewMongoDB(connectionString)
	case string(TypeSQLite):
		return NewSQLite(connectionString)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}
