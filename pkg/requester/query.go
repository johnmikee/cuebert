package requester

import (
	"fmt"
	"net/url"
	"reflect"
)

type Params map[string]string

type Query struct {
	Endpoint string
	Params   Params
}

func StructToQuery(obj interface{}) map[string]string {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Struct {
		fmt.Printf("Expected struct, got %v", objValue.Kind())
		return nil
	}

	objType := objValue.Type()
	data := make(map[string]string)

	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i)

		// Get the JSON tag value or use the field name
		tagValue := field.Tag.Get("json")
		if tagValue == "" {
			tagValue = field.Name
		}

		// Convert the field value to a string
		var value string
		switch fieldValue.Kind() {
		case reflect.Int:
			value = fmt.Sprintf("%d", fieldValue.Int())
		case reflect.String:
			value = fieldValue.String()
		case reflect.Bool:
			value = fmt.Sprintf("%t", fieldValue.Bool())
		// Add more cases for other types as needed
		default:
			value = fmt.Sprintf("%v", fieldValue.Interface())
		}

		data[tagValue] = value
	}

	return data
}

// BuildQuery only builds the query string, it does not append this to the base URL.
func BuildQuery(q *Query) (string, error) {
	u, err := url.Parse(q.Endpoint)
	if err != nil {
		return "", err
	}

	if q.Params == nil {
		return u.String(), nil
	}

	query := u.Query()

	for key, value := range q.Params {
		if value != "" {
			query.Set(key, value)
		}
	}

	u.RawQuery = query.Encode()

	return u.String(), nil
}
