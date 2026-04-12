package paytbank

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientInit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: got %s, want %s", r.Method, http.MethodPost)
		}

		if r.URL.Path != "/Init" {
			t.Fatalf("unexpected path: got %s, want /Init", r.URL.Path)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected Content-Type: got %q", got)
		}

		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected Accept: got %q", got)
		}

		var reqBody TInit
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("json.NewDecoder().Decode(): %v", err)
		}

		if reqBody.TerminalKey != "term-key" {
			t.Fatalf("unexpected TerminalKey: got %q", reqBody.TerminalKey)
		}

		if reqBody.Amount != 125500 {
			t.Fatalf("unexpected Amount: got %d", reqBody.Amount)
		}

		if reqBody.Token != "token-value" {
			t.Fatalf("unexpected Token: got %q", reqBody.Token)
		}

		if reqBody.OrderID != "order-123" {
			t.Fatalf("unexpected OrderID: got %q", reqBody.OrderID)
		}

		if reqBody.Description != "test payment" {
			t.Fatalf("unexpected Description: got %q", reqBody.Description)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TInitResponse{
			BasePayResponse: BasePayResponse{
				Success:     true,
				ErrorCode:   "0",
				TerminalKey: "term-key",
				Status:      "NEW",
				PaymentID:   "payment-999",
				OrderID:     "order-123",
				Amount:      125500,
				Message:     "",
				Details:     "",
			},
			PaymentURL: "https://securepay.tinkoff.ru/mock-payment-url",
		}); err != nil {
			t.Fatalf("json.NewEncoder().Encode(): %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	req := &TInit{
		TerminalKey: "term-key",
		Amount:      125500,
		Token:       "token-value",
		OrderID:     "order-123",
		Description: "test payment",
	}

	resp, err := client.Init(context.Background(), req)
	if err != nil {
		t.Fatalf("client.Init() error = %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected Success=true, got false")
	}

	if resp.ErrorCode != "0" {
		t.Fatalf("unexpected ErrorCode: got %q", resp.ErrorCode)
	}

	if resp.Status != "NEW" {
		t.Fatalf("unexpected Status: got %q", resp.Status)
	}

	if resp.PaymentID != "payment-999" {
		t.Fatalf("unexpected PaymentID: got %q", resp.PaymentID)
	}

	if resp.OrderID != "order-123" {
		t.Fatalf("unexpected OrderID: got %q", resp.OrderID)
	}

	if resp.Amount != 125500 {
		t.Fatalf("unexpected Amount: got %d", resp.Amount)
	}

	if resp.PaymentURL != "https://securepay.tinkoff.ru/mock-payment-url" {
		t.Fatalf("unexpected PaymentURL: got %q", resp.PaymentURL)
	}
}

func TestClientInit_HTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"internal error"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	req := &TInit{
		TerminalKey: "term-key",
		Amount:      1000,
		Token:       "token-value",
		OrderID:     "order-123",
	}

	_, err := client.Init(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected http status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientInit_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Success":true,"PaymentId":`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	req := &TInit{
		TerminalKey: "term-key",
		Amount:      1000,
		Token:       "token-value",
		OrderID:     "order-123",
	}

	_, err := client.Init(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "json.Unmarshal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTInitAPIExecute_UsesDefaultClient(t *testing.T) {
	oldDefaultClient := DefaultClient
	t.Cleanup(func() {
		DefaultClient = oldDefaultClient
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TInitResponse{
			BasePayResponse: BasePayResponse{
				Success:   true,
				ErrorCode: "0",
				Status:    "NEW",
				PaymentID: "payment-321",
				OrderID:   "order-321",
				Amount:    5000,
			},
			PaymentURL: "https://securepay.tinkoff.ru/mock-payment-url-321",
		}); err != nil {
			t.Fatalf("json.NewEncoder().Encode(): %v", err)
		}
	}))
	defer server.Close()

	DefaultClient = NewClient(server.URL, server.Client())

	req := &TInit{
		TerminalKey: "term-key",
		Amount:      5000,
		Token:       "token-value",
		OrderID:     "order-321",
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
}

func TestClientInit_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &TInit{
		TerminalKey: "term-key",
		Amount:      1000,
		Token:       "token-value",
		OrderID:     "order-123",
	}

	_, err := client.Init(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "client.Do()") {
		t.Fatalf("unexpected error: %v", err)
	}
}
