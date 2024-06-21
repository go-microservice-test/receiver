package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	ginratelimit "github.com/ljahier/gin-ratelimit"
	"go-test/middleware"
	"go-test/routers"
	"go-test/service"
	"go-test/utils"
	"log"
	"time"
)

func main() {
	// load configuration
	_cfg := utils.LoadConfiguration("config.json")
	// startup databases and begin handlers
	service := service.NewService(&_cfg)

	// get an engine instance
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// middleware
	r.Use(middleware.CORSMiddleware())                                       // preflight requests
	tb := ginratelimit.NewTokenBucket(_cfg.RequestsPerMinute, 1*time.Minute) // rate limiting
	r.Use(ginratelimit.RateLimitByIP(tb))

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
	r.OPTIONS("/*path", routers.OptionsHandler) // all URLs

	r.GET("/animals", service.GetAnimal)
	r.HEAD("/animals", service.GetAnimalCount)
	r.GET("/animals/:id", service.GetAnimal)
	r.POST("/animals", service.CreateAnimal)
	r.PUT("/animals/:id", service.ReplaceAnimal)
	r.DELETE("/animals/:id", service.DeleteAnimal)
	r.PATCH("/animals/:id/description", service.UpdateAnimalDescription) // change only description field

	// setup database health checking loop every 10 seconds
	go utils.DataBaseHealthPollingLoop(service.PostgresClient, time.Duration(_cfg.DBHeathInterval)*time.Second)
	// run the server
	err := r.Run(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
