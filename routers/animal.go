package routers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"go-test/models"
	"log"
	"net/http"
	"strconv"
	"sync"
)

func retrieveObjects(c *gin.Context) (*sql.DB, sync.Mutex) {
	// retrieve middleware db and mutex objects
	var db *sql.DB
	var mu sync.Mutex
	var ok bool
	db, ok = c.MustGet("databaseObject").(*sql.DB)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve database object"})
	}
	mu, ok = c.MustGet("mutexObject").(sync.Mutex)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve mutex object"})
	}
	return db, mu
}

func GetAnimals(c *gin.Context) {
	db, mu := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// select all records from the animals table
	rows, err := db.Query("SELECT * FROM animals")
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)
	// convert results into JSON parseable format
	var resAnimalList []models.AnimalWithID
	for rows.Next() {
		var id int
		var animal models.Animal
		// parse it into id and animal
		err := rows.Scan(&id, &animal.Name, &animal.Type, &animal.Description, &animal.IsActive)
		if err != nil {
			log.Fatal(err)
		}

		// append to the list
		resAnimalList = append(resAnimalList, models.AnimalWithID{ID: id, Animal: animal})
	}
	// return all animals
	c.JSON(http.StatusOK, resAnimalList)
}

func GetAnimalCount(c *gin.Context) {
	db, mu := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// get count from the animals table
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM animals").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	// set the custom item length header to number of records in DB
	c.Header("X-Item-Length", strconv.Itoa(count))
	c.Status(http.StatusOK)
}
