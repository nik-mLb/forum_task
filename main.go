package main

import (
	"context"
	"database/sql"
	"log"
	"nik-mLb/forum_task.com/internal/repository"
	"nik-mLb/forum_task.com/internal/transport"
	"nik-mLb/forum_task.com/internal/usecase"
	"runtime"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	LoggerFormat = "${time_rfc3339}, method = ${method}, uri = ${uri}," +
		" status = ${status}, remote_ip = ${remote_ip}\n"
)

func main() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: LoggerFormat}))

	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU * 2)
	log.Printf("Running with GOMAXPROCS=%d", numCPU*2)	

	// Инициализация базы данных
	db, err := sql.Open("pgx", "postgresql://forum:forum@localhost:5432/forum?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(100)              
	db.SetMaxIdleConns(50)              
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute) 

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Инициализация структуры БД
	if err := repository.InitDB(db); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Инициализация репозиториев
	svRepo := repository.NewServiceRepo(db)
	svUC := usecase.NewServiceUseCase(svRepo)
	serviceHandler := transport.NewServiceHandlers(svUC)
	serviceHandler.NewServiceHandlers(e)

	userRepo := repository.NewUserRepository(db)
    userUC := usecase.NewUserUseCase(userRepo)
    userHandler := transport.NewUserHandlers(userUC)
    userHandler.NewUserHandlers(e)

	forumRepo := repository.NewForumRepository(db)
	forumUC := usecase.NewForumUseCase(forumRepo, userRepo)
	forumHandler := transport.NewForumHandlers(forumUC)
	forumHandler.NewForumHandlers(e)

	threadRepo := repository.NewThreadRepository(db)
	threadUC := usecase.NewThreadUseCase(threadRepo, userRepo)
	threadHandler := transport.NewThreadHandlers(threadUC)
	threadHandler.NewThreadHandlers(e)

	postRepo := repository.NewPostRepository(db)
	postUC := usecase.NewPostUseCase(postRepo, userRepo, forumRepo, threadRepo)
	postHandler := transport.NewPostHandlers(postUC)
	postHandler.NewPostHandlers(e)

	// Запуск сервера
	e.Logger.Fatal(e.Start(":5000"))
}