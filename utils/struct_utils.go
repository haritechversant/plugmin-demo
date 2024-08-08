package utils

import (
	"reflect"
)

// CreateDynamicStruct creates a dynamic struct based on a map's keys and values.
func CreateDynamicStruct(data map[string]interface{}) reflect.Value {
	var structFields []reflect.StructField
	for key, value := range data {
		fieldName := ToCamelCase(key)
		field := reflect.StructField{
			Name: fieldName,
			Type: reflect.TypeOf(value),
			Tag:  reflect.StructTag(`json:"` + key + `"`),
		}
		structFields = append(structFields, field)
	}
	structType := reflect.StructOf(structFields)
	structInstance := reflect.New(structType).Elem()
	return structInstance
}

// PopulateStruct populates a dynamic struct with values from the map.
func PopulateStruct(dynamicStruct reflect.Value, data map[string]interface{}) {
	for key, value := range data {
		fieldName := ToCamelCase(key)
		field := dynamicStruct.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(value))
		}
	}
}
