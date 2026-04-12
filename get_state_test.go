package paytbank

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGetState_Success(t *testing.T) {
	var expectedToken string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: got %s, want %s", r.Method, http.MethodPost)
		}

		if r.URL.Path != "/GetState" {
			t.Fatalf("unexpected path: got %s, want /GetState", r.URL.Path)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected Content-Type: got %q", got)
		}

		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected Accept: got %q", got)
		}

		var reqBody TGetState
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("json.NewDecoder().Decode(): %v", err)
		}

		if reqBody.TerminalKey != "term-key" {
			t.Fatalf("unexpected TerminalKey: got %q", reqBody.TerminalKey)
		}

		if reqBody.PaymentID != "payment-123" {
			t.Fatalf("unexpected PaymentID: got %q", reqBody.PaymentID)
		}

		if reqBody.Token != expectedToken {
			t.Fatalf("unexpected Token: got %q", reqBody.Token)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TGetStateResponse{
			BasePayResponse: BasePayResponse{
				Success:     true,
				ErrorCode:   "0",
				TerminalKey: "term-key",
				Status:      "CONFIRMED",
				PaymentID:   "payment-123",
				OrderID:     "order-123",
				Amount:      150000,
			},
			Params: []TGetStateResponseParam{
				{
					Key:   "Email",
					Value: "user@example.com",
				},
				{
					Key:   "Phone",
					Value: "+79990000000",
				},
			},
		}); err != nil {
			t.Fatalf("json.NewEncoder().Encode(): %v", err)
		}
	}))
	defer server.Close()

	req := &TGetState{
		TerminalKey: "term-key",
		PaymentID:   "payment-123",
		Token:       "token-value",
	}

	password := "test-password"

	client := NewClient(server.URL, server.Client(), req.TerminalKey, password)

	var err error
	expectedToken, err = BuildRequestToken(req, password)
	if err != nil {
		t.Fatalf("BuildRequestToken(): %v", err)
	}

	resp, err := client.GetState(context.Background(), req)
	if err != nil {
		t.Fatalf("client.GetState() error = %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected Success=true, got false")
	}

	if resp.ErrorCode != "0" {
		t.Fatalf("unexpected ErrorCode: got %q", resp.ErrorCode)
	}

	if resp.Status != "CONFIRMED" {
		t.Fatalf("unexpected Status: got %q", resp.Status)
	}

	if resp.PaymentID != "payment-123" {
		t.Fatalf("unexpected PaymentID: got %q", resp.PaymentID)
	}

	if resp.OrderID != "order-123" {
		t.Fatalf("unexpected OrderID: got %q", resp.OrderID)
	}

	if resp.Amount != 150000 {
		t.Fatalf("unexpected Amount: got %d", resp.Amount)
	}

	if len(resp.Params) != 2 {
		t.Fatalf("unexpected Params length: got %d, want 2", len(resp.Params))
	}

	if resp.Params[0].Key != "Email" || resp.Params[0].Value != "user@example.com" {
		t.Fatalf("unexpected first param: %+v", resp.Params[0])
	}

	if resp.Params[1].Key != "Phone" || resp.Params[1].Value != "+79990000000" {
		t.Fatalf("unexpected second param: %+v", resp.Params[1])
	}
}

func TestClientGetState_HTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"internal error"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client(), "", "")

	req := &TGetState{
		TerminalKey: "term-key",
		PaymentID:   "payment-123",
		Token:       "token-value",
	}

	_, err := client.GetState(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected http status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetState_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Success":true,"PaymentId":`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client(), "", "")

	req := &TGetState{
		TerminalKey: "term-key",
		PaymentID:   "payment-123",
		Token:       "token-value",
	}

	_, err := client.GetState(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "json.Unmarshal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTGetStateAPIExecute_UsesDefaultClient(t *testing.T) {
	oldDefaultClient := DefaultClient
	t.Cleanup(func() {
		DefaultClient = oldDefaultClient
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/GetState" {
			t.Fatalf("unexpected path: got %s, want /GetState", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TGetStateResponse{
			BasePayResponse: BasePayResponse{
				Success:     true,
				ErrorCode:   "0",
				TerminalKey: "term-key",
				Status:      "AUTHORIZED",
				PaymentID:   "payment-321",
				OrderID:     "order-321",
				Amount:      5000,
			},
			Params: []TGetStateResponseParam{
				{
					Key:   "CardId",
					Value: "card-1",
				},
			},
		}); err != nil {
			t.Fatalf("json.NewEncoder().Encode(): %v", err)
		}
	}))
	defer server.Close()

	DefaultClient = NewClient(server.URL, server.Client(), "", "")

	req := &TGetState{
		TerminalKey: "term-key",
		PaymentID:   "payment-321",
		Token:       "token-value",
	}

	resp, err := req.APIExecute(context.Background())
	if err != nil {
		t.Fatalf("APIExecute() error = %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected Success=true, got false")
	}

	if resp.PaymentID != "payment-321" {
		t.Fatalf("unexpected PaymentID: got %q", resp.PaymentID)
	}

	if resp.Status != "AUTHORIZED" {
		t.Fatalf("unexpected Status: got %q", resp.Status)
	}
}

func TestClientGetState_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client(), "", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &TGetState{
		TerminalKey: "term-key",
		PaymentID:   "payment-123",
		Token:       "token-value",
	}

	_, err := client.GetState(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "client.Do()") {
		t.Fatalf("unexpected error: %v", err)
	}
}
