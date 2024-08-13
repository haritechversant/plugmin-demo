package main

import (
	"encoding/json"
	"fmt"
	"log"
	"orm-sample/db"
	"strings"
)

// QuoteIdentifier quotes a PostgreSQL identifier (table or column name)
func QuoteIdentifier(identifier string) string {
	return fmt.Sprintf(`"%s"`, identifier)
}

func HandleRequest(method, dbType, requestBody string, dbConfig db.DBConfig, queryParams map[string]interface{}) {
	if method != "POST" {
		log.Fatalf("Unsupported method: %s", method)
	}

	// Initialize the database connection
	db, err := db.InitDB(dbConfig, dbType)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	var requestData map[string]map[string]interface{}
	err = json.Unmarshal([]byte(requestBody), &requestData)
	if err != nil {
		log.Fatalf("Error parsing request body: %v", err)
	}

	// To store IDs of inserted rows for reference keys
	ids := make(map[string]interface{})

	for tableName, data := range requestData {
		columnVals, ok := data["columnVals"].(map[string]interface{})
		if !ok {
			log.Fatalf("Invalid columnVals for table: %s", tableName)
		}

		// Handle reference keys
		refKeys, _ := data["referenceKey"].(map[string]interface{})
		for key, value := range refKeys {
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "$") {
				// Extract value from previously inserted rows
				refColumn := strings.TrimPrefix(valueStr, "$")
				if refVal, found := ids[refColumn]; found {
					columnVals[key] = refVal
				} else {
					log.Fatalf("Reference key %s not found in columnVals", refColumn)
				}
			}
		}

		// Generate the SQL INSERT statement with quoted identifiers
		var columns []string
		var values []interface{}
		for col, val := range columnVals {
			columns = append(columns, QuoteIdentifier(col))
			values = append(values, val)
		}

		sqlStatement := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
			QuoteIdentifier(tableName),
			strings.Join(columns, ", "),
			strings.Repeat("?,", len(values)-1)+"?")

		var insertedID int64
		err := db.Raw(sqlStatement, values...).Scan(&insertedID).Error
		if err != nil {
			log.Fatalf("Error executing query for table %s: %v", tableName, err)
		}
		fmt.Printf("Successfully inserted data into table: %s with ID %d\n", tableName, insertedID)

		// Store the inserted ID for use in reference keys
		ids[tableName+".id"] = insertedID
	}
}
func main() {
	dbType := "postgres"

	dbConfig := &db.PostgresConfig{
		Host:     "10.1.0.195",
		User:     "tuneverse_user",
		Password: "S3cretPassWord",
		DbName:   "plugmin",
		Port:     5432,
		SSLMode:  "disable",
	}

	// dbType := "mysql"

	// dbConfig := &db.MySQLConfig{
	// 	Host:     "localhost",
	// 	User:     "admin",
	// 	Password: "Str0ngP@ssw0rd!",
	// 	DbName:   "plugmin",
	// 	Port:     3306,
	// }

	requestBody := `{
		"user": {
			"columnVals": {
				"name": "Aiswarya",
				"email": "annmit@ggmail.com"
			}
		},
		"address_details": {
			"columnVals": {
				"address": "thiruvathira",
				"user_id": "user.id",
				"city": "abc"
			},
			"referenceKey": {
				"user_id": "$user.id"
			}
		}
	}`

	method := "POST"
	queryParams := make(map[string]interface{})

	HandleRequest(method, dbType, requestBody, dbConfig, queryParams)
}
