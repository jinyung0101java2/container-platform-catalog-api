package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go-api/config"
	_ "go-api/docs"
	"go-api/handler"
	"go-api/middleware"
	"go-api/router"
)

func init() {
	config.InitEnvConfigs()
}

// @title Container Platform Catalog Rest API
// @version 1.0
// @description K-PaaS Container Platform Catalog Rest API
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8093
// @BasePath /
func main() {
	log.Info("Hello, Helm Catalog Rest API!")
	app := fiber.New()
	router.AccessibleRoute(app)
	middleware.FiberMiddleware(app)
	router.APIRoutes(app)
	handler.Settings()
	err := app.Listen(config.Env.ServerPort)
	if err != nil {
		log.Fatal("SERVER IS NOT RUNNING! REASON :: %v", err)
	}
}
