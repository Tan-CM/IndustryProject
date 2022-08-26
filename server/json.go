package main

import (
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

type mapInterface map[string]interface{}

// validate JSON object to validate the interface type
// key must match
// value type must match
// value string must not be empty
func validateKeysValues(mapX mapInterface, keys map[string]string) bool {
	// validate JSON data is correct

	// empty key on map
	if len(mapX) == 0 {
		return false
	}

	var countKeys, countValueTypes int
	// validate keys
	for k, v := range mapX {
		// validate keys
		if vt, ok := keys[k]; ok {
			fmt.Printf("Key : %+v\n", k)
			countKeys++
			// validate value Type
			if reflect.TypeOf(v).Name() == vt {
				fmt.Printf("ValueType : %+v\n", reflect.TypeOf(v))

				if vt != "string" {
					countValueTypes++
				} else {
					// string type must not be empty
					if len(v.(string)) != 0 {
						countValueTypes++
					}
				}

			}
		}
	}

	// mismatch keys
	if len(mapX) != countKeys {
		return false
	}
	// mismatch value types
	if len(mapX) != countValueTypes {
		return false
	}
	return true
}
