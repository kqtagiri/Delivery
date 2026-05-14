package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	Conn *pgx.Conn
	Ctx  context.Context
}

func NewDB(ctx context.Context) (*DB, error) {

	Conn_string := os.Getenv("CONN_STRING")
	Conn, err := pgx.Connect(ctx, Conn_string)
	if err != nil {
		return nil, err
	}

	return &DB{
		Conn: Conn,
		Ctx:  ctx,
	}, nil

}

func (db *DB) Close() error {

	if db.Conn != nil {
		return db.Conn.Close(db.Ctx)
	}
	return nil

}
