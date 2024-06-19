package mocks

import (
	"github.com/stretchr/testify/mock"
	"go-test/db-utils/models"
	inputModels "go-test/models"
)

// MockRepository - mock repository implementation
type MockRepository struct {
	mock.Mock
}

// mock methods to satisfy interface

func (m *MockRepository) FindAll() ([]models.Animal, error) {
	args := m.Called()
	return args.Get(0).([]models.Animal), args.Error(1)
}

func (m *MockRepository) GetCount() (int64, error) {
	return 0, nil
}

func (m *MockRepository) FindByID(id uint) (models.Animal, error) {
	return models.Animal{}, nil
}

func (m *MockRepository) Create(animal inputModels.Animal) (models.Animal, error) {
	return models.Animal{}, nil
}

func (m *MockRepository) Replace(id uint, animal inputModels.Animal) (models.Animal, error) {
	return models.Animal{}, nil
}

func (m *MockRepository) Delete(id uint) (models.Animal, error) {
	return models.Animal{}, nil
}

func (m *MockRepository) UpdateDescription(id uint, description string) (models.Animal, error) {
	return models.Animal{}, nil
}
