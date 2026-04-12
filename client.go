// Package paytbank implements bank payment API.
// https://developer.tbank.ru/eacq/api
package paytbank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const DefaultAPIURL = "https://securepay.tinkoff.ru/v2"

type apiEndpoint = string

type BasePayResponse struct {
	Success     bool   `json:"Success"`
	ErrorCode   string `json:"ErrorCode"`
	TerminalKey string `json:"TerminalKey"`
	Status      string `json:"Status"`
	PaymentID   string `json:"PaymentId"`
	OrderID     string `json:"OrderId"`
	Amount      int64  `json:"Amount"`
	Message     string `json:"Message"` // Краткое описание ошибки.
	Details     string `json:"Details"` // Подробное описание ошибки.
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

var DefaultClient = NewClient(DefaultAPIURL, nil)

func NewClient(baseURL string, httpClient *http.Client) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *Client) apiExecute(
	ctx context.Context,
	endpoint apiEndpoint,
	payload any,
	respPayload any,
) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}

	if !isPointerToStruct(payload) {
		return fmt.Errorf("payload should be a pointer to a struct")
	}

	if !isPointerToStruct(respPayload) {
		return fmt.Errorf("respPayload should be a pointer to a struct")
	}

	payloadB, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal(): %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/"+endpoint,
		bytes.NewReader(payloadB),
	)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext(): %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do(): %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("res.Body ReadAll(): %v", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("unexpected http status %d: %s", res.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, respPayload); err != nil {
		return fmt.Errorf("res.Body json.Unmarshal(): %v", err)
	}

	return nil
}

func isPointerToStruct(v any) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}

	if t.Kind() != reflect.Pointer {
		return false
	}

	return t.Elem().Kind() == reflect.Struct
}
