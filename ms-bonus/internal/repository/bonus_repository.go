package repository

import (
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	coreRepository "github.com/CakeForKit/rsoi-lab2/lb-core/repository"
)

var bonusRepo *bonusRepository
var bonusRepoMutex sync.Mutex

type BonusRepository interface {
	coreRepository.Repository[model.Privilege]
}

type bonusRepository struct {
	coreRepository.Repository[model.Privilege]
}

func GetBonusRepository() (BonusRepository, error) {
	bonusRepoMutex.Lock()
	defer bonusRepoMutex.Unlock()

	if bonusRepo != nil {
		return bonusRepo, nil
	}

	repository, err := coreRepository.NewPostgresRepository[model.Privilege]()
	if err != nil {
		return nil, err
	}

	bonusRepo = &bonusRepository{repository}
	return bonusRepo, nil
}
