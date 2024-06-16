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
			log.Fatal(err)
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
			log.Fatal(err)
		}
	}(clientConn)

	log.Println("TCP connection established. Starting to forward traffic")

	// launch a go-routine to forward traffic
	go func() {
		defer func(clientConn net.Conn) {
			err := clientConn.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(clientConn)
		defer func(destConn net.Conn) {
			err := destConn.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(destConn)
		_, err := io.Copy(destConn, clientConn)
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.Copy(clientConn, destConn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection closed")
}
