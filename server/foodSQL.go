// Lab 3 This is the server implementation for REST API
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var errEmptyRow = errors.New("sql: Empty Row")
var errInvalidID = errors.New("foodMap: Invalid ID")
var errIllegalID = errors.New("foodMap: Illegal Zero ID")
var errInvalidKey = errors.New("foodMap: Invalid Key")
var errAlreadyUsedID = errors.New("foodMap: ID is already used")
var errIncompleteMapStruct = errors.New("foodMap:  Incomplete Map structure")

// SQL read cache for food data, adding this cache greatlyspeeds up HTTP GET operations, since SQL read is skipped.
var foodMap = map[string]foodType{}

// Mutex for critical section of SQL write to DB and cache
var m1 = sync.Mutex{}

func foodCacheInit() {
	db, err := sql.Open("mysql", cfgFood.FormatDSN())
	// handle error
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
	err = foodGetRecordsInitDB(db)
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
}

// foodGetRecordsInit gets all the rows to the Read Cache map
func foodGetRecordsInitDB(db *sql.DB) error {

	m1.Lock()
	defer m1.Unlock()

	// query to get all rows of table (persons) of my_db
	rows, err := db.Query("Select * FROM Foods")
	if err != nil {
		fmt.Println("Error initial reading from SQL Food")
		//panic(err.Error())
		return err
	}
	defer rows.Close()

	var id string

	// extract row by row to create slice of foodType
	for rows.Next() {
		// map this type to the record in the table
		var food foodType
		err = rows.Scan(&id, &food.Category, &food.Name, &food.Weight, &food.Energy, &food.Protein,
			&food.FatTotal, &food.FatSat, &food.Fibre, &food.Carb, &food.Cholesterol, &food.Sodium)

		if err != nil {
			fmt.Println("Error reading rows from SQL Food!!!")
			//panic(err.Error())
			return err
		}

		// initialise map
		foodMap[id] = food
	}

	//fmt.Printf("%+v", foodMap)
	return nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func foodGetRecords() (*map[string]foodType, error) {
	return &foodMap, nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func foodGetPrefixedRecords(prefix string) (*map[string]foodType, error) {

	// create a food map to be populated to match search
	selectFoodMap := map[string]foodType{}
	for k, v := range foodMap {
		if strings.HasPrefix(k, prefix) {
			selectFoodMap[k] = v
		}
	}

	if len(selectFoodMap) == 0 {
		return &selectFoodMap, fmt.Errorf("invalid prefix ID (%v)", prefix)
	}

	return &selectFoodMap, nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func foodfGetPrefixPostfixRecords(prefix string, postfix string) (*map[string]foodType, error) {

	// create a food map to be populated to match search
	selectFoodMap := map[string]foodType{}

	// If prefix and postfix are empty return complete list
	if prefix == "" && postfix == "" {
		return &foodMap, nil
	}

	// populate value that has prefix
	if prefix == "" {
		for k, v := range foodMap {
			selectFoodMap[k] = v
		}
	} else {
		for k, v := range foodMap {
			if strings.HasPrefix(k, prefix) {
				selectFoodMap[k] = v
			}
		}

		if len(selectFoodMap) == 0 {
			return &selectFoodMap, fmt.Errorf("invalid prefix ID (%v)", prefix)
		}
	}

	// remove key, value pair that that not have postfix key
	if postfix == "" {
		return &selectFoodMap, nil
	} else {
		for k, _ := range selectFoodMap {
			if !strings.HasSuffix(k, postfix) {
				delete(selectFoodMap, k)
			}
		}

		if len(selectFoodMap) == 0 {
			return &selectFoodMap, fmt.Errorf("invalid postfix ID (%v)", postfix)
		}
	}

	// fmt.Printf("selectFoodMap : %+v\n", selectFoodMap)

	return &selectFoodMap, nil
}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
func foodGetOneRecord(ID string) (*foodType, error) {

	if len(ID) == 0 {
		return nil, errIllegalID
	}

	// check for validity of ID
	if v, ok := foodMap[ID]; ok {
		return &v, nil
	} else {
		return &v, fmt.Errorf("invalid ID (%v)", ID)
	}

}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
func foodGetOneRecordDB(db *sql.DB, ID string) (*foodType, error) {

	row, err := db.Query("Select * FROM foods where ID=?", ID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	var food foodType
	var dummyId string

	if row.Next() {
		err = row.Scan(&dummyId, &food.Category, &food.Name, &food.Weight, &food.Energy, &food.Protein,
			&food.FatTotal, &food.FatSat, &food.Fibre, &food.Carb, &food.Cholesterol, &food.Sodium)
		if err != nil {
			panic(err.Error())
		}
		return &food, nil
	} else {
		return &food, errEmptyRow
	}
}

// Returns the number of rows that match the ID
func foodGetRowCountDB(db *sql.DB, ID string) (int, error) {

	row, err := db.Query("Select count(*) FROM foods where ID=?", ID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	if row.Next() {
		var count int
		err = row.Scan(&count)
		if err != nil {
			panic(err.Error())
		}
		return count, nil
	} else {
		return -1, errEmptyRow
	}
}

// DeleteRecord deletes a record from the current table using the ID primary key
func foodDeleteRecordDB(db *sql.DB, ID string) {
	m1.Lock()
	defer m1.Unlock()

	row, err := db.Query("DELETE FROM foods WHERE ID=?", ID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Delete Map Key from Read Cache
	delete(foodMap, ID)

	fmt.Println("Delete Successful")
}

// EditRecord edits the record of the current table based on the primary key ID with title
func foodEditRecordDB(db *sql.DB, newID string, f foodType, oldID string) error {
	m1.Lock()
	defer m1.Unlock()

	// Check if the New ID is to be updated is already used (except its own ID)
	//if err := foodValidateId(newID); err == nil {
	if _, err := foodGetOneRecord(newID); err == nil {
		if newID != oldID {
			return fmt.Errorf("new ID is in use (%v)", newID)
		}
	} else {
		// illegal Id, Id=""
		if err == errIllegalID {
			return err
		}
	}

	// validate tha old ID exist before update
	//if err := foodValidateId(oldID); err != nil {
	if _, err := foodGetOneRecord(oldID); err != nil {
		return fmt.Errorf("invalid Old ID (%v)", oldID)
	}

	row, err := db.Query("UPDATE foods SET ID=?, Category=?, Name=?, Weight=?, Energy=?,Protein=?, FatTotal=?, FatSat=?, Fibre=?, Carb=?, Cholesterol=?, Sodium=? WHERE Id=?",
		newID, f.Category, f.Name, f.Weight, f.Energy, f.Protein, f.FatTotal, f.FatSat, f.Fibre, f.Carb, f.Cholesterol, f.Sodium, oldID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Check if the ID is also updated
	if newID != oldID {
		// remove the old key and update with new keys in foodMap
		delete(foodMap, oldID)
	}
	foodMap[newID] = f

	fmt.Println("Edit Successful")

	return nil
}

func foodInsertRecordDB(db *sql.DB, fd foodType, ID string) error {
	m1.Lock()
	defer m1.Unlock()

	row, err := db.Query("INSERT INTO foods VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ID, fd.Category, fd.Name, fd.Weight, fd.Energy, fd.Protein, fd.FatTotal, fd.FatSat,
		fd.Fibre, fd.Carb, fd.Cholesterol, fd.Sodium)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Update Food Cache, Add Read Cache with New record
	foodMap[ID] = fd
	fmt.Printf("foodMap = %+v", foodMap[ID])

	fmt.Println("Insert Successful")
	return nil
}

// updateRecord with dynamic JSON map
func foodUpdateRecordDB(db *sql.DB, food mapInterface, keyRules mapInterface, oldID string) error {
	m1.Lock()
	defer m1.Unlock()

	// Initialise the food record first with original values
	foodTemp, err := foodGetOneRecord(oldID)
	if err != nil {
		return err
	}

	// Initialise with current ID
	var newId string = oldID
	var row *sql.Rows

	// Check if the New ID is to be updated is already used (except its own ID)
	if v, ok := food["Id"]; ok {
		newId = v.(string)
		//if err := foodValidateId(newId); err == nil {
		if _, err := foodGetOneRecord(newId); err == nil {
			if newId != oldID {
				return fmt.Errorf("new ID is in use (%v)", newId)
			}
		} else {
			// illegal Id, Id=""
			if err == errIllegalID {
				return err
			}
		}
	}

	// create base on food interface, no need to check count because it is dynamic
	_, err = updateFoodMapToStruct(foodTemp, food, keyRules)
	if err != nil {
		fmt.Println("Error :", err)
		return err
	}

	fmt.Printf("FoodTemp :%+v\n", *foodTemp)

	row, err = db.Query("Update FOODS SET Id=?, Category=?, Name=?, Weight=?, Energy=?, Protein=?, FatTotal=?, FatSat=?, Fibre=?, Carb=?, Cholesterol=?, Sodium=?  where ID=?",
		newId, (*foodTemp).Category, (*foodTemp).Name, (*foodTemp).Weight, (*foodTemp).Energy, (*foodTemp).Protein, (*foodTemp).FatTotal, (*foodTemp).FatSat, (*foodTemp).Fibre,
		(*foodTemp).Carb, (*foodTemp).Cholesterol, (*foodTemp).Sodium, oldID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Update food cache, Check if the ID is changed, then foodMap of respective Key is to be deleted
	if newId != oldID {
		delete(foodMap, oldID)
	}
	foodMap[newId] = *foodTemp

	//	fmt.Println("UserMap ", userMap)
	fmt.Println("Update Successful")
	return nil
}
