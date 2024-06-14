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

func GetAnimalByID(c *gin.Context) {
	db, mu := retrieveObjects(c)
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// select records with specific id
	query := `SELECT * FROM animals WHERE id = $1`
	rows, err := db.Query(query, id)
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	// check if there are any rows returned
	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	} else {
		// return specific animal
		var animal models.Animal
		// parse animal
		err := rows.Scan(&id, &animal.Name, &animal.Type, &animal.Description, &animal.IsActive)
		if err != nil {
			log.Fatal(err)
		}
		// send the requested animal
		c.JSON(http.StatusOK, models.AnimalWithID{ID: id, Animal: animal})
		return
	}
}

func CreateAnimal(c *gin.Context) {
	db, mu := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var animal models.Animal
	if err := c.ShouldBindJSON(&animal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// prepare sql statement
	query := `
        INSERT INTO animals (name, type, description, is_active)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	var insertedID int
	err := db.QueryRow(query, animal.Name, animal.Type, animal.Description, animal.IsActive).Scan(&insertedID)
	if err != nil {
		log.Fatal(err)
	}

	// return created animal
	c.JSON(http.StatusOK, models.AnimalWithID{ID: insertedID, Animal: animal})
}

func ReplaceAnimal(c *gin.Context) {
	db, mu := retrieveObjects(c)
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	// incorrect input format handling
	var animal models.Animal
	if err := c.ShouldBindJSON(&animal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	query := `
		UPDATE animals
		SET name = $2, type = $3, description = $4, is_active = $5
		WHERE id = $1
	`

	// execute with parameters
	result, err := db.Exec(query, id, animal.Name, animal.Type, animal.Description, animal.IsActive)
	if err != nil {
		log.Fatal(err)
	}
	// check rowsAffected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	} else {
		c.JSON(http.StatusOK, models.AnimalWithID{ID: id, Animal: animal})
	}
}

func DeleteAnimal(c *gin.Context) {
	db, mu := retrieveObjects(c)
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	query := `DELETE FROM animals WHERE id = $1`

	// execute with parameters
	result, err := db.Exec(query, id)
	if err != nil {
		log.Fatal(err)
	}

	// check rowsAffected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	} else {
		c.Status(http.StatusNoContent)
	}
}

func UpdateAnimalDescription(c *gin.Context) {
	db, mu := retrieveObjects(c)
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}

	// incorrect input format handling
	var input struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	query := `
			UPDATE animals
			SET description = $2
			WHERE id = $1
		`

	// execute with parameters
	result, err := db.Exec(query, id, input.Description)
	if err != nil {
		log.Fatal(err)
	}
	// check rowsAffected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	} else {
		c.Status(http.StatusOK)
	}
}

func TraceAnimalRoute(c *gin.Context) {
	var animal models.Animal
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
