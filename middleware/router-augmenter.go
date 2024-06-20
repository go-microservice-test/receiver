package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go-test/db-utils/repository"
	"sync"
)

// ApiMiddleware - adds db and mutex objects for routers to use
func ApiMiddleware(mu sync.Mutex, rp repository.AnimalRepository, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("mutexObject", mu)
		c.Set("animalRepository", rp)
		c.Set("cacheObject", rdb)
		c.Next()
	}
}
