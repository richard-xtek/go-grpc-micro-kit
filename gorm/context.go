package gorm

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"google.golang.org/grpc"
)

type contextKey string

func (c contextKey) String() string {
	return "api context key " + string(c)
}

const (
	dbKey = contextKey("db")
)

// DBUnaryServerInterceptor ...
func DBUnaryServerInterceptor(logger log.Factory, db *gorm.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if db == nil {
			return handler(ctx, req)
		}

		newDb := db.New()
		newCtx := WithDB(ctx, newDb)
		resp, err := handler(newCtx, req)
		return resp, err
	}
}

// GetDB reads the database from the context.
func GetDB(ctx context.Context) *DB {
	obj := ctx.Value(dbKey)
	if obj == nil {
		return nil
	}
	db := obj.(*gorm.DB)
	return &DB{db: db, ctx: ctx}
}

// WithDB adds the database to the context.
func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

// DB ...
type DB struct {
	ctx context.Context
	db  *gorm.DB
}

// Begin ...
func (b *DB) Begin() error {
	b.db = b.db.Begin()
	return nil
}

// NewCtx ...
func (b *DB) NewCtx() context.Context {
	return WithDB(b.ctx, b.db)
}

// Rollback ...
func (b *DB) Rollback() error {
	b.db.Rollback()
	return nil
}

// Commit ...
func (b *DB) Commit() error {
	b.db.Commit()
	return nil
}

// Bg ...
func (b *DB) Bg() *gorm.DB {
	return b.db
}
