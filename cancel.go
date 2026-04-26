package paytbank

import (
	"context"
	"fmt"
)

// Отменить платеж
// https://developer.tbank.ru/eacq/api/cancel

const (
	CancelEndpoint apiEndpoint = "Cancel"
)

// TCancel Отменить платеж
type TCancel struct {
	TerminalKey string `json:"TerminalKey"`
	Token       string `json:"Token"`
	PaymentID   string `json:"PaymentId"`
	Amount      *int64  `json:"Amount,omitzero"`
}

type TCancelResponse struct {
	BasePayResponse
	OriginalAmount    int64  `json:"OriginalAmount"`
	NewAmount         int64  `json:"NewAmount"`
	ExternalRequestID string `json:"ExternalRequestId"`
}

func (c *Client) Cancel(ctx context.Context, p *TCancel) (TCancelResponse, error) {
	respPayload := TCancelResponse{}

	req := *p

	if req.TerminalKey == "" {
		req.TerminalKey = c.terminalKey
	}

	token, err := BuildRequestToken(&req, c.password)
	if err != nil {
		return respPayload, fmt.Errorf("BuildRequestToken(): %v", err)
	}
	req.Token = token

	if err := c.apiExecute(ctx, CancelEndpoint, &req, &respPayload); err != nil {
		return respPayload, err
	}

	return respPayload, nil
}

func (p *TCancel) APIExecute(ctx context.Context) (TCancelResponse, error) {
	return DefaultClient.Cancel(ctx, p)
}
