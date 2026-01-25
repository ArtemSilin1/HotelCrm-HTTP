package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/config"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/handlers/auth"
	clients_handler "github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/handlers/clients"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/handlers/rooms"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/logger"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/server"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	var serverConfig config.ServerConf
	if err := serverConfig.ReadConfig(); err != nil {
		log.Fatal(err.Error())
	}

	var databaseConfig config.StorageConfig
	if err := databaseConfig.ReadConfig(); err != nil {
		log.Fatal(err.Error())
	}
	var databaseClient storage.DatabaseClient

	startupLog, err := logger.New("System Startup", "main.go", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	startupLog.MessageType = "INFO"
	startupLog.Message = "Приложение запускается..."

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := databaseClient.OpenDBClient(ctx, databaseConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pool.Close()

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	//var __master_user_init__ users.Users
	//if err := __master_user_init__.Test(pool); err != nil {
	//	log.Fatal(err.Error())
	//}

	// Routes
	userHadnler := auth.NewHandler(pool, startupLog)
	userHadnler.InitHandler(router)

	clientHandler := clients_handler.NewHandler(pool, startupLog)
	clientHandler.InitHandler(router)

	roomHandler := rooms.NewHandler(pool, startupLog)
	roomHandler.InitHandler(router)

	initingServer := &server.Server{}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := initingServer.RunServer(serverConfig.Port, router); err != nil {
			log.Fatal(err.Error())
		}
	}()

	log.Printf("\033[32mСервер запущен на: %s:%s\n\033[0m", serverConfig.Host, serverConfig.Port)

	<-done

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := initingServer.StopServer(ctxShutdown); err != nil {
		log.Printf("Ошибка при завершении работы сервера: %v", err)
	}
	log.Println("Сервер успешно остановлен")
}
