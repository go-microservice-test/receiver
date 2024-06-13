package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// Item - some abstract data holding type.
type Item string

// JSONItem - map item processed into JSON parseable object.
type JSONItem struct {
	ID   int  `json:"id"`
	Item Item `json:"data"`
}

// JSONItemInput - item input json format
type JSONItemInput struct {
	Data string `json:"data"`
}

var (
	items     = make(map[int]Item) // mapping from ID to Item
	idCounter = 1                  // counter for the ids
	mu        sync.Mutex           // DB mutex
)

func main() {
	// get an engine instance
	r := gin.Default()

	// middleware for connect and trace handlers
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "CONNECT" {
			connectHandler(c)
		} else if c.Request.Method == "TRACE" && c.Request.URL.Path == "/items" {
			traceHandler(c)
		} else {
			c.Next()
		}
	})

	// connect routes
	r.GET("/items", getItems)
	r.HEAD("/items", getItemsLength)
	r.GET("/items/:id", getItem)
	r.POST("/items", createItem)
	r.PUT("/items/:id", updateItem)
	r.DELETE("/items/:id", deleteItem)
	r.OPTIONS("/*path", optionsHandler) // all URLs
	r.PATCH("/items/:id", patchItem)

	// run the server
	r.Run(":3000")
}

func getItems(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	// convert map into JSON parseable format
	var resItemList []JSONItem
	for id, item := range items {
		resItemList = append(resItemList, JSONItem{ID: id, Item: item})
	}
	// return all items
	c.JSON(http.StatusOK, resItemList)
}

func getItemsLength(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	// set the custom item length header to item length
	c.Header("X-Item-Length", strconv.Itoa(len(items)))
	c.Status(http.StatusOK)
}

func getItem(c *gin.Context) {
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// retrieve an item with id
	item, ok := items[id]
	// item does not exist
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}
	// send the requested item
	c.JSON(http.StatusOK, JSONItem{ID: id, Item: item})
}

func createItem(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var itemInput JSONItemInput
	if err := c.ShouldBindJSON(&itemInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// add new item with the correct id
	var item = Item(itemInput.Data)
	items[idCounter] = item
	// return created item
	c.JSON(http.StatusOK, JSONItem{ID: idCounter, Item: item})
	// update the counter
	idCounter++
}

func updateItem(c *gin.Context) {
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
	var itemInput JSONItemInput
	if err := c.ShouldBindJSON(&itemInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if id exists
	_, ok := items[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// update item from id (here created anew)
	var updatedItem = Item(itemInput.Data)
	items[id] = updatedItem
	// return updated item
	c.JSON(http.StatusOK, JSONItem{ID: id, Item: updatedItem})
}

func deleteItem(c *gin.Context) {
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
	_, ok := items[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// delete an item
	delete(items, id)
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

func patchItem(c *gin.Context) {
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
	var itemInput JSONItemInput
	if err := c.ShouldBindJSON(&itemInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if id exists
	oldItem, ok := items[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// update item from id by concatenating input
	var updatedItem = Item(string(oldItem) + itemInput.Data)
	items[id] = updatedItem
	// return updated item
	c.JSON(http.StatusOK, JSONItem{ID: id, Item: updatedItem})
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
	var itemInput JSONItemInput
	if err := c.ShouldBindJSON(&itemInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// correct header
	c.Header("Content-Type", "message/http")
	// send processed proxy list
	c.Header("Via", c.GetHeader("Via"))
	// send body as is
	c.JSON(http.StatusOK, itemInput)
}
