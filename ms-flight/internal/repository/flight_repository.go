package repository

import (
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	coreRepository "github.com/CakeForKit/rsoi-lab2/lb-core/repository"
)

var flightRepo *flightRepository
var flightRepoMutex sync.Mutex

type FlightRepository interface {
	coreRepository.Repository[model.Flight]
}

type flightRepository struct {
	coreRepository.Repository[model.Flight]
}

func GetFlightRepository() (FlightRepository, error) {
	flightRepoMutex.Lock()
	defer flightRepoMutex.Unlock()

	if flightRepo != nil {
		return flightRepo, nil
	}

	repository, err := coreRepository.NewPostgresRepository[model.Flight]()
	if err != nil {
		return nil, err
	}

	flightRepo = &flightRepository{repository}
	return flightRepo, nil
}
