package routers

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

func OptionsHandler(c *gin.Context) {
	// success
	c.Status(http.StatusNoContent)
}

func ConnectHandler(c *gin.Context) {
	// parse the destination url
	remote, err := url.Parse("http://" + c.Request.Host)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Host URL"})
		return
	}

	// connecting to the destination server via tcp
	destConn, err := net.Dial("tcp", remote.Host)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to destination"})
		return
	}
	defer func(destConn net.Conn) {
		err := destConn.Close()
		if err != nil {
			// log the error
			c.Error(err)
			// respond with an internal server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close destination connection"})
			return
		}
	}(destConn)

	// make it callers responsibility to close the connection
	clientConn, _, err := c.Writer.Hijack()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to hijack the connection"})
		return
	}
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			// log the error
			c.Error(err)
			// respond with an internal server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close client connection"})
			return
		}
	}(clientConn)

	log.Println("TCP connection established. Starting to forward traffic")

	// launch a go-routine to forward traffic
	go func() {
		defer func(clientConn net.Conn) {
			err := clientConn.Close()
			if err != nil {
				// log the error
				c.Error(err)
				// respond with an internal server error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close client connection"})
				return
			}
		}(clientConn)
		defer func(destConn net.Conn) {
			err := destConn.Close()
			if err != nil {
				// log the error
				c.Error(err)
				// respond with an internal server error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close destination connection"})
				return

			}
		}(destConn)
		_, err := io.Copy(destConn, clientConn)
		if err != nil {
			// log the error
			c.Error(err)
			// respond with an internal server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward traffic"})
			return
		}
	}()

	_, err = io.Copy(clientConn, destConn)
	if err != nil {
		// log the error
		c.Error(err)
		// respond with an internal server error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward traffic"})
		return
	}
	log.Println("Connection closed")
}
