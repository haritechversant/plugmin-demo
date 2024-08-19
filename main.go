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

func handlePatch(db *gorm.DB, requestData map[string]map[string]interface{}, ids map[string]interface{}) {
	for tableName, data := range requestData {
		columnVals, ok := data["columnVals"].(map[string]interface{})
		if !ok {
			log.Fatalf("Invalid columnVals for table: %s", tableName)
		}

		identityVal, hasIdentityVal := data["identityVal"].(map[string]interface{})

		refKeys, _ := data["referenceKey"].(map[string]interface{})
		for key, value := range refKeys {
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "$") {
				refColumn := strings.TrimPrefix(valueStr, "$")

				// First, check in identityVal
				if hasIdentityVal {
					if idVal, found := identityVal[refColumn]; found {
						columnVals[key] = idVal
					} else {
						// If not found in identityVal, then check in ids map
						if refVal, found := ids[refColumn]; found {
							columnVals[key] = refVal
						} else {
							log.Fatalf("Reference key %s not found in columnVals or identityVal", refColumn)
						}
					}
				} else if refVal, found := ids[refColumn]; found {
					columnVals[key] = refVal
				} else {
					log.Fatalf("Reference key %s not found in columnVals or identityVal", refColumn)
				}
			}
		}

		var setClauses []string
		var values []interface{}
		for col, val := range columnVals {
			if _, isRefKey := refKeys[col]; !isRefKey {
				setClauses = append(setClauses, fmt.Sprintf("%s = ?", QuoteIdentifier(col)))
				values = append(values, val)
			}
		}

		var whereClauses []string
		for col, val := range refKeys {
			if valueStr, ok := val.(string); ok && strings.HasPrefix(valueStr, "$") {
				refColumn := strings.TrimPrefix(valueStr, "$")
				if refVal, found := ids[refColumn]; found {
					whereClauses = append(whereClauses, fmt.Sprintf("%s IN (?)", QuoteIdentifier(col)))
					values = append(values, refVal)
				} else {
					log.Fatalf("Reference key %s not found in ids", refColumn)
				}
			} else {
				whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", QuoteIdentifier(col)))
				values = append(values, val)
			}
		}

		if hasIdentityVal {
			for col, val := range identityVal {
				whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", QuoteIdentifier(col)))
				values = append(values, val)
			}
		}

		sqlStatement := fmt.Sprintf("UPDATE %s SET %s WHERE %s RETURNING id",
			QuoteIdentifier(tableName),
			strings.Join(setClauses, ", "),
			strings.Join(whereClauses, " AND "))

		var updatedIDs []int64
		err := db.Raw(sqlStatement, values...).Scan(&updatedIDs).Error
		if err != nil {
			log.Fatalf("Error executing query for table %s: %v", tableName, err)
		}

		fmt.Printf("Successfully updated data in table: %s\nSQL: %s\nValues: %v\n", tableName, sqlStatement, values)

		ids[tableName+".id"] = updatedIDs

		if len(updatedIDs) > 1 {
			ids[tableName+".ids"] = updatedIDs
		} else {
			ids[tableName+".id"] = updatedIDs[0]
		}

		for _, nextData := range requestData {
			nextRefKeys, _ := nextData["referenceKey"].(map[string]interface{})
			for nextKey, nextValue := range nextRefKeys {
				if nextValueStr, ok := nextValue.(string); ok && strings.HasPrefix(nextValueStr, "$"+tableName+".id") {
					nextData["columnVals"].(map[string]interface{})[nextKey] = ids[tableName+".ids"]
				}
			}
		}
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
				"name": "amirtha vj",
				"email": "naew@gmail.com"
			},
			"identityVal": {
				"id": "7"
			}
		},
		"address_details": {
			"columnVals": {
				"address": "amirtha vj",
				"user_id": "user.id",
				"city": "saun"
			},
			"referenceKey": {
				"user_id": "$user.id"
			}
		},
		"profile": {
			"columnVals": {
				"work": "amirtha vijayan",
				"address": "address_details.id",
				"city": "abc"
			},
			"referenceKey": {
				"address": "$address_details.id"
			}
		}
	}`

	method := "PATCH"
	queryParams := make(map[string]interface{})

	HandleRequest(method, dbType, requestBody, dbConfig, queryParams)
}
