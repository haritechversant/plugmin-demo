DB_HOST="10.1.0.195"
DB_USER="tuneverse_user"
DB_PORT="5432"
DB_PASSWORD="S3cretPassWord"
DB_DATABASE="tuneverse_dev"

DB_HOST="127.0.0.1"
DB_USER="root"
DB_PORT="3306"
DB_PASSWORD="Str0ngP@ssw0rd!"
DB_DATABASE="tuneverse_dev"


CREATE USER 'my_user'@'localhost' IDENTIFIED BY 'Str0ngP@ssw0rd!';


{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/user.schema.json",
  "title": "User",
  "description": "Schema for user table",
  "type": "object",
  "properties": {
    "id": {
      "description": "Unique identifier for the user",
      "type": "string",
      "format": "uuid"
    },
    "name": {
      "description": "Name of the user",
      "type": "string"
    },
    "email": {
      "description": "Email address of the user",
      "type": "string",
      "format": "email"
    },
    "age": {
      "description": "Age of the user",
      "type": "integer",
      "minimum": 0
    }
  },
  "required": [
    "id",
    "name",
    "email"
  ],
  "additionalProperties": false
}

{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/product.schema.json",
  "title": "Product",
  "description": "Schema for adding a product with user reference",
  "type": "object",
  "properties": {
    "id": {
      "description": "Unique identifier for the product",
      "type": "integer",
      "value": 10
    },
    "product_name": {
      "description": "Name of the product",
      "type": "string",
      "value": "Example Product"
    },
    "user_id": {
      "description": "ID of the user associated with the product",
      "type": "string",
      "value": "e.g. 123e4567-e89b-12d3-a456-426614174000",
      "foreignKey": {
        "table": "user",
        "column": "id"
      }
    }
  },
  "required": [
    "product_id",
    "product_name",
    "user_id"
  ],
  "additionalProperties": false
}