package helpers

import (
	"fmt"
	"reflect"
	"strings"
)

var fieldsByTag = make(map[reflect.Type]map[string]map[string]string)

const keyType = "json"

// GetStructKeys will return the json tags from a struct.
func GetStructKeys(t interface{}) []string {
	val := reflect.ValueOf(t)
	typ := val.Type()
	// If the type is a pointer, get the type it points to
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	keys := []string{}
	for i := 0; i < typ.NumField(); i++ {
		k := typ.Field(i).Tag.Get("json")
		v := strings.Split(k, ",")[0]
		if v == "" || v == "-" {
			continue
		}
		keys = append(keys, v)
	}
	return keys
}

func GetFieldName(tag string, s interface{}) (string, error) {
	rt := reflect.TypeOf(s)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected a struct or a pointer to struct, got %s", rt.Kind())
	}

	buildFieldsByTagMap(s)
	if name, ok := fieldsByTag[rt][keyType][tag]; ok {
		return name, nil
	}
	return "", fmt.Errorf("tag not found")
}

func BuildMap(s interface{}) {
	buildFieldsByTagMap(s)
}

func buildFieldsByTagMap(s interface{}) {
	rt := reflect.TypeOf(s)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return
	}

	if fieldsByTag[rt] == nil {
		fieldsByTag[rt] = make(map[string]map[string]string)
	}
	if fieldsByTag[rt][keyType] == nil {
		fieldsByTag[rt][keyType] = make(map[string]string)
	}

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := strings.Split(f.Tag.Get(keyType), ",")[0]
		if v == "" || v == "-" {
			continue
		}
		fieldsByTag[rt][keyType][v] = f.Name
	}
}

func GetTagMapping(structType reflect.Type) map[string]interface{} {
	tagMapping := make(map[string]interface{})

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			tagMapping[jsonTag] = field.Name
		}
	}

	return tagMapping
}
