package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type Animal struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Description string `json:"description"`
	isActive    bool   `json:"isactive"`
}

// JSONMap - record processed into json parseable object.
type JSONMap struct {
	ID     int    `json:"id"`
	Animal Animal `json:"data"`
}

var (
	animals   = make(map[int]Animal) // mapping from ID to Animal
	idCounter = 1                    // counter for the ids
	mu        sync.Mutex             // DB mutex
	connStr   = "user=nur password= host=localhost port=5432 dbname=nur sslmode=disable"
	db        *sql.DB
)

func main() {
	// get an engine instance
	r := gin.Default()

	// middleware for connect and trace handlers
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "CONNECT" {
			connectHandler(c)
		} else if c.Request.Method == "TRACE" && c.Request.URL.Path == "/animals" {
			traceHandler(c)
		} else {
			c.Next()
		}
	})

	// connect routes
	r.GET("/animals", getHandler)
	r.HEAD("/animals", headHandler)
	r.GET("/animals/:id", getByIdHandler)
	r.POST("/animals", postHandler)
	r.PUT("/animals/:id", putHandler)
	r.DELETE("/animals/:id", deleteHandler)
	r.OPTIONS("/*path", optionsHandler)               // all URLs
	r.PATCH("/animals/:id/description", patchHandler) // change only description field

	// connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database")
	// creating table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS animals (
    	id SERIAL PRIMARY KEY,
    	name VARCHAR(100),
    	type INT,
    	description VARCHAR(500),
    	isactive BOOLEAN
    )`)
	if err != nil {
		log.Fatal(err)
	}

	// run the server
	r.Run(":3000")
}

func getHandler(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	// convert map into JSON parseable format
	var resAnimalList []JSONMap
	for id, animal := range animals {
		resAnimalList = append(resAnimalList, JSONMap{ID: id, Animal: animal})
	}
	// return all animals
	c.JSON(http.StatusOK, resAnimalList)
}

func headHandler(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	// set the custom item length header to number of records in DB
	c.Header("X-Item-Length", strconv.Itoa(len(animals)))
	c.Status(http.StatusOK)
}

func getByIdHandler(c *gin.Context) {
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// retrieve an animal with id
	animal, ok := animals[id]
	// item does not exist
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}
	// send the requested animal
	c.JSON(http.StatusOK, JSONMap{ID: id, Animal: animal})
}

func postHandler(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var animal Animal
	if err := c.ShouldBindJSON(&animal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// add new animal with the correct id
	animals[idCounter] = animal
	// return created animal
	c.JSON(http.StatusOK, JSONMap{ID: idCounter, Animal: animal})
	// update the counter
	idCounter++
}

func putHandler(c *gin.Context) {
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var animal Animal
	if err := c.ShouldBindJSON(&animal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if id exists
	_, ok := animals[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}

	// update animal from id (here created anew)
	animals[id] = animal
	// return updated animal
	c.JSON(http.StatusOK, JSONMap{ID: id, Animal: animal})
}

func deleteHandler(c *gin.Context) {
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// check if id exists
	_, ok := animals[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}

	// delete an animal item
	delete(animals, id)
	c.Status(http.StatusNoContent)
}

func optionsHandler(c *gin.Context) {
	// setting cors headers
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH, CONNECT, TRACE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	c.Header("Access-Control-Max-Age", "86400")

	// success
	c.Status(http.StatusNoContent)
}

func patchHandler(c *gin.Context) {
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var input struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if id exists
	oldAnimal, ok := animals[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}

	// update animal from id by concatenating input
	var updatedAnimal = Animal{
		Name:        oldAnimal.Name,
		Type:        oldAnimal.Type,
		Description: input.Description, // change the description
		isActive:    oldAnimal.isActive,
	}
	animals[id] = updatedAnimal
	// return updated animal
	c.JSON(http.StatusOK, JSONMap{ID: id, Animal: updatedAnimal})
}

func connectHandler(c *gin.Context) {
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
	defer destConn.Close()

	// make it callers responsibility to close the connection
	clientConn, _, err := c.Writer.Hijack()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to hijack the connection"})
		return
	}
	defer clientConn.Close()

	log.Println("TCP connection established. Starting to forward traffic")

	// launch a go-routine to forward traffic
	go func() {
		defer clientConn.Close()
		defer destConn.Close()
		io.Copy(destConn, clientConn)
	}()

	io.Copy(clientConn, destConn)
	log.Println("Connection closed")
}

func traceHandler(c *gin.Context) {
	var animal Animal
	if err := c.ShouldBindJSON(&animal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// correct header
	c.Header("Content-Type", "message/http")
	// send processed proxy list
	c.Header("Via", c.GetHeader("Via"))
	// send body as is
	c.JSON(http.StatusOK, animal)
}
