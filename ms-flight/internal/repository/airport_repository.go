package repository

import (
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	coreRepository "github.com/CakeForKit/rsoi-lab2/lb-core/repository"
)

var airportRepo *airportRepository
var airportRepoMutex sync.Mutex

type AirportRepository interface {
	coreRepository.Repository[model.Airport]
}

type airportRepository struct {
	coreRepository.Repository[model.Airport]
}

func GetAirportRepository() (AirportRepository, error) {
	airportRepoMutex.Lock()
	defer airportRepoMutex.Unlock()

	if airportRepo != nil {
		return airportRepo, nil
	}

	repository, err := coreRepository.NewPostgresRepository[model.Airport]()
	if err != nil {
		return nil, err
	}

	airportRepo = &airportRepository{repository}
	return airportRepo, nil
}
