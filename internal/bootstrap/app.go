package bootstrap

import (
	"fmt"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Router *gin.Engine
	Port   string
}

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.New()

	registerMiddleware(router, cfg)
	registerModules(router, db, cfg)

	return router
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	return &App{
		Router: NewRouter(db, cfg),
		Port:   cfg.Server.Port,
	}, nil
}

func (a *App) Run() error {
	return a.Router.Run(":" + a.Port)
}
