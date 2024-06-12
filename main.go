package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
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

	// connect routes
	r.GET("/items", getItems)
	r.HEAD("/items", getItemsLength)
	r.GET("/items/:id", getItem)
	r.POST("/items", createItem)

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

	var item JSONItemInput
	// incorrect input format handling
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// add new item with the correct id
	var processedItem = Item(item.Data)
	items[idCounter] = processedItem
	// return created item
	c.JSON(http.StatusOK, JSONItem{ID: idCounter, Item: processedItem})
	// update the counter
	idCounter++
}
