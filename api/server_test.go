package api

import (
	"context"
	"testing"
	"time"
)

type Service struct{}

type Args struct {
	S string
}

type Result struct {
	String string
	Int    int
	Args   *Args
}

func (s *Service) NoArgsReturns() {
}

func (s *Service) Echo(str string, i int, args *Args) Result {
	return Result{str, i, args}
}

func (s *Service) EchoWithCtx(ctx context.Context, str string, i int, args *Args) Result {
	return Result{str, i, args}
}

func (s *Service) Sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
}

func (s *Service) LastErr() (string, error) {
	return "", nil
}

func (s *Service) FirstErr() (error, string) {
	return nil, ""
}

func (s *Service) NoErr() (string, string) {
	return "", ""
}

func (s *Service) MoreThanResult() (string, string, error) {
	return "", "", nil
}

func (s *Service) Subscription(ctx context.Context) (chan int, error) {
	return nil, nil
}

func TestServerRegisterName(t *testing.T) {
	server := NewServer()
	service := new(Service)

	if err := server.RegisterName("calc", service); err != nil {
		t.Fatalf("%v", err)
	}

	if len(server.services) != 2 {
		t.Fatalf("Expected 2 service entries, got %d", len(server.services))
	}

	svc, ok := server.services["calc"]
	if !ok {
		t.Fatalf("Expected service calc to be registered")
	}

	if len(svc.callbacks) != 5 {
		t.Errorf("Expected 5 callbacks for service 'calc', got %d", len(svc.callbacks))
	}

	if len(svc.subscriptions) != 1 {
		t.Errorf("Expected 1 subscription for service 'calc', got %d", len(svc.subscriptions))
	}
}
