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
	r.PUT("/items/:id", updateItem)
	r.DELETE("/items/:id", deleteItem)

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
	c.Status(http.StatusOK)
}
