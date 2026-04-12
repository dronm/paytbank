package paytbank

import "context"

//Получить статус платежа

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
	err := c.apiExecute(ctx, GetStateEndpoint, p, &respPayload)
	if err != nil {
		return respPayload, err
	}

	return respPayload, nil
}

func (p *TGetState) APIExecute(ctx context.Context) (TGetStateResponse, error) {
	return DefaultClient.GetState(ctx, p)
}
