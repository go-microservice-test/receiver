package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	dbutils "go-test/db-utils"
	"go-test/middleware"
	"go-test/routers"
	"go-test/utils"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

var (
	mu sync.Mutex // DB mutex
	db *gorm.DB
)

func main() {
	// load configuration
	_cfg := LoadConfiguration("config.json")
	// setup connection
	db = dbutils.Connect(_cfg.DBUser, _cfg.DBPassword, _cfg.DBHost, _cfg.DBName, _cfg.DBSSLMode, _cfg.DBPort)
	// get an engine instance
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// middleware
	r.Use(middleware.CORSMiddleware())
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

	// setup database health checking loop every 10 seconds
	go utils.DataBaseHealthPollingLoop(db, time.Duration(_cfg.DBHeathInterval)*time.Second)
	// run the server
	err := r.Run(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
