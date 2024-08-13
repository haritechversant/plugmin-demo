package utils

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TableInfo struct to hold schema details
type TableInfo struct {
	SchemaName        string   `json:"schemaName"`
	TableName         string   `json:"tableName"`
	ColumnName        string   `json:"columnName"`
	DataType          string   `json:"dataType"`
	IsNullable        string   `json:"isNullable"`
	OrdinalPosition   int      `json:"ordinalPosition"`
	ColumnComment     string   `json:"columnComment"`
	DefaultValue      string   `json:"defaultValue"`
	IndexName         string   `json:"indexName"`
	IsUnique          bool     `json:"isUnique"`
	IsPrimary         bool     `json:"isPrimary"`
	IndexColumnName   string   `json:"indexColumnName"`
	ConstraintName    string   `json:"constraintName"`
	ConstraintType    string   `json:"constraintType"`
	ForeignTableName  string   `json:"foreignTableName"`
	ForeignColumnName string   `json:"foreignColumnName"`
	AutoIncrement     bool     `json:"autoIncrement"`
	CompositeKey      []string `json:"compositeKey,omitempty"` // `omitempty` to ignore empty slices
}

func FetchTableSchema(db *gorm.DB, tableName string) ([]TableInfo, error) {
	var schemaInfo []TableInfo
	var compositeKeys map[string][]string

	// Query to get schema details without composite keys
	query := `
		SELECT 
			c.table_schema AS schema_name,
			c.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.ordinal_position,
			c.column_default AS default_value,
			'' AS column_comment,
			coalesce(tc.constraint_name, '') AS constraint_name,
			coalesce(tc.constraint_type, '') AS constraint_type,
			coalesce(kcu.constraint_name, '') AS index_name,
			coalesce(tc.constraint_type = 'PRIMARY KEY', false) AS is_primary,
			coalesce(tc.constraint_type = 'UNIQUE', false) AS is_unique,
			coalesce(kcu.column_name, '') AS index_column_name,
			coalesce(ccu.table_name, '') AS foreign_table_name,
			coalesce(ccu.column_name, '') AS foreign_column_name,
			false AS auto_increment
		FROM 
			information_schema.columns c
		LEFT JOIN 
			information_schema.key_column_usage kcu 
			ON c.table_name = kcu.table_name 
			AND c.column_name = kcu.column_name
		LEFT JOIN 
			information_schema.table_constraints tc 
			ON kcu.constraint_name = tc.constraint_name 
			AND tc.table_name = c.table_name
		LEFT JOIN 
			information_schema.constraint_column_usage ccu 
			ON tc.constraint_name = ccu.constraint_name
		WHERE 
			c.table_name = $1
		ORDER BY 
			c.ordinal_position;
	`

	err := db.Raw(query, tableName).Scan(&schemaInfo).Error
	if err != nil {
		return nil, err
	}

	// Fetch composite key details
	compositeKeyQuery := `
		SELECT
			tc.constraint_name,
			kcu.column_name
		FROM 
			information_schema.table_constraints tc
		JOIN 
			information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE 
			tc.table_name = $1
			AND tc.constraint_type = 'PRIMARY KEY'
		ORDER BY
			tc.constraint_name, kcu.ordinal_position;
	`

	var keyDetails []struct {
		ConstraintName string
		ColumnName     string
	}

	err = db.Raw(compositeKeyQuery, tableName).Scan(&keyDetails).Error
	if err != nil {
		return nil, err
	}

	// Build the composite key map
	compositeKeys = make(map[string][]string)
	for _, key := range keyDetails {
		compositeKeys[key.ConstraintName] = append(compositeKeys[key.ConstraintName], key.ColumnName)
	}

	// Assign composite key values to schemaInfo and filter out empty CompositeKey
	for i := range schemaInfo {
		if keyCols, exists := compositeKeys[schemaInfo[i].ConstraintName]; exists {
			schemaInfo[i].CompositeKey = keyCols
		} else {
			schemaInfo[i].CompositeKey = nil // Ensures empty arrays are omitted
		}
	}

	return schemaInfo, nil
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	tableName := "user"

	tableInfos, err := FetchTableSchema(db, tableName)
	if err != nil {
		log.Fatalf("failed to fetch table schema: %v", err)
	}

	for _, info := range tableInfos {
		fmt.Printf("%+v\n\n", info)
	}
}
