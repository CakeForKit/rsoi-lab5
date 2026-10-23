package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	"github.com/CakeForKit/rsoi-lab2/ms-bonus/internal/repository"
)

var bonusSrv BonusService
var bonusSrvMutex sync.Mutex

type BonusService interface {
	Get(ctx context.Context, username string) (model.PrivilegeInfoResponse, error)
	Apply(ctx context.Context, username string, operation model.PrivilegeOperationRequest) (model.PrivilegeShortInfo, error)
}

type bonusService struct {
	bonusRepository   repository.BonusRepository
	historyRepository repository.BonusHistoryRepository
}

func GetBonusService() (BonusService, error) {
	bonusSrvMutex.Lock()
	defer bonusSrvMutex.Unlock()

	if bonusSrv != nil {
		return bonusSrv, nil
	}

	bonusRepository, err := repository.GetBonusRepository()
	if err != nil {
		return nil, err
	}
	historyRepository, err := repository.GetBonusHistoryRepository()
	if err != nil {
		return nil, err
	}

	bonusSrv = &bonusService{bonusRepository: bonusRepository, historyRepository: historyRepository}
	return bonusSrv, nil
}

func (service *bonusService) Get(ctx context.Context, username string) (model.PrivilegeInfoResponse, error) {
	bonus, err := service.getBonus(ctx, username)
	if err != nil {
		return model.PrivilegeInfoResponse{}, err
	}
	history, err := service.historyRepository.GetAll(ctx)
	if err != nil {
		return model.PrivilegeInfoResponse{}, err
	}

	result := make([]model.PrivilegeHistory, 0)
	for _, item := range history {
		if item.PrivilegeID == bonus.ID {
			result = append(result, item)
		}
	}
	return model.PrivilegeInfoResponse{Balance: bonus.Balance, Status: bonus.Status, History: result}, nil
}

func (service *bonusService) Apply(ctx context.Context, username string, operation model.PrivilegeOperationRequest) (model.PrivilegeShortInfo, error) {
	if username == "" || operation.Amount < 0 {
		return model.PrivilegeShortInfo{}, fmt.Errorf("invalid bonus operation")
	}
	bonus, err := service.getBonus(ctx, username)
	if err != nil {
		return model.PrivilegeShortInfo{}, err
	}

	difference := operation.Amount
	if operation.OperationType == model.PrivilegeOperationDebit {
		difference = -operation.Amount
	}
	if bonus.Balance+difference < 0 {
		return model.PrivilegeShortInfo{}, fmt.Errorf("insufficient bonus balance")
	}
	bonus.Balance += difference
	bonus.Status = bonusStatus(bonus.Balance)
	if _, err = service.bonusRepository.Update(ctx, []model.Privilege{bonus}); err != nil {
		return model.PrivilegeShortInfo{}, err
	}
	if operation.Amount > 0 {
		_, err = service.historyRepository.Create(ctx, []model.PrivilegeHistory{{PrivilegeID: bonus.ID, TicketUID: operation.TicketUID, DateTime: time.Now().UTC(), BalanceDiff: difference, OperationType: operation.OperationType}})
		if err != nil {
			return model.PrivilegeShortInfo{}, err
		}
	}
	return model.PrivilegeShortInfo{Balance: bonus.Balance, Status: bonus.Status}, nil
}

func (service *bonusService) getBonus(ctx context.Context, username string) (model.Privilege, error) {
	bonuses, err := service.bonusRepository.GetAll(ctx)
	if err != nil {
		return model.Privilege{}, err
	}
	for _, bonus := range bonuses {
		if bonus.Username == username {
			return bonus, nil
		}
	}
	bonuses, err = service.bonusRepository.Create(ctx, []model.Privilege{{Username: username, Status: "BRONZE"}})
	if err != nil {
		return model.Privilege{}, err
	}
	return bonuses[0], nil
}

func bonusStatus(balance int) string {
	if balance >= 5000 {
		return "GOLD"
	}
	if balance >= 1000 {
		return "SILVER"
	}
	return "BRONZE"
}
