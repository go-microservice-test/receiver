package routers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	dbmodels "go-test/db-utils/models"
	"go-test/models"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func retrieveObjects(c *gin.Context) (*gorm.DB, sync.Mutex) {
	// retrieve middleware db and mutex objects

	// ideally this should communicate with result channels in handlers
	// so that for example response is not sent twice
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
	// setting various timeouts for different handlers
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// channels to receive the result
	resultChan := make(chan []models.AnimalWithID, 1)

	// launch fetching in go-routine
	go func() {
		db, mu := retrieveObjects(c)
		mu.Lock()
		defer mu.Unlock()

		// select all records from the animals table
		var animals []dbmodels.Animal
		result := db.Find(&animals)
		if result.Error != nil {
			log.Fatal(result.Error)
		}
		// convert results into JSON parseable format
		var resAnimalList []models.AnimalWithID
		for _, animal := range animals {
			// append to the list if not deleted
			if animal.IsActive {
				resAnimalList = append(resAnimalList, models.AnimalWithID{
					ID: int(animal.ID),
					Animal: models.Animal{
						Name:        animal.Name,
						Type:        animal.Type,
						Description: animal.Description,
					},
				})
			}
		}
		// exit on normal execution or on timeout
		select {
		case resultChan <- resAnimalList:
		case <-ctx.Done():
		}
	}()
	select {
	case res := <-resultChan:
		c.JSON(http.StatusOK, res)
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request timeout"})
	}
}

func GetAnimalCount(c *gin.Context) {
	db, mu := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// get count from the animals table
	var count int64
	result := db.Model(&dbmodels.Animal{}).Where("is_active = ?", true).Count(&count)
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
	var animal dbmodels.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}
	// check if deleted
	if !animal.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}
	// send the requested animal
	c.JSON(http.StatusOK, models.AnimalWithID{
		ID: id,
		Animal: models.Animal{
			Name:        animal.Name,
			Type:        animal.Type,
			Description: animal.Description,
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
	var animal dbmodels.Animal
	animal.Name = animalInput.Name
	animal.Description = animalInput.Description
	animal.Type = animalInput.Type

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

	var animal dbmodels.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}
	// check if deleted
	if !animal.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}

	// replace all field values
	animal.Name = animalInput.Name
	animal.Description = animalInput.Description
	animal.Type = animalInput.Type

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

	// check if exists
	var animal dbmodels.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}
	// check if deleted
	if !animal.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
	}
	// set deleted flag
	animal.IsActive = false
	// update in the database
	result = db.Save(&animal)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	// send deleted animal
	c.JSON(http.StatusOK, models.AnimalWithID{
		ID: int(animal.ID),
		Animal: models.Animal{
			Name:        animal.Name,
			Type:        animal.Type,
			Description: animal.Description,
		},
	})
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

	var animal dbmodels.Animal
	result := db.First(&animal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		} else {
			log.Fatal(result.Error)
		}
	}
	// check if deleted
	if !animal.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
		return
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
