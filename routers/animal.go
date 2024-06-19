package routers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go-test/db-utils/repository"
	"go-test/models"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func retrieveObjects(c *gin.Context) (sync.Mutex, repository.AnimalRepository) {
	// retrieve middleware mutex object and repository implementation

	// ideally this should communicate with result channels in handlers
	// so that for example response is not sent twice
	var mu sync.Mutex
	var rp repository.AnimalRepository
	var ok bool
	mu, ok = c.MustGet("mutexObject").(sync.Mutex)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve mutex object"})
	}
	rp, ok = c.MustGet("animalRepository").(repository.AnimalRepository)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve repository implementation"})
	}
	return mu, rp
}

func GetAnimals(c *gin.Context) {
	// setting various timeouts for different handlers
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// channels to receive the result
	resultChan := make(chan []models.AnimalWithID, 1)

	// launch fetching in go-routine
	go func() {
		mu, rp := retrieveObjects(c)
		mu.Lock()
		defer mu.Unlock()

		// select all records from the animals table
		var animals, err = rp.FindAll()
		if err != nil {
			log.Fatal(err)
		}
		// convert results into JSON parseable format
		var resAnimalList []models.AnimalWithID
		for _, animal := range animals {
			resAnimalList = append(resAnimalList, models.AnimalWithID{
				ID: int(animal.ID),
				Animal: models.Animal{
					Name:        animal.Name,
					Type:        animal.Type,
					Description: animal.Description,
				},
			})
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
	mu, rp := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// get count from the animals table
	count, err := rp.GetCount()
	if err != nil {
		log.Fatal(err)
	}
	// set the custom item length header to number of records in DB
	c.Header("X-Item-Length", strconv.Itoa(int(count)))
	c.Status(http.StatusOK)
}

func GetAnimalByID(c *gin.Context) {
	mu, rp := retrieveObjects(c)
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	animal, err := rp.FindByID(uint(id))
	if err != nil {
		var notFound *repository.NotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		}
		log.Fatal(err)
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
	mu, rp := retrieveObjects(c)
	mu.Lock()
	defer mu.Unlock()

	// incorrect input format handling
	var animalInput models.Animal
	if err := c.ShouldBindJSON(&animalInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	animal, err := rp.Create(animalInput)
	if err != nil {
		log.Fatal(err)
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
	mu, rp := retrieveObjects(c)
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

	animal, err := rp.Replace(uint(id), animalInput)
	if err != nil {
		var notFound *repository.NotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		}
		log.Fatal(err)
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
	mu, rp := retrieveObjects(c)
	// retrieving URL id param
	id, err := strconv.Atoi(c.Param("id"))
	// invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}
	mu.Lock()
	defer mu.Unlock()

	animal, err := rp.Delete(uint(id))
	if err != nil {
		var notFound *repository.NotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		}
		log.Fatal(err)
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
	mu, rp := retrieveObjects(c)
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

	animal, err := rp.UpdateDescription(uint(id), input.Description)
	if err != nil {
		var notFound *repository.NotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Animal not found"})
			return
		}
		log.Fatal(err)
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
