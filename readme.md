
########### POST Reqbody ###############

requestBody := `{
		"user": {
			"columnVals": {
				"name": "chella",
				"email": "chella@gmail.com"
			},
			"identityVal": {
				"id": "5"
			}
		},
		"address_details": {
			"columnVals": {
				"address": "thirdOnam",
				"user_id": "user.id",
				"city": "tvm"
			},
			"referenceKey": {
				"user_id": "$user.id"
			}
		}
	}`

#################  PATCH Request Body #################

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