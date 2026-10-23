package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/CakeForKit/rsoi-lab2/lb-core/config"
	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
)

const (
	queueRetryTimeout = 10 * time.Second
)

// UnavailableError lets HTTP handlers distinguish a temporary dependency
// outage from a regular request-validation error.
type UnavailableError struct{ Service string }

func (e *UnavailableError) Error() string { return e.Service + " Service unavailable" }

type remoteError struct {
	status  int
	message string
}

func (e *remoteError) Error() string { return e.message }

type cancelJob struct {
	username string
	token    string
	ticket   model.Ticket
}

type GatewayService struct {
	client                         *http.Client
	flightURL, ticketURL, bonusURL string
	breakers                       map[string]*circuitBreaker
	cancelQueue                    chan cancelJob
}

// Authorize obtains an access token using the configured Resource Owner
// Password credentials flow.
func Authorize(ctx context.Context, auth config.AuthProperty, username, password string) (json.RawMessage, int, error) {
	if auth.TokenURL == "" || auth.ClientID == "" || auth.ClientSecret == "" {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("OAuth provider is not configured")
	}
	form := url.Values{
		"grant_type":    {"password"},
		"client_id":     {auth.ClientID},
		"client_secret": {auth.ClientSecret},
		"username":      {username},
		"password":      {password},
		"scope":         {"openid profile email"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, auth.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("OAuth provider is unavailable")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid credentials")
	}
	if !json.Valid(body) {
		return nil, http.StatusBadGateway, fmt.Errorf("OAuth provider returned invalid response")
	}
	return body, http.StatusOK, nil
}

func New() *GatewayService {
	s := &GatewayService{
		client:    &http.Client{Timeout: 5 * time.Second},
		flightURL: "http://ms-flight:8060",
		ticketURL: "http://ms-ticket:8070",
		bonusURL:  "http://ms-bonus:8050",
		breakers: map[string]*circuitBreaker{
			"Flight": {}, "Ticket": {}, "Bonus": {},
		},
		cancelQueue: make(chan cancelJob, 100),
	}
	go s.processCancellationQueue()
	return s
}

func (s *GatewayService) serviceName(base string) string {
	switch base {
	case s.flightURL:
		return "Flight"
	case s.ticketURL:
		return "Ticket"
	default:
		return "Bonus"
	}
}

func (s *GatewayService) call(method, base, path, token string, body any, out any) error {
	if method == http.MethodGet {
		breaker := s.breakers[s.serviceName(base)]
		if !breaker.allow() {
			return &UnavailableError{Service: s.serviceName(base)}
		}
		err := s.doCall(method, base, path, token, body, out)
		if isUnavailable(err) {
			breaker.failure()
			return &UnavailableError{Service: s.serviceName(base)}
		}
		breaker.success()
		return err
	}

	err := s.doCall(method, base, path, token, body, out)
	if isUnavailable(err) {
		return &UnavailableError{Service: s.serviceName(base)}
	}
	return err
}

func (s *GatewayService) doCall(method, base, path, token string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		var payload struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		if payload.Message == "" {
			payload.Message = resp.Status
		}
		return &remoteError{status: resp.StatusCode, message: payload.Message}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func isUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var remote *remoteError
	return !errors.As(err, &remote) || remote.status >= http.StatusInternalServerError
}

func (s *GatewayService) Flights(token string, page, size int) (model.PaginationResponse, error) {
	var result model.PaginationResponse
	return result, s.call(http.MethodGet, s.flightURL, "/internal/flights?page="+fmt.Sprint(page)+"&size="+fmt.Sprint(size), token, nil, &result)
}

func (s *GatewayService) flight(token, number string) (model.FlightResponse, error) {
	var result model.FlightResponse
	return result, s.call(http.MethodGet, s.flightURL, "/internal/flights/"+url.PathEscape(number), token, nil, &result)
}

func (s *GatewayService) GetBonus(token, username string) (model.PrivilegeInfoResponse, error) {
	var result model.PrivilegeInfoResponse
	return result, s.call(http.MethodGet, s.bonusURL, "/internal/bonus", token, nil, &result)
}

func (s *GatewayService) tickets(token string) ([]model.Ticket, error) {
	var result []model.Ticket
	return result, s.call(http.MethodGet, s.ticketURL, "/internal/tickets", token, nil, &result)
}

func (s *GatewayService) ticket(token, uid string) (model.Ticket, error) {
	var result model.Ticket
	return result, s.call(http.MethodGet, s.ticketURL, "/internal/tickets/"+url.PathEscape(uid), token, nil, &result)
}

func (s *GatewayService) ticketResponse(token string, ticket model.Ticket) model.TicketResponse {
	flight, err := s.flight(token, ticket.FlightNumber)
	if err != nil {
		// Flight data enriches a ticket, but the ticket itself remains valid.
		return model.TicketResponse{TicketUID: ticket.TicketUID, FlightNumber: ticket.FlightNumber, Price: ticket.Price, Status: ticket.Status}
	}
	return model.TicketResponse{TicketUID: ticket.TicketUID, FlightNumber: ticket.FlightNumber, FromAirport: flight.FromAirport, ToAirport: flight.ToAirport, Date: flight.Date, Price: ticket.Price, Status: ticket.Status}
}

func (s *GatewayService) Tickets(token, username string) ([]model.TicketResponse, error) {
	tickets, err := s.tickets(token)
	if err != nil {
		return nil, err
	}
	result := make([]model.TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, s.ticketResponse(token, ticket))
	}
	return result, nil
}

func (s *GatewayService) Ticket(token, username, uid string) (model.TicketResponse, error) {
	ticket, err := s.ticket(token, uid)
	if err != nil {
		return model.TicketResponse{}, err
	}
	return s.ticketResponse(token, ticket), nil
}

func (s *GatewayService) applyBonus(token string, operation model.PrivilegeOperationRequest) (model.PrivilegeShortInfo, error) {
	var result model.PrivilegeShortInfo
	return result, s.call(http.MethodPost, s.bonusURL, "/internal/bonus/operations", token, operation, &result)
}

func (s *GatewayService) rollbackTicket(token, ticketUID string) {
	_ = s.call(http.MethodDelete, s.ticketURL, "/internal/tickets/"+url.PathEscape(ticketUID), token, nil, nil)
}

func (s *GatewayService) Purchase(token, username string, req model.TicketPurchaseRequest) (model.TicketPurchaseResponse, error) {
	if username == "" || req.FlightNumber == "" || req.Price <= 0 {
		return model.TicketPurchaseResponse{}, fmt.Errorf("invalid ticket purchase request")
	}
	flight, err := s.flight(token, req.FlightNumber)
	if err != nil {
		return model.TicketPurchaseResponse{}, err
	}
	if req.Price != flight.Price {
		return model.TicketPurchaseResponse{}, fmt.Errorf("ticket price does not match flight price")
	}
	var balance model.PrivilegeInfoResponse
	if req.PaidFromBalance {
		balance, err = s.GetBonus(token, username)
		if err != nil {
			return model.TicketPurchaseResponse{}, err
		}
	}
	var ticket model.Ticket
	err = s.call(http.MethodPost, s.ticketURL, "/internal/tickets", token, map[string]any{"flightNumber": req.FlightNumber, "price": req.Price}, &ticket)
	if err != nil {
		return model.TicketPurchaseResponse{}, err
	}
	paidBonuses := 0
	operation := model.PrivilegeOperationRequest{TicketUID: ticket.TicketUID}
	if req.PaidFromBalance {
		paidBonuses = min(balance.Balance, req.Price)
		operation.Amount = paidBonuses
		operation.OperationType = model.PrivilegeOperationDebit
	} else {
		operation.Amount = req.Price / 10
		operation.OperationType = model.PrivilegeOperationFillInBalance
	}
	privilege, err := s.applyBonus(token, operation)
	if err != nil {
		s.rollbackTicket(token, ticket.TicketUID.String())
		return model.TicketPurchaseResponse{}, err
	}
	return model.TicketPurchaseResponse{TicketResponse: s.ticketResponse(token, ticket), PaidByMoney: req.Price - paidBonuses, PaidByBonuses: paidBonuses, Privilege: privilege}, nil
}

func (s *GatewayService) Cancel(token, username, uid string) error {
	ticket, err := s.ticket(token, uid)
	if err != nil {
		return err
	}
	if err = s.call(http.MethodPut, s.ticketURL, "/internal/tickets/"+url.PathEscape(uid)+"/cancel", token, nil, nil); err != nil {
		return err
	}
	if err = s.reverseBonus(token, username, ticket); err != nil {
		s.enqueueCancellation(cancelJob{username: username, token: token, ticket: ticket})
	}
	return nil
}

func (s *GatewayService) reverseBonus(token, username string, ticket model.Ticket) error {
	privilege, err := s.GetBonus(token, username)
	if err != nil {
		return err
	}
	for _, item := range privilege.History {
		if item.TicketUID != ticket.TicketUID {
			continue
		}
		operationType := model.PrivilegeOperationDebit
		if item.OperationType == model.PrivilegeOperationDebit {
			operationType = model.PrivilegeOperationFillInBalance
		}
		_, err = s.applyBonus(token, model.PrivilegeOperationRequest{TicketUID: ticket.TicketUID, Amount: abs(item.BalanceDiff), OperationType: operationType})
		return err
	}
	return nil
}

func (s *GatewayService) enqueueCancellation(job cancelJob) {
	select {
	case s.cancelQueue <- job:
	default:
		go func() { s.cancelQueue <- job }()
	}
}

func (s *GatewayService) processCancellationQueue() {
	for job := range s.cancelQueue {
		if err := s.reverseBonus(job.token, job.username, job.ticket); err != nil {
			time.AfterFunc(queueRetryTimeout, func() { s.enqueueCancellation(job) })
		}
	}
}

func (s *GatewayService) Me(token, username string) (model.UserInfoResponse, error) {
	tickets, err := s.Tickets(token, username)
	if err != nil {
		return model.UserInfoResponse{}, err
	}
	privilege, err := s.GetBonus(token, username)
	if err != nil {
		return model.UserInfoResponse{Tickets: tickets}, nil
	}
	return model.UserInfoResponse{Tickets: tickets, Privilege: model.PrivilegeShortInfo{Balance: privilege.Balance, Status: privilege.Status}}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
