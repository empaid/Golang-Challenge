package queries

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func GetConnection() *pgx.Conn {
	print("New DB connection")
	conn, err := pgx.Connect(
		context.Background(),
		os.Getenv("DB_CONNECTION_STRING"),
	)

	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}

	return conn
}

func NewUserDB(conn *pgx.Conn) *UserDB {
	return &UserDB{
		conn: conn,
	}
}
