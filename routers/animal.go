package routers

import (
	"errors"
	"github.com/gin-gonic/gin"
	models2 "go-test/db-utils/models"
	"go-test/models"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"sync"
)

func retrieveObjects(c *gin.Context) (*gorm.DB, sync.Mutex) {
	// retrieve middleware db and mutex objects
	var db *gorm.DB
	var mu sync.Mutex
	var ok bool
	db, ok = c.MustGet("databaseObject").(*gorm.DB)
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
	var animals []models2.Animal
	result := db.Find(&animals)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	// convert results into JSON parseable format
	var resAnimalList []models.AnimalWithID
	for _, animal := range animals {
		// append to the list
		resAnimalList = append(resAnimalList, models.AnimalWithID{
			ID: int(animal.ID),
			Animal: models.Animal{
				Name:        animal.Name,
				Type:        animal.Type,
				Description: animal.Description,
				IsActive:    animal.IsActive,
			},
		})
	}
	// return all animals
	c.JSON(http.StatusOK, resAnimalList)
}

func GetAnimalCount(c *gin.Context) {
	db, mu := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// get count from the animals table
	var count int64
	result := db.Model(&models2.Animal{}).Count(&count)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	// set the custom item length header to number of records in DB
	c.Header("X-Item-Length", strconv.Itoa(int(count)))
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

	// find first record with id
	var animal models2.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}
	// send the requested animal
	c.JSON(http.StatusOK, models.AnimalWithID{
		ID: id,
		Animal: models.Animal{
			Name:        animal.Name,
			Type:        animal.Type,
			Description: animal.Description,
			IsActive:    animal.IsActive,
		},
	})
	return
}

func CreateAnimal(c *gin.Context) {
	db, mu := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var animalInput models.Animal
	if err := c.ShouldBindJSON(&animalInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// copy fields from input
	var animal models2.Animal
	animal.Name = animalInput.Name
	animal.Description = animalInput.Description
	animal.Type = animalInput.Type
	animal.IsActive = animalInput.IsActive
	// create a new record
	result := db.Create(&animal)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	// return created animal
	c.JSON(http.StatusOK, models.AnimalWithID{
		ID: int(animal.ID),
		Animal: models.Animal{
			Name:        animal.Name,
			Type:        animal.Type,
			Description: animal.Description,
			IsActive:    animal.IsActive,
		},
	})
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
	var animalInput models.Animal
	if err := c.ShouldBindJSON(&animalInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	var animal models2.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}

	// replace all field values
	animal.Name = animalInput.Name
	animal.Description = animalInput.Description
	animal.Type = animalInput.Type
	animal.IsActive = animalInput.IsActive

	result = db.Save(&animal)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	c.JSON(http.StatusOK, models.AnimalWithID{
		ID: int(animal.ID),
		Animal: models.Animal{
			Name:        animal.Name,
			Type:        animal.Type,
			Description: animal.Description,
			IsActive:    animal.IsActive,
		},
	})
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

	// delete in the database
	result := db.Delete(&models2.Animal{}, id)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if result.RowsAffected == 0 {
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

	var animal models2.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}

	// replace description
	animal.Description = input.Description

	result = db.Save(&animal)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	c.JSON(http.StatusOK, models.AnimalWithID{
		ID: int(animal.ID),
		Animal: models.Animal{
			Name:        animal.Name,
			Type:        animal.Type,
			Description: animal.Description,
			IsActive:    animal.IsActive,
		},
	})
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
