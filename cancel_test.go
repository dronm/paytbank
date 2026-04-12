package paytbank

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientCancel_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: got %s, want %s", r.Method, http.MethodPost)
		}

		if r.URL.Path != "/Cancel" {
			t.Fatalf("unexpected path: got %s, want /Cancel", r.URL.Path)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected Content-Type: got %q", got)
		}

		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected Accept: got %q", got)
		}

		var reqBody TCancel
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("json.NewDecoder().Decode(): %v", err)
		}

		if reqBody.TerminalKey != "term-key" {
			t.Fatalf("unexpected TerminalKey: got %q", reqBody.TerminalKey)
		}

		if reqBody.Token != "token-value" {
			t.Fatalf("unexpected Token: got %q", reqBody.Token)
		}

		if reqBody.PaymentID != "payment-123" {
			t.Fatalf("unexpected PaymentID: got %q", reqBody.PaymentID)
		}

		if reqBody.Amount != 150000 {
			t.Fatalf("unexpected Amount: got %d", reqBody.Amount)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TCancelResponse{
			BasePayResponse: BasePayResponse{
				Success:     true,
				ErrorCode:   "0",
				TerminalKey: "term-key",
				Status:      "CANCELED",
				PaymentID:   "payment-123",
				OrderID:     "order-123",
				Amount:      150000,
			},
			OriginalAmount:    150000,
			NewAmount:         0,
			ExternalRequestID: "ext-req-1",
		}); err != nil {
			t.Fatalf("json.NewEncoder().Encode(): %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	req := &TCancel{
		TerminalKey: "term-key",
		Token:       "token-value",
		PaymentID:   "payment-123",
		Amount:      150000,
	}

	resp, err := client.Cancel(context.Background(), req)
	if err != nil {
		t.Fatalf("client.Cancel() error = %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected Success=true, got false")
	}

	if resp.ErrorCode != "0" {
		t.Fatalf("unexpected ErrorCode: got %q", resp.ErrorCode)
	}

	if resp.Status != "CANCELED" {
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

	if resp.OriginalAmount != 150000 {
		t.Fatalf("unexpected OriginalAmount: got %d", resp.OriginalAmount)
	}

	if resp.NewAmount != 0 {
		t.Fatalf("unexpected NewAmount: got %d", resp.NewAmount)
	}

	if resp.ExternalRequestID != "ext-req-1" {
		t.Fatalf("unexpected ExternalRequestID: got %q", resp.ExternalRequestID)
	}
}

func TestClientCancel_HTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"internal error"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	req := &TCancel{
		TerminalKey: "term-key",
		Token:       "token-value",
		PaymentID:   "payment-123",
		Amount:      150000,
	}

	_, err := client.Cancel(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected http status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientCancel_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Success":true,"PaymentId":`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	req := &TCancel{
		TerminalKey: "term-key",
		Token:       "token-value",
		PaymentID:   "payment-123",
		Amount:      150000,
	}

	_, err := client.Cancel(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "json.Unmarshal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTCancelAPIExecute_UsesDefaultClient(t *testing.T) {
	oldDefaultClient := DefaultClient
	t.Cleanup(func() {
		DefaultClient = oldDefaultClient
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Cancel" {
			t.Fatalf("unexpected path: got %s, want /Cancel", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TCancelResponse{
			BasePayResponse: BasePayResponse{
				Success:     true,
				ErrorCode:   "0",
				TerminalKey: "term-key",
				Status:      "CANCELED",
				PaymentID:   "payment-321",
				OrderID:     "order-321",
				Amount:      5000,
			},
			OriginalAmount:    5000,
			NewAmount:         0,
			ExternalRequestID: "ext-req-321",
		}); err != nil {
			t.Fatalf("json.NewEncoder().Encode(): %v", err)
		}
	}))
	defer server.Close()

	DefaultClient = NewClient(server.URL, server.Client())

	req := &TCancel{
		TerminalKey: "term-key",
		Token:       "token-value",
		PaymentID:   "payment-321",
		Amount:      5000,
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

func TestClientCancel_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &TCancel{
		TerminalKey: "term-key",
		Token:       "token-value",
		PaymentID:   "payment-123",
		Amount:      150000,
	}

	_, err := client.Cancel(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "client.Do()") {
		t.Fatalf("unexpected error: %v", err)
	}
}
