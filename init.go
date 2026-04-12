package paytbank

import (
	"context"
	"fmt"
)

// Инициировать платеж
// https://developer.tbank.ru/eacq/api/init

const (
	InitEndpoint apiEndpoint = "Init"
)

// TInit Инициировать платеж
type TInit struct {
	TerminalKey     string `json:"TerminalKey"`
	Amount          int64  `json:"Amount"`
	Token           string `json:"Token"`
	OrderID         string `json:"OrderId" maxlen:"36"`                // unique order ID
	Description     string `json:"Description,omitempty" maxlen:"140"` // order description
	CustomerKey     string `json:"CustomerKey,omitempty"`
	NotificationURL string `json:"NotificationURL,omitempty"`
	SuccessURL      string `json:"SuccessURL,omitempty"`
	FailURL         string `json:"FailURL,omitempty"`
}

type TInitResponse struct {
	BasePayResponse
	PaymentURL string `json:"PaymentURL"`
}

func (c *Client) Init(ctx context.Context, p *TInit) (TInitResponse, error) {
	respPayload := TInitResponse{}

	req := *p

	if req.TerminalKey == "" {
		req.TerminalKey = c.terminalKey
	}

	token, err := BuildRequestToken(&req, c.password)
	if err != nil {
		return respPayload, fmt.Errorf("BuildRequestToken(): %v", err)
	}
	req.Token = token

	if err := c.apiExecute(ctx, InitEndpoint, &req, &respPayload); err != nil {
		return respPayload, err
	}

	return respPayload, nil
}

func (p *TInit) APIExecute(ctx context.Context) (TInitResponse, error) {
	return DefaultClient.Init(ctx, p)
}

