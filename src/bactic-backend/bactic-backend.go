package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	tx, err := conn.Begin(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to initialize transaction: %v\n", err)
		os.Exit(1)
	}

	schemaQuery, err := os.ReadFile("./database/schema.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read schema file: %v\n", err)
		os.Exit(1)
	}

	tx.Exec(context.Background(), string(schemaQuery))

	if tx.Commit(context.Background()) != nil {
		fmt.Fprintf(os.Stderr, "unable to commit transaction: %v\n", err)
	}

	return
}
