package helpers

import (
	"fmt"
	"reflect"
	"strings"
)

var fieldsByTag = make(map[reflect.Type]map[string]map[string]string)

const keyType = "json"

// GetStructKeys will return the json tags from a struct
// https://stackoverflow.com/questions/55879028/golang-get-structs-field-name-by-json-tag
func GetStructKeys(t interface{}) []string {
	keys := []string{}

	val := reflect.ValueOf(t)
	for i := 0; i < val.Type().NumField(); i++ {
		k := val.Type().Field(i).Tag.Get("json")
		v := strings.Split(k, ",")[0]
		if v == "" || v == "-" {
			continue
		}
		keys = append(keys, v)
	}

	return keys
}

func GetFieldName(tag string, s interface{}) (string, error) {
	buildFieldsByTagMap(s)

	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		return "", fmt.Errorf("could not get field name")
	}
	return fieldsByTag[rt][keyType][tag], nil
}

func buildFieldsByTagMap(s interface{}) {
	rt := reflect.TypeOf(s)

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
