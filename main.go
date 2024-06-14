package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	dbutils "go-test/db-utils"
	"go-test/middleware"
	"go-test/routers"
	"log"
	"sync"
)

var (
	mu sync.Mutex // DB mutex
	db *sql.DB
)

func main() {
	// load configuration
	_cfg := LoadConfiguration("config.json")
	// setup connection
	db = dbutils.Connect(_cfg.DBUser, _cfg.DBPassword, _cfg.DBHost, _cfg.DBName, _cfg.DBSSLMode, _cfg.DBPort)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)
	// get an engine instance
	r := gin.Default()

	r.Use(middleware.ApiMiddleware(db, mu))

	// connect routers
	// middleware for connect and trace handlers
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "CONNECT" {
			routers.ConnectHandler(c)
		} else if c.Request.Method == "TRACE" && c.Request.URL.Path == "/animals" {
			routers.TraceAnimalRoute(c)
		} else {
			c.Next()
		}
	})
	r.GET("/animals", routers.GetAnimals)
	r.HEAD("/animals", routers.GetAnimalCount)
	r.GET("/animals/:id", routers.GetAnimalByID)
	r.POST("/animals", routers.CreateAnimal)
	r.PUT("/animals/:id", routers.ReplaceAnimal)
	r.DELETE("/animals/:id", routers.DeleteAnimal)
	r.OPTIONS("/*path", routers.OptionsHandler)                          // all URLs
	r.PATCH("/animals/:id/description", routers.UpdateAnimalDescription) // change only description field

	// run the server
	err := r.Run(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
