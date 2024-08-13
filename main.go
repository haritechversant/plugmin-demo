package main

import (
	"encoding/json"
	"fmt"
	"log"
	"orm-sample/db"
	"strings"

	"gorm.io/gorm"
)

// QuoteIdentifier quotes a PostgreSQL identifier (table or column name)
func QuoteIdentifier(identifier string) string {
	return fmt.Sprintf(`"%s"`, identifier)
}

// HandleRequest processes database operations based on the HTTP method
func HandleRequest(method, dbType, requestBody string, dbConfig db.DBConfig, queryParams map[string]interface{}) {
	db, err := db.InitDB(dbConfig, dbType)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	var requestData map[string]map[string]interface{}
	err = json.Unmarshal([]byte(requestBody), &requestData)
	if err != nil {
		log.Fatalf("Error parsing request body: %v", err)
	}

	ids := make(map[string]interface{})

	switch method {
	case "POST":
		handlePost(db, requestData, ids)
	case "PATCH":
		handlePatch(db, requestData, ids)
	default:
		log.Fatalf("Unsupported method: %s", method)
	}
}

func handlePatch(db *gorm.DB, requestData map[string]map[string]interface{}, ids map[string]interface{}) {
	for tableName, data := range requestData {
		columnVals, ok := data["columnVals"].(map[string]interface{})
		if !ok {
			log.Fatalf("Invalid columnVals for table: %s", tableName)
		}

		refKeys, _ := data["referenceKey"].(map[string]interface{})

		// Debugging information
		fmt.Printf("IDs Map: %+v\n", ids)
		fmt.Printf("Reference Keys: %+v\n", refKeys)

		// Resolve reference keys
		for key, value := range refKeys {
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "$") {
				refColumn := strings.TrimPrefix(valueStr, "$")
				if refVal, found := ids[refColumn]; found {
					columnVals[key] = refVal
				} else {
					log.Fatalf("Reference key %s not found in ids", refColumn)
				}
			}
		}

		// Prepare SQL components
		var setClauses []string
		var values []interface{}
		var whereClauses []string

		id, idOk := columnVals["id"]

		if idOk {
			// ID is present, use it in WHERE clause
			for col, val := range columnVals {
				if col != "id" {
					setClauses = append(setClauses, fmt.Sprintf("%s = ?", QuoteIdentifier(col)))
					values = append(values, val)
				}
			}
			values = append(values, id)
			whereClauses = append(whereClauses, fmt.Sprintf("id = ?"))
		} else {
			// No ID provided, use reference keys for WHERE clause
			if refKeys != nil {
				for refKey, refValue := range refKeys {
					if refValueStr, ok := refValue.(string); ok && strings.HasPrefix(refValueStr, "$") {
						refColumn := strings.TrimPrefix(refValueStr, "$")
						if refVal, found := ids[refColumn]; found {
							whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", QuoteIdentifier(refKey)))
							values = append(values, refVal)
						} else {
							log.Fatalf("Reference key %s not found in ids", refColumn)
						}
					}
				}
			}

			if len(whereClauses) == 0 {
				log.Fatalf("ID is missing and no valid reference keys found for WHERE clause in table: %s", tableName)
			}

			// Set clauses without the ID field
			for col, val := range columnVals {
				if col != "id" {
					setClauses = append(setClauses, fmt.Sprintf("%s = ?", QuoteIdentifier(col)))
					values = append(values, val)
				}
			}
		}

		// Build SQL statement
		sqlStatement := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
			QuoteIdentifier(tableName),
			strings.Join(setClauses, ", "),
			strings.Join(whereClauses, " AND "))

		err := db.Exec(sqlStatement, values...).Error
		if err != nil {
			log.Fatalf("Error executing query for table %s: %v", tableName, err)
		}
		fmt.Printf("Successfully updated data in table: %s\n", tableName)
	}
}

// handlePost processes POST requests
func handlePost(db *gorm.DB, requestData map[string]map[string]interface{}, ids map[string]interface{}) {
	for tableName, data := range requestData {
		columnVals, ok := data["columnVals"].(map[string]interface{})
		if !ok {
			log.Fatalf("Invalid columnVals for table: %s", tableName)
		}

		refKeys, _ := data["referenceKey"].(map[string]interface{})
		for key, value := range refKeys {
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "$") {
				refColumn := strings.TrimPrefix(valueStr, "$")
				if refVal, found := ids[refColumn]; found {
					columnVals[key] = refVal
				} else {
					log.Fatalf("Reference key %s not found in columnVals", refColumn)
				}
			}
		}

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
				"name": "hhhaaarrriii",
				"email": "amiruutsha2@ggmail.com"
			}
		},
		"address_details": {
			"columnVals": {
				"address": "newaaaaaa",
				"user_id": "user.id",
				"city": "aaaaabc"
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
