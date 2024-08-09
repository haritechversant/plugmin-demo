package utils

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// CreateDynamicStruct creates a new struct type based on the provided fields and their types.
func CreateDynamicStruct(fields map[string]reflect.Type) interface{} {
	var structFields []reflect.StructField

	for name, fieldType := range fields {
		if name == "id" {
			// Skip auto-generated primary key
			continue
		}
		// Capitalize the field name to make it exported.
		exportedName := CapitalizeFirstLetter(name)

		structField := reflect.StructField{
			Name: exportedName,
			Type: fieldType,
			Tag:  reflect.StructTag(`json:"` + name + `"`),
		}
		structFields = append(structFields, structField)
	}

	// Create a new struct type with the fields.
	structType := reflect.StructOf(structFields)

	// Create a new instance of the struct type.
	newStruct := reflect.New(structType).Interface()

	return newStruct
}

// PopulateStruct populates a struct based on a map of data.
func PopulateStruct(structPtr interface{}, data map[string]interface{}) {
	val := reflect.ValueOf(structPtr)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		fmt.Println("Invalid struct pointer")
		return
	}

	val = val.Elem() // Use Elem() to get the value of the pointer
	if !val.IsValid() || val.Kind() != reflect.Struct {
		fmt.Println("Invalid struct value or not a struct")
		return
	}

	for key, value := range data {
		// Capitalize the first letter of the key to match the struct field name.
		capitalizedKey := CapitalizeFirstLetter(key)
		fieldVal := val.FieldByName(capitalizedKey)
		if fieldVal.IsValid() && fieldVal.CanSet() {
			val := reflect.ValueOf(value)

			// Handle type conversion
			if fieldVal.Type() == reflect.TypeOf(int(0)) && val.Kind() == reflect.Float64 {
				val = reflect.ValueOf(int(val.Float()))
			}

			if fieldVal.Type() == val.Type() {
				fieldVal.Set(val)
			} else {
				fmt.Printf("Type mismatch for field %s: expected %s but got %s\n", key, fieldVal.Type(), val.Type())
			}
		} else {
			fmt.Printf("Cannot set field %s\n", key)
		}
	}
}

// ValidateStruct validates a struct based on JSON schema.
func ValidateStruct(data interface{}, schema map[string]interface{}) error {
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		return err
	}
	return nil
}
