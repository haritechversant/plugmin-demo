package utils

import (
	"fmt"
	"reflect"
)

// CreateDynamicStruct creates a dynamic struct based on a map's keys and values.
func CreateDynamicStruct(data map[string]interface{}) reflect.Value {
	fmt.Println("----output1------", data)
	//----output1------ map[age:30 email:sagar@example.com name:Sagar P type:free_user]
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
	fmt.Println("----output2------", structType)
	//----output2------ struct { Name string "json:\"name\""; Age float64 "json:\"age\""; Type string "json:\"type\""; Email string "json:\"email\"" }
	structInstance := reflect.New(structType).Elem()
	fmt.Println("----output3------", structInstance)
	//----output3------ { 0  }
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
