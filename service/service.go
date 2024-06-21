package service

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	dbutils "go-test/db-utils"
	"go-test/db-utils/repository"
	"go-test/routers"
	"go-test/utils"
	"gorm.io/gorm"
	"sync"
)

type Service struct {
	Config         *utils.Config
	Repository     *repository.AnimalRepository
	PostgresClient *gorm.DB
	RedisClient    *redis.Client
	Mutex          *sync.Mutex
}

func NewService(config *utils.Config) *Service {
	// DB mutex
	var mu sync.Mutex
	// setup db connection
	db := dbutils.Connect(config.DBUser, config.DBPassword, config.DBHost, config.DBName, config.DBSSLMode, config.DBPort)
	// setup cache connection
	rdb := dbutils.ConnectRedis(config.RedisAddress, config.RedisAddress, config.RedisDB)
	// setup repositories
	animalRepository := repository.NewAnimalsRepositoryImpl(db)
	return &Service{
		Config:         config,
		Repository:     &animalRepository,
		RedisClient:    rdb,
		PostgresClient: db,
		Mutex:          &mu,
	}
}

func (service *Service) GetAnimal(c *gin.Context) {
	routers.GetAnimals(c, service.Mutex, service.Repository)
}

func (service *Service) GetAnimalCount(c *gin.Context) {
	routers.GetAnimalCount(c, service.Mutex, service.Repository)
}

func (service *Service) GetAnimalById(c *gin.Context) {
	routers.GetAnimalByID(c, service.Mutex, service.Repository, service.RedisClient)
}

func (service *Service) CreateAnimal(c *gin.Context) {
	routers.CreateAnimal(c, service.Mutex, service.Repository)
}

func (service *Service) ReplaceAnimal(c *gin.Context) {
	routers.ReplaceAnimal(c, service.Mutex, service.Repository, service.RedisClient)
}

func (service *Service) DeleteAnimal(c *gin.Context) {
	routers.DeleteAnimal(c, service.Mutex, service.Repository, service.RedisClient)
}

func (service *Service) UpdateAnimalDescription(c *gin.Context) {
	routers.UpdateAnimalDescription(c, service.Mutex, service.Repository, service.RedisClient)
}
