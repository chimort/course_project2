package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/pkg/logger"
	"github.com/chimort/course_project2/iternal/user/repository"
	"github.com/chimort/course_project2/iternal/user/service"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/grpc"
)

func main() {
	log := logger.NewLogger("user-service", slog.LevelInfo)

	// Параметры подключения к БД из env
	dbHost := os.Getenv("DB_HOST")         // например "db"
	dbPort := os.Getenv("DB_PORT")         // например "5432"
	dbUser := os.Getenv("DB_USER")         // например "postgres"
	dbPassword := os.Getenv("DB_PASSWORD") // например "postgres"
	dbName := os.Getenv("DB_NAME")         // например "users"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Error("DB ping failed, retrying...", "error", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Error("failed to connect to DB after retries", "error", err)
		os.Exit(1)
	}
	log.Info("DB connected")

	// Миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("failed to create migration driver", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/iternal/user/migrations", 
		"postgres",
		driver,
	)
	if err != nil {
		log.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("No migrations to apply")
		} else {
			log.Error("failed to apply migrations", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Migrations applied")
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, log)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, service.NewUserServer(userService))

	log.Info("UserService running", "addr", ":50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
