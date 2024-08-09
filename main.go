package main

import (
	"encoding/json"
	"fmt"
	"orm-sample/db"
	"orm-sample/utils"
	"reflect"

	"gorm.io/gorm"
)

type AdminConfig struct {
	TableName  string `gorm:"column:tablename"`
	SchemaJSON string `gorm:"column:jsonschema"`
}

func getSchemaFromDB(dbConn *gorm.DB, identifier string) (map[string]interface{}, string, error) {
	var config AdminConfig
	err := dbConn.Table("admin_config").
		Where("id = ?", identifier).
		First(&config).Error
	if err != nil {
		return nil, "", fmt.Errorf("error fetching schema from DB: %v", err)
	}

	var schema map[string]interface{}
	err = json.Unmarshal([]byte(config.SchemaJSON), &schema)
	if err != nil {
		return nil, "", fmt.Errorf("error unmarshalling schema JSON: %v", err)
	}

	return schema, config.TableName, nil
}

func createDynamicStructFromSchema(schema map[string]interface{}) (interface{}, error) {
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid schema format")
	}

	fields := make(map[string]reflect.Type)
	for key, prop := range properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid property format")
		}

		fieldType, exists := propMap["type"].(string)
		if !exists {
			return nil, fmt.Errorf("missing type in property %s", key)
		}

		var goType reflect.Type
		switch fieldType {
		case "string":
			goType = reflect.TypeOf("")
		case "integer":
			goType = reflect.TypeOf(int(0))
		default:
			return nil, fmt.Errorf("unsupported field type %s", fieldType)
		}

		fields[key] = goType
	}

	return utils.CreateDynamicStruct(fields), nil
}

func HandleRequest(method, identifier, dbType, requestBody string, config db.DBConfig, queryParams map[string]interface{}) {
	dbConn, err := db.InitDB(config, dbType)
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	schema, tableName, err := getSchemaFromDB(dbConn, identifier)
	if err != nil {
		fmt.Println(err)
		return
	}
	if schema == nil || tableName == "" {
		fmt.Println("Error: Schema or table name not found.")
		return
	}

	dynamicStruct, err := createDynamicStructFromSchema(schema)
	if err != nil {
		fmt.Println(err)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(requestBody), &data)
	if err != nil {
		fmt.Println("Error unmarshalling request body:", err)
		return
	}

	if method == "POST" && (requestBody == "" || len(data) == 0) {
		fmt.Println("Error: Request body cannot be empty for POST operation.")
		return
	}

	if method == "PATCH" {
		id, ok := queryParams["id"]
		if !ok || id == "" {
			fmt.Println("Error: ID must be provided for PATCH operation.")
			return
		}
	}

	// Populate the dynamic struct
	utils.PopulateStruct(dynamicStruct, data)

	if err := utils.ValidateStruct(dynamicStruct, schema); err != nil {
		fmt.Println("Validation error:", err)
		return
	}

	if err := dbConn.Table(tableName).AutoMigrate(dynamicStruct); err != nil {
		fmt.Println("Error auto-migrating schema:", err)
		return
	}

	switch method {
	case "POST":
		delete(data, "id")
		if err := dbConn.Table(tableName).Create(dynamicStruct).Error; err != nil {
			fmt.Println("Error inserting data:", err)
		} else {
			fmt.Println("Operation success")
		}
	case "GET":
		var results []map[string]interface{}
		fmt.Printf("Query parameters: %+v\n", queryParams)
		query := dbConn.Table(tableName)
		if len(queryParams) > 0 {
			query = query.Where(queryParams)
		}

		if err := query.Find(&results).Error; err != nil {
			fmt.Println("Error retrieving data:", err)
		} else if len(results) == 0 {
			fmt.Println("No data found matching query parameters.")
		} else {
			resultJSON, _ := json.Marshal(results)
			fmt.Println("Retrieved data:", string(resultJSON))
		}
	case "PATCH":
		id := queryParams["id"]
		if err := dbConn.Table(tableName).Where("id = ?", id).Updates(dynamicStruct).Error; err != nil {
			fmt.Println("Error updating data:", err)
		} else {
			fmt.Println("Operation success")
		}
	default:
		fmt.Println("Unsupported method:", method)
	}
}

func main() {
	// dbType := "mysql"
	dbType := "postgres"

	dbConfig := &db.PostgresConfig{
		Host:     "10.1.0.195",
		User:     "tuneverse_user",
		Password: "S3cretPassWord",
		DbName:   "plugmin",
		Port:     5432,
		SSLMode:  "disable",
	}

	// dbConfig := &db.MySQLConfig{
	// 	Host:     "localhost",
	// 	User:     "admin",
	// 	Password: "Str0ngP@ssw0rd!",
	// 	DbName:   "plugmin",
	// 	Port:     3306,
	// }

	//mysql identifier
	//identifier := "92ad7e1a-5611-11ef-acb1-c81f664b7676"
	//postgres identifier
	identifier := "b157a6eb-debb-4882-bb2e-85f1c39b0c86"

	method := "POST"

	queryParams := map[string]interface{}{
		"id": "0ca05367-6850-44a9-be05-c262860f62aa",
	}

	requestBody := `{
	    "name": "Assim Sang",
		"email": "assim@example.com",
		"age": 20
	}`

	HandleRequest(method, identifier, dbType, requestBody, dbConfig, queryParams)
}
