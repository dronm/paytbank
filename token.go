package paytbank

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// BuildRequestToken builds a T-Bank request token for any request struct.
// Rules:
// - include only root-level fields present in the request;
// - exclude Token itself;
// - exclude nested objects and arrays;
// - add Password;
// - sort by key;
// - concatenate only values;
// - SHA-256 the resulting string.
func BuildRequestToken(req any, password string) (string, error) {
	if req == nil {
		return "", fmt.Errorf("req is nil")
	}

	rv := reflect.ValueOf(req)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return "", fmt.Errorf("req must be a non-nil pointer to struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return "", fmt.Errorf("req must point to a struct")
	}

	rt := rv.Type()

	data := make(map[string]string, rv.NumField()+1)
	data["Password"] = password

	for i := 0; i < rv.NumField(); i++ {
		sf := rt.Field(i)
		fv := rv.Field(i)

		if !fv.CanInterface() {
			continue
		}

		tag := sf.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		jsonName, omitEmpty := parseJSONTag(tag)
		if jsonName == "" || jsonName == "Token" {
			continue
		}

		if shouldSkipField(fv, omitEmpty) {
			continue
		}

		strVal, ok := scalarFieldToString(fv)
		if !ok {
			continue
		}

		data[jsonName] = strVal
	}

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
	return hex.EncodeToString(sum[:]), nil
}

func parseJSONTag(tag string) (string, bool) {
	parts := strings.Split(tag, ",")
	name := parts[0]

	omitZero := false
	for _, p := range parts[1:] {
		switch p {
		case "omitempty", "omitzero":
			omitZero = true
		}
	}

	return name, omitZero
}

func shouldSkipField(v reflect.Value, omitEmpty bool) bool {
	if !v.IsValid() {
		return true
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return true
		}

		elem := v.Elem()
		if isNestedKind(elem.Kind()) {
			return true
		}

		if omitEmpty && elem.IsZero() {
			return true
		}

		return false
	}

	if isNestedKind(v.Kind()) {
		return true
	}

	if omitEmpty && v.IsZero() {
		return true
	}

	return false
}

func isNestedKind(k reflect.Kind) bool {
	switch k {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

func scalarFieldToString(v reflect.Value) (string, bool) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), true
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), true
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), true
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), true
	default:
		return "", false
	}
}
