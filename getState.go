package paytbank

import (
	"context"
	"fmt"
)

// Получить статус платежа
// https://developer.tbank.ru/eacq/api/get-state

const (
	GetStateEndpoint apiEndpoint = "GetState"
)

// TGetState возвращает статус платежа.
type TGetState struct {
	TerminalKey string `json:"TerminalKey"`
	PaymentID   string `json:"PaymentId"`
	Token       string `json:"Token"`
}

type TGetStateResponseParam struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type TGetStateResponse struct {
	BasePayResponse
	Params      []TGetStateResponseParam `json:"Params"`
}

func (c *Client) GetState(ctx context.Context, p *TGetState) (TGetStateResponse, error) {
	respPayload := TGetStateResponse{}

	req := *p

	if req.TerminalKey == "" {
		req.TerminalKey = c.terminalKey
	}

	token, err := BuildRequestToken(&req, c.password)
	if err != nil {
		return respPayload, fmt.Errorf("BuildRequestToken(): %v", err)
	}
	req.Token = token

	if err := c.apiExecute(ctx, GetStateEndpoint, &req, &respPayload); err != nil {
		return respPayload, err
	}

	return respPayload, nil
}

func (p *TGetState) APIExecute(ctx context.Context) (TGetStateResponse, error) {
	return DefaultClient.GetState(ctx, p)
}
