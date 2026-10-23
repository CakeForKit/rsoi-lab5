package repository

import (
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	coreRepository "github.com/CakeForKit/rsoi-lab2/lb-core/repository"
)

var bonusHistoryRepo *bonusHistoryRepository
var bonusHistoryRepoMutex sync.Mutex

type BonusHistoryRepository interface {
	coreRepository.Repository[model.PrivilegeHistory]
}

type bonusHistoryRepository struct {
	coreRepository.Repository[model.PrivilegeHistory]
}

func GetBonusHistoryRepository() (BonusHistoryRepository, error) {
	bonusHistoryRepoMutex.Lock()
	defer bonusHistoryRepoMutex.Unlock()

	if bonusHistoryRepo != nil {
		return bonusHistoryRepo, nil
	}

	repository, err := coreRepository.NewPostgresRepository[model.PrivilegeHistory]()
	if err != nil {
		return nil, err
	}

	bonusHistoryRepo = &bonusHistoryRepository{repository}
	return bonusHistoryRepo, nil
}
