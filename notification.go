package paytbank

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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
		// Per T-Bank docs: exclude Token and nested objects like Data / Receipt.
		if k == "Token" || k == "Data" || k == "Receipt" {
			continue
		}

		switch val := v.(type) {
		case nil:
			// Docs note null values are not included in notification formation.
			continue
		case map[string]any, []any:
			// Defensive skip for any nested structures.
			continue
		default:
			data[k] = scalarToString(val)
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

func scalarToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		// JSON numbers decode into float64 in map[string]any.
		// For payment fields they are expected to be integer-like.
		return strconv.FormatInt(int64(val), 10)
	case json.Number:
		return val.String()
	default:
		return fmt.Sprint(val)
	}
}
