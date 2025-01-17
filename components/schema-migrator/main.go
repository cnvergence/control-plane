package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	envs := []string{"DB_USER", "DB_HOST", "DB_NAME", "DB_PORT",
		"DB_PASSWORD", "MIGRATION_PATH", "DIRECTION"}
	for _, env := range envs {
		_, present := os.LookupEnv(env)
		if !present {
			fmt.Printf("ERROR: %s is not set\n", env)
			os.Exit(1)
		}
	}

	direction := os.Getenv("DIRECTION")
	switch direction {
	case "up":
		fmt.Println("Migration UP")
	case "down":
		fmt.Println("Migration DOWN")
	default:
		fmt.Println("ERROR: DIRECTION variable accepts only two values: up or down")
		os.Exit(1)
	}

	dbName := os.Getenv("DB_NAME")

	_, present := os.LookupEnv("DB_SSL")
	if present {
		dbName = fmt.Sprintf("%s?sslmode=%s", dbName, os.Getenv("DB_SSL"))
	}

	hostPort := net.JoinHostPort(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"))

	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		hostPort,
		dbName,
	)

	fmt.Println("# WAITING FOR CONNECTION WITH DATABASE #")
	db, err := sql.Open("postgres", connectionString)

	for i := 0; i < 30 && err != nil; i++ {
		fmt.Printf("Error while connecting to the database, %s\n", err)
		db, err = sql.Open("postgres", connectionString)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		fmt.Printf("# COULD NOT ESTABLISH CONNECTION TO DATABASE #")
		os.Exit(1)
	}

	fmt.Println("# STARTING MIGRATION #")

	migrationPath := fmt.Sprintf("file:///migrate/migrations/%s", os.Getenv("MIGRATION_PATH"))

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	for i := 0; i < 30 && err != nil; i++ {
		fmt.Printf("Error during driver initialization, %s\n", err)
		driver, err = postgres.WithInstance(db, &postgres.Config{})
		time.Sleep(1 * time.Second)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"postgres", driver)
	if err != nil {
		fmt.Printf("Error during migration initialization, %s\n", err)
	}

	defer m.Close()
	m.Log = &Logger{}

	if direction == "up" {
		err = m.Up()
	} else if direction == "down" {
		err = m.Down()
	}

	if err != nil {
		fmt.Printf("Error during migration, %s\n", err)
	}
}

type Logger struct {
}

func (l *Logger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (l *Logger) Verbose() bool {
	return false
}
