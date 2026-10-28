package bootstrap

import (
	"log"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
	Port   string
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %w", err)
	}

	router := gin.Default()
	registerMiddleware(router, cfg)
	registerModules(router, db, cfg)
	return &App{
		Router: router,
		Port:   cfg.Server.Port,
	}, nil
}

func (a *App) Run() error {
	return a.Router.Run(":" + a.Port)
}
