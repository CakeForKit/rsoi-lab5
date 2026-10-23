package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CakeForKit/rsoi-lab2/lb-core/custom_error"
	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	"github.com/CakeForKit/rsoi-lab2/ms-flight/internal/repository"
)

var flightSrv FlightService
var flightSrvMutex sync.Mutex

type FlightService interface {
	GetAll(ctx context.Context, page int, size int) (model.PaginationResponse, error)
	Get(ctx context.Context, number string) (model.FlightResponse, error)
}

type flightService struct {
	flightRepository  repository.FlightRepository
	airportRepository repository.AirportRepository
}

func GetFlightService() (FlightService, error) {
	flightSrvMutex.Lock()
	defer flightSrvMutex.Unlock()

	if flightSrv != nil {
		return flightSrv, nil
	}

	flightRepository, err := repository.GetFlightRepository()
	if err != nil {
		return nil, err
	}
	airportRepository, err := repository.GetAirportRepository()
	if err != nil {
		return nil, err
	}

	service := &flightService{flightRepository: flightRepository, airportRepository: airportRepository}
	if err := service.seed(context.Background()); err != nil {
		return nil, err
	}
	flightSrv = service
	return flightSrv, nil
}

func (service *flightService) seed(ctx context.Context) error {
	flights, err := service.flightRepository.GetAll(ctx)
	if err != nil || len(flights) != 0 {
		return err
	}

	airports, err := service.airportRepository.Create(ctx, []model.Airport{
		{Name: "Пулково", City: "Санкт-Петербург", Country: "Россия"},
		{Name: "Шереметьево", City: "Москва", Country: "Россия"},
	})
	if err != nil {
		return err
	}

	date, _ := time.Parse("2006-01-02 15:04", "2021-10-08 20:00")
	_, err = service.flightRepository.Create(ctx, []model.Flight{{FlightNumber: "AFL031", DateTime: date, FromAirportID: airports[0].ID, ToAirportID: airports[1].ID, Price: 1500}})
	return err
}

func (service *flightService) GetAll(ctx context.Context, page int, size int) (model.PaginationResponse, error) {
	if page < 0 || size < 1 || size > 100 {
		return model.PaginationResponse{}, custom_error.IllegalArgumentError("page or size")
	}
	flights, err := service.flightRepository.GetAll(ctx)
	if err != nil {
		return model.PaginationResponse{}, err
	}

	items := make([]model.FlightResponse, 0, size)
	start := page * size
	if page > 0 {
		start = (page - 1) * size
	}
	for i := start; i < len(flights) && len(items) < size; i++ {
		flight, err := service.toResponse(ctx, flights[i])
		if err != nil {
			return model.PaginationResponse{}, err
		}
		items = append(items, flight)
	}
	return model.PaginationResponse{Page: page, PageSize: len(items), TotalElements: int64(len(flights)), Items: items}, nil
}

func (service *flightService) Get(ctx context.Context, number string) (model.FlightResponse, error) {
	flights, err := service.flightRepository.GetAll(ctx)
	if err != nil {
		return model.FlightResponse{}, err
	}
	for _, flight := range flights {
		if flight.FlightNumber == number {
			return service.toResponse(ctx, flight)
		}
	}
	return model.FlightResponse{}, custom_error.NotFoundError("flight")
}

func (service *flightService) toResponse(ctx context.Context, flight model.Flight) (model.FlightResponse, error) {
	airports, err := service.airportRepository.GetAll(ctx)
	if err != nil {
		return model.FlightResponse{}, err
	}

	var fromAirport model.Airport
	var toAirport model.Airport
	for _, airport := range airports {
		if airport.ID == flight.FromAirportID {
			fromAirport = airport
		}
		if airport.ID == flight.ToAirportID {
			toAirport = airport
		}
	}

	return model.FlightResponse{FlightNumber: flight.FlightNumber, FromAirport: fmt.Sprintf("%s %s", fromAirport.City, fromAirport.Name), ToAirport: fmt.Sprintf("%s %s", toAirport.City, toAirport.Name), Date: flight.DateTime.Format("2006-01-02 15:04"), Price: flight.Price}, nil
}
