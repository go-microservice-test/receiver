package middleware

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"sync"
)

// ApiMiddleware - adds db and mutex objects for routers to use
func ApiMiddleware(db *sql.DB, mu sync.Mutex) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("databaseObject", db)
		c.Set("mutexObject", mu)
		c.Next()
	}
}
