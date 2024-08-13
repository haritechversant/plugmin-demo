package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)

	}

	jsonData := getSampleJSONData()
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Fatalf("failed to unmarshal JSON data: %v", err)
	}
	ProcessAndInsertData(db, data)
}

func getSampleJSONData() []byte {
	return []byte(`{
		"user": {
			"columnVals": {
				"id": "11212",
				"name": "rhye",
				"email": "mailto:sachin@ggmail.com"
			}
		},
		"address_details": {
			"columnVals": {
				"address": "Techversant",
				"city": "Tvm"
			},
			"ReferenceKey": {
				"user_id": "$user.id"
			}
		},
		"department": {
			"columnVals": {
				"id":"2333122",
				"name": "Computer"
			},
			"ReferenceKey": {
				"user_id": "$user.id"
			}
		},
		"location": {
			"columnVals": {
				"name": "America"
			},
			"ReferenceKey": {
				"user_id": "$user.id",
				"department_id": "$department.id"
			}
		}
	}`)
}

func ProcessAndInsertData(db *gorm.DB, data map[string]interface{}) {
	primaryKeyMap := make(map[string]string)

	for tableName, tableData := range data {
		tableMap, ok := tableData.(map[string]interface{})
		if !ok {
			log.Printf("Invalid data format for table %s", tableName)
			continue
		}

		columnVals, _ := tableMap["columnVals"].(map[string]interface{})

		ResolveReferences(columnVals, tableMap, primaryKeyMap)

		fmt.Printf("Processed columnVals for table %s: %+v\n", tableName, columnVals)

		structValue := createStructFromMap(columnVals)
		if err := db.Table(tableName).Create(structValue.Addr().Interface()).Error; err != nil {
			log.Printf("Failed to insert into table %s: %v", tableName, err)
			continue
		}

		if id, ok := columnVals["id"]; ok {
			primaryKeyMap[tableName+".id"] = id.(string)
		}
	}
}

func ResolveReferences(columnVals map[string]interface{}, tableMap map[string]interface{}, primaryKeyMap map[string]string) {
	refKeysMap, ok := tableMap["ReferenceKey"].(map[string]interface{})
	if !ok {
		return
	}

	for field, ref := range refKeysMap {
		refStr, ok := ref.(string)
		if !ok {
			continue
		}

		refParts := strings.SplitN(refStr[1:], ".", 2)
		if len(refParts) == 2 {
			key := refParts[0] + "." + refParts[1]
			if val, exists := primaryKeyMap[key]; exists {
				columnVals[field] = val
			} else {
				log.Printf("Reference key %s not found in primaryKeyMap", key)
			}
		}
	}
}

func createStructFromMap(fields map[string]interface{}) reflect.Value {
	structFields := make([]reflect.StructField, 0, len(fields))

	for fieldName, fieldValue := range fields {
		structFields = append(structFields, reflect.StructField{
			Name: capitalize(fieldName),
			Type: reflect.TypeOf(fieldValue),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s" gorm:"column:%s"`, fieldName, fieldName)),
		})
	}

	structType := reflect.StructOf(structFields)
	structInstance := reflect.New(structType).Elem()

	for fieldName, fieldValue := range fields {
		structInstance.FieldByName(capitalize(fieldName)).Set(reflect.ValueOf(fieldValue))
	}

	return structInstance
}

func capitalize(str string) string {
	if str == "" {
		return str
	}
	return strings.ToUpper(str[:1]) + str[1:]
}
