package repository

import (
	"errors"
	"fmt"
	"go-test/db-utils/models"
	inputModels "go-test/models"
	"gorm.io/gorm"
	"time"
)

type AnimalRepository interface {
	FindAll() ([]models.Animal, error)
	GetCount() (int64, error)
	FindByID(id uint) (models.Animal, error)
	Create(animal inputModels.Animal) (models.Animal, error)
	Replace(id uint, animal inputModels.Animal) (models.Animal, error)
	Delete(id uint) (models.Animal, error)
	UpdateDescription(id uint, description string) (models.Animal, error)
}

type NotFoundError struct {
	When time.Time
	Id   uint
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Id Not Found: %d at %v",
		e.Id, e.When)
}

type AnimalRepositoryImpl struct {
	db *gorm.DB
}

func NewAnimalsRepositoryImpl(DB *gorm.DB) AnimalRepository {
	return &AnimalRepositoryImpl{db: DB}
}

func (a *AnimalRepositoryImpl) FindAll() ([]models.Animal, error) {
	var animals []models.Animal
	result := a.db.Find(&animals)
	if result.Error != nil {
		return animals, result.Error
	}
	// filter out deleted animals
	isDeleted := func(animal models.Animal) bool {
		return !animal.IsActive
	}
	filteredAnimals := []models.Animal{}
	for _, animal := range animals {
		if !isDeleted(animal) {
			filteredAnimals = append(filteredAnimals, animal)
		}
	}
	return filteredAnimals, nil
}

func (a *AnimalRepositoryImpl) GetCount() (int64, error) {
	var count int64
	result := a.db.Model(&models.Animal{}).Where("is_active = ?", true).Count(&count)
	if result.Error != nil {
		return -1, result.Error
	}
	return count, nil
}

func (a *AnimalRepositoryImpl) FindByID(id uint) (models.Animal, error) {
	var animal models.Animal
	// find first record with id
	result := a.db.First(&animal, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return animal, &NotFoundError{Id: id, When: time.Now()}
		} else {
			return animal, result.Error
		}
	}
	// check if deleted
	if !animal.IsActive {
		return animal, &NotFoundError{Id: id, When: time.Now()}
	}
	return animal, nil
}

func (a *AnimalRepositoryImpl) Create(animalInput inputModels.Animal) (models.Animal, error) {
	// set exactly those fields which are needed
	var animal models.Animal
	animal.Name = animalInput.Name
	animal.Description = animalInput.Description
	animal.Type = animalInput.Type
	// create in the DB
	result := a.db.Create(&animal)
	if result.Error != nil {
		return animal, result.Error
	}
	return animal, nil
}

func (a *AnimalRepositoryImpl) Replace(id uint, animalInput inputModels.Animal) (models.Animal, error) {
	var animal models.Animal
	result := a.db.First(&animal, id)
	// find animal by id
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return animal, &NotFoundError{Id: id, When: time.Now()}
		} else {
			return animal, result.Error
		}
	}
	if !animal.IsActive {
		return animal, &NotFoundError{Id: id, When: time.Now()}
	}
	// replace needed field values
	animal.Name = animalInput.Name
	animal.Description = animalInput.Description
	animal.Type = animalInput.Type
	// save replaced animal
	result = a.db.Save(&animal)
	if result.Error != nil {
		return animal, result.Error
	}
	return animal, nil
}

func (a *AnimalRepositoryImpl) Delete(id uint) (models.Animal, error) {
	var animal models.Animal
	result := a.db.First(&animal, id)
	// find animal by id
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return animal, &NotFoundError{Id: id, When: time.Now()}
		} else {
			return animal, result.Error
		}
	}
	if !animal.IsActive {
		return animal, &NotFoundError{Id: id, When: time.Now()}
	}
	// set him to deleted state
	animal.IsActive = false
	// apply changes
	result = a.db.Save(&animal)
	if result.Error != nil {
		return animal, result.Error
	}
	return animal, nil
}

func (a *AnimalRepositoryImpl) UpdateDescription(id uint, description string) (models.Animal, error) {
	var animal models.Animal
	result := a.db.First(&animal, id)
	// find animal by id
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return animal, &NotFoundError{Id: id, When: time.Now()}
		} else {
			return animal, result.Error
		}
	}
	if !animal.IsActive {
		return animal, &NotFoundError{Id: id, When: time.Now()}
	}
	// update his description
	animal.Description = description
	// apply changes
	result = a.db.Save(&animal)
	if result.Error != nil {
		return animal, result.Error
	}
	return animal, nil
}
