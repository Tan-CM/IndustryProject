package main

import (
	"errors"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

var errInvalidMapKey = errors.New("validation; Invalid Map Key in Body")
var errInvalidMapInterface = errors.New("validation: Invalid Map Interface")

// var errInvalidDataType = errors.New("validation: Invalid Map Data Type Value")
// var errIllegalStruct = errors.New("validation: Invalid Structure")

type mapInterface map[string]interface{}

// validate JSON object to validate the interface type
// Can only validate map because it is a interface{} type
// key must match
// value type must match
// value string must not be empty
func validateKeysValueTypes(mapX map[string]interface{}, keys map[string]string) (bool, error) {

	// empty key on map
	if len(mapX) == 0 {
		return false, errInvalidMapInterface
	}

	// var countKeys, countValueTypes int
	// validate keys
	for k, v := range mapX {
		// validate keys
		if vt, ok := keys[k]; ok {
			fmt.Printf("Key : %+v\n", k)
			//countKeys++
			// validate value Type
			if reflect.TypeOf(v).Name() == vt {
				fmt.Printf("ValueType : %+v\n", reflect.TypeOf(v))

			} else {
				fmt.Println("Error : Key Types Mismatch")
				return false, fmt.Errorf("validation Failed : Key Types (%v) : Given: %v, Expect: %v", k, reflect.TypeOf(v), vt)
			}
		} else {
			fmt.Println("Error: Invalid Key")
			return false, fmt.Errorf("validation Failed : Invalid Key : (%v)", k)
		}
	}

	return true, nil
}

func buildVMapTemplate(mapX map[string]interface{}, rules map[string]interface{}) (*map[string]interface{}, error) {

	// empty key on map
	if len(mapX) == 0 {
		return nil, errInvalidMapInterface
	}

	var template = make(map[string]interface{})

	for k, _ := range mapX {
		if v, ok := foodMapRules[k]; ok {
			template[k] = v
		} else {
			return &template, errInvalidMapKey
		}
	}
	return &template, nil
}
