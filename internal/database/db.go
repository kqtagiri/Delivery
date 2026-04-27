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

func NewDB(ctx *context.Context) (error, *DB) {

	Conn_string := os.Getenv("CONN_STRING")
	Conn, err := pgx.Connect(*ctx, Conn_string)
	if err != nil {
		return err, nil
	}
	return nil, &DB{
		Conn: Conn,
		Ctx:  *ctx,
	}

}
