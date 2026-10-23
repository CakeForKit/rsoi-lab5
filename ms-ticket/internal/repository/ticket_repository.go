package repository

import (
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	coreRepository "github.com/CakeForKit/rsoi-lab2/lb-core/repository"
)

var ticketRepo *ticketRepository
var ticketRepoMutex sync.Mutex

type TicketRepository interface {
	coreRepository.Repository[model.Ticket]
}

type ticketRepository struct {
	coreRepository.Repository[model.Ticket]
}

func GetTicketRepository() (TicketRepository, error) {
	ticketRepoMutex.Lock()
	defer ticketRepoMutex.Unlock()

	if ticketRepo != nil {
		return ticketRepo, nil
	}

	repository, err := coreRepository.NewPostgresRepository[model.Ticket]()
	if err != nil {
		return nil, err
	}

	ticketRepo = &ticketRepository{repository}
	return ticketRepo, nil
}
