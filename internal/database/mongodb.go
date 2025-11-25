// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "MongoDB Database Handler"
//   Timestamp: "2025-11-25T13:42:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python MongoDB operations from core.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S, Interface Segregation"
//   Quality_Check: "Connection pooling, error handling, and index creation implemented"
// }}

package database

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB implements the Database interface
type MongoDB struct {
	client   *mongo.Client
	db       *mongo.Database
	threads  *mongo.Collection
	comments *mongo.Collection
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(uri string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("无法连接到 MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("无法 ping MongoDB: %w", err)
	}

	db := client.Database("forum_monitor")
	m := &MongoDB{
		client:   client,
		db:       db,
		threads:  db.Collection("threads"),
		comments: db.Collection("comments"),
	}

	if err := m.createIndexes(); err != nil {
		return nil, fmt.Errorf("无法创建索引: %w", err)
	}

	log.Info("MongoDB 连接成功")
	return m, nil
}

// createIndexes creates necessary indexes
func (m *MongoDB) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Thread indexes
	threadIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "link", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "pub_date", Value: -1}},
		},
	}

	if _, err := m.threads.Indexes().CreateMany(ctx, threadIndexes); err != nil {
		return fmt.Errorf("创建 threads 索引失败: %w", err)
	}

	// Comment indexes
	commentIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "comment_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "thread_url", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	if _, err := m.comments.Indexes().CreateMany(ctx, commentIndexes); err != nil {
		return fmt.Errorf("创建 comments 索引失败: %w", err)
	}

	return nil
}

// InsertThread inserts a new thread
func (m *MongoDB) InsertThread(thread *Thread) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.threads.InsertOne(ctx, thread)
	if mongo.IsDuplicateKeyError(err) {
		return nil // Already exists, ignore
	}
	return err
}

// FindThread finds a thread by link
func (m *MongoDB) FindThread(link string) (*Thread, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var thread Thread
	err := m.threads.FindOne(ctx, bson.M{"link": link}).Decode(&thread)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &thread, err
}

// UpdateThreadLastPage updates the last_page field
func (m *MongoDB) UpdateThreadLastPage(link string, page int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.threads.UpdateOne(
		ctx,
		bson.M{"link": link},
		bson.M{"$set": bson.M{"last_page": page}},
	)
	return err
}

// InsertComment inserts a new comment
func (m *MongoDB) InsertComment(comment *Comment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.comments.InsertOne(ctx, comment)
	if mongo.IsDuplicateKeyError(err) {
		return nil // Already exists, ignore
	}
	return err
}

// FindComment finds a comment by comment_id
func (m *MongoDB) FindComment(commentID string) (*Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var comment Comment
	err := m.comments.FindOne(ctx, bson.M{"comment_id": commentID}).Decode(&comment)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &comment, err
}

// CommentExists checks if a comment exists
func (m *MongoDB) CommentExists(commentID string) bool {
	comment, err := m.FindComment(commentID)
	return err == nil && comment != nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("断开 MongoDB 连接失败: %w", err)
	}

	log.Info("MongoDB 连接已关闭")
	return nil
}

// Ping checks if the connection is alive
func (m *MongoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.client.Ping(ctx, nil)
}
