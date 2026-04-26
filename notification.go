package paytbank

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Уведомления об операциях
// https://developer.tbank.ru/eacq/intro/developer/notification

type PaymentStatus string

const (
	StatusNew        PaymentStatus = "NEW"
	StatusAuthorized PaymentStatus = "AUTHORIZED"
	StatusConfirmed  PaymentStatus = "CONFIRMED"
	StatusRejected   PaymentStatus = "REJECTED"
	StatusCanceled   PaymentStatus = "CANCELED"
	StatusReversed   PaymentStatus = "REVERSED"
	StatusRefunded   PaymentStatus = "REFUNDED"
)

type Notification struct {
	TerminalKey string          `json:"TerminalKey"`
	OrderID     string          `json:"OrderId"`
	Success     bool            `json:"Success"`
	Status      PaymentStatus   `json:"Status"`
	PaymentID   int64           `json:"PaymentId"`
	ErrorCode   string          `json:"ErrorCode"`
	Amount      int64           `json:"Amount"`
	Token       string          `json:"Token"`
	Data        json.RawMessage `json:"Data,omitempty"`
	Receipt     json.RawMessage `json:"Receipt,omitempty"`
}

func VerifyNotificationToken(raw map[string]any, password string, providedToken string) bool {
	expected := BuildNotificationToken(raw, password)
	return subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(expected)),
		[]byte(strings.ToLower(providedToken)),
	) == 1
}

func BuildNotificationToken(raw map[string]any, password string) string {
	data := make(map[string]string, len(raw)+1)

	for k, v := range raw {
		if k == "Token" || k == "Data" || k == "Receipt" {
			continue
		}

		if v == nil {
			continue
		}

		switch val := v.(type) {
		case map[string]any, []any:
			continue
		case json.Number:
			data[k] = val.String()
		case string:
			data[k] = val
		case bool:
			if val {
				data[k] = "true"
			} else {
				data[k] = "false"
			}
		case float64, float32:
			// panic(fmt.Sprintf("float64 in token source for field %s: decode JSON with UseNumber to avoid precision loss", k))
			data[k] = fmt.Sprintf("%f", val)
		default:
			data[k] = fmt.Sprint(val)
		}
	}

	data["Password"] = password

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(data[k])
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
