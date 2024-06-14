package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
)

// ApiMiddleware - adds db and mutex objects for routers to use
func ApiMiddleware(db *gorm.DB, mu sync.Mutex) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("databaseObject", db)
		c.Set("mutexObject", mu)
		c.Next()
	}
}
