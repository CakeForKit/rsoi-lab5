package service

import (
	"context"
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/custom_error"
	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	"github.com/CakeForKit/rsoi-lab2/ms-ticket/internal/repository"
	"github.com/google/uuid"
)

var ticketSrv TicketService
var ticketSrvMutex sync.Mutex

type TicketService interface {
	Create(ctx context.Context, username string, request model.TicketCreateRequest) (model.Ticket, error)
	Delete(ctx context.Context, username string, uid string) error
	GetAll(ctx context.Context, username string) ([]model.Ticket, error)
	Get(ctx context.Context, username string, uid string) (model.Ticket, error)
	Cancel(ctx context.Context, username string, uid string) (model.Ticket, error)
}

// Delete is used by the gateway to compensate a ticket creation when the
// subsequent bonus operation cannot be completed.
func (service *ticketService) Delete(ctx context.Context, username string, uid string) error {
	ticket, err := service.Get(ctx, username, uid)
	if err != nil {
		return err
	}
	return service.repository.DeleteById(ctx, []uint{ticket.ID})
}

type ticketService struct {
	repository repository.TicketRepository
}

func GetTicketService() (TicketService, error) {
	ticketSrvMutex.Lock()
	defer ticketSrvMutex.Unlock()

	if ticketSrv != nil {
		return ticketSrv, nil
	}

	repo, err := repository.GetTicketRepository()
	if err != nil {
		return nil, err
	}

	ticketSrv = &ticketService{repository: repo}
	return ticketSrv, nil
}

func (service *ticketService) Create(ctx context.Context, username string, request model.TicketCreateRequest) (model.Ticket, error) {
	tickets, err := service.repository.Create(ctx, []model.Ticket{{
		TicketUID:    uuid.New(),
		Username:     username,
		FlightNumber: request.FlightNumber,
		Price:        request.Price,
		Status:       "PAID",
	}})
	if err != nil {
		return model.Ticket{}, err
	}
	return tickets[0], nil
}

func (service *ticketService) GetAll(ctx context.Context, username string) ([]model.Ticket, error) {
	tickets, err := service.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]model.Ticket, 0)
	for _, ticket := range tickets {
		if ticket.Username == username {
			result = append(result, ticket)
		}
	}
	return result, nil
}

func (service *ticketService) Get(ctx context.Context, username string, uid string) (model.Ticket, error) {
	ticketUID, err := uuid.Parse(uid)
	if err != nil {
		return model.Ticket{}, custom_error.NotFoundError("ticket")
	}
	tickets, err := service.GetAll(ctx, username)
	if err != nil {
		return model.Ticket{}, err
	}
	for _, ticket := range tickets {
		if ticket.TicketUID == ticketUID {
			return ticket, nil
		}
	}
	return model.Ticket{}, custom_error.NotFoundError("ticket")
}

func (service *ticketService) Cancel(ctx context.Context, username string, uid string) (model.Ticket, error) {
	ticket, err := service.Get(ctx, username, uid)
	if err != nil {
		return model.Ticket{}, err
	}
	if ticket.Status == "CANCELED" {
		return model.Ticket{}, custom_error.IllegalArgumentError("ticket already canceled")
	}
	ticket.Status = "CANCELED"
	tickets, err := service.repository.Update(ctx, []model.Ticket{ticket})
	if err != nil {
		return model.Ticket{}, err
	}
	return tickets[0], nil
}
