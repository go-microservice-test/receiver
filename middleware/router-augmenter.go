package middleware

import (
	"github.com/gin-gonic/gin"
	"go-test/db-utils/repository"
	"sync"
)

// ApiMiddleware - adds db and mutex objects for routers to use
func ApiMiddleware(mu sync.Mutex, rp repository.AnimalRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("mutexObject", mu)
		c.Set("animalRepository", rp)
		c.Next()
	}
}
