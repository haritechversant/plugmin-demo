package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Function to replace dynamic values with static values and remove references
func replaceDynamicValues(jsonInput string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonInput), &data); err != nil {
		return "", err
	}

	// Extract dynamic values
	staticValues := make(map[string]interface{})
	extractValues(data, staticValues, "")

	// Replace dynamic values in the rest of the data
	replaceValues(data, staticValues)

	// Convert the modified map back to JSON
	result, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// Recursive function to extract values into a map
func extractValues(data map[string]interface{}, values map[string]interface{}, prefix string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			if key == "columnVals" {
				// Store column values with their full key
				for k, val := range v {
					if str, ok := val.(string); ok {
						values[fullKey+"."+k] = str
					}
				}
			} else {
				// Continue extracting from nested maps
				extractValues(v, values, fullKey)
			}
		}
	}
}

// Recursive function to replace dynamic values
func replaceValues(data map[string]interface{}, staticValues map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			if key == "columnVals" {
				for k, val := range v {
					if str, ok := val.(string); ok {
						if strings.HasPrefix(str, "$") {
							refKey := strings.TrimPrefix(str, "$")
							if refValue, exists := staticValues[refKey]; exists {
								v[k] = refValue
							}
						}
					}
				}
				// Remove any referenceKey objects in columnVals
				// for k := range v {
				// 	if strings.HasPrefix(k, "referenceKey") {
				// 		delete(v, k)
				// 	}
				// }
			} else {
				replaceValues(v, staticValues)
			}
		}
	}
}

func main() {
	jsonInput := `{
		"user": {
			"columnVals": {
				"id": "122344",
				"name": "Aiswarya",
				"email": "aiswarya@gmail.com"
			}
		},
		"address_details": {
			"columnVals": {
				"address": "thiruvathira",
				"user_id": "$user.columnVals.id",
				"city": "abc",
				"referenceKey": {
					"user_id": "$user.columnVals.id"
				}
			}
		}
	}`

	result, err := replaceDynamicValues(jsonInput)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(result)
}
