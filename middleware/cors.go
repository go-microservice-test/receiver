package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH, CONNECT, TRACE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Via")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		c.Next()
	}
}
