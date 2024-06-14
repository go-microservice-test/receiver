package routers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"go-test/models"
	"log"
	"net/http"
	"sync"
)

func GetAnimals(c *gin.Context) {
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
