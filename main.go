package main

import (
	"encoding/json"
	"fmt"
	"orm-sample/db"
	"orm-sample/utils"
)

func HandleRequest(method, tableName, dbType, requestBody string, config db.DBConfig, queryParams map[string]interface{}) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(requestBody), &data)
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

	dynamicStruct := utils.CreateDynamicStruct(data)
	utils.PopulateStruct(dynamicStruct, data)

	dbConn, err := db.InitDB(config, dbType)
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	// Ensure the model is used correctly for auto-migration
	if err := dbConn.Table(tableName).AutoMigrate(dynamicStruct.Interface()); err != nil {
		fmt.Println("Error auto-migrating schema:", err)
		return
	}

	switch method {
	case "POST":
		if err := dbConn.Table(tableName).Create(dynamicStruct.Interface()).Error; err != nil {
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
		if err := dbConn.Table(tableName).Where("id = ?", id).Updates(dynamicStruct.Interface()).Error; err != nil {
			fmt.Println("Error updating data:", err)
		} else {
			fmt.Println("Operation success")
		}
	default:
		fmt.Println("Unsupported method:", method)
	}
}

func main() {
	dbType := "mysql"
	// dbType := "postgres"

	// dbConfig := &db.PostgresConfig{
	// 	Host:     "10.1.0.195",
	// 	User:     "tuneverse_user",
	// 	Password: "S3cretPassWord",
	// 	DbName:   "plugmin",
	// 	Port:     5432,
	// 	SSLMode:  "disable",
	// }

	dbConfig := &db.MySQLConfig{
		Host:     "localhost",
		User:     "admin",
		Password: "Str0ngP@ssw0rd!",
		DbName:   "plugmin",
		Port:     3306,
	}

	tableName := "user"

	queryParams := map[string]interface{}{}

	// requestBody := `{

	// }`

	requestBody := `{
	    "name":"Sagar P",
		"age": 30,
		"type":"free_user",
		"email": "sagar@example.com"
	}`

	// queryParams := map[string]interface{}{
	// 	"id": 10,
	// }

	method := "POST"

	HandleRequest(method, tableName, dbType, requestBody, dbConfig, queryParams)
}
