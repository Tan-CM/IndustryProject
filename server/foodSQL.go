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
var errInvalidKey = errors.New("foodMap: Invalid Key")
var errAlreadyUsedID = errors.New("foodMap: ID is already used")
var errIncompleteMapStruct = errors.New("foodMap:  Incomplete Map structure")

// SQL read cache for food data, adding this cache greatlyspeeds up HTTP GET operations, since SQL read is skipped.
var foodMap = map[string]productType{}

// Mutex for critical section of SQL write to DB and cache
var m1 = sync.Mutex{}

// GetProductRecordsInit gets all the rows to the Read Cache map
func getProductRecordsInit(db *sql.DB) error {

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

	// extract row by row to create slice of productType
	for rows.Next() {
		// map this type to the record in the table
		var product productType
		err = rows.Scan(&id, &product.Category, &product.Name, &product.Weight, &product.Energy, &product.Protein,
			&product.FatTotal, &product.FatSat, &product.Fibre, &product.Carb, &product.Cholesterol, &product.Sodium)

		if err != nil {
			fmt.Println("Error reading rows from SQL Food!!!")
			//panic(err.Error())
			return err
		}

		// initialise map
		foodMap[id] = product
	}

	//fmt.Printf("%+v", foodMap)
	return nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func getProductRecords() (*map[string]productType, error) {
	return &foodMap, nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func getPrefixedRecords(prefix string) (*map[string]productType, error) {

	// create a food map to be populated to match search
	selectFoodMap := map[string]productType{}
	for k, v := range foodMap {
		if strings.HasPrefix(k, prefix) {
			selectFoodMap[k] = v
		}
	}

	if len(selectFoodMap) != 0 {
		return &selectFoodMap, nil
	}
	return &selectFoodMap, errInvalidID
}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
func getOneRecord(ID string) (*map[string]productType, error) {

	foodMapX := map[string]productType{}

	// check for validity of ID
	if v, ok := foodMap[ID]; ok {
		foodMapX[ID] = v
		return &foodMapX, nil
	} else {
		return &foodMapX, errInvalidID
	}

}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
func getOneRecordDB(db *sql.DB, ID string) (*productType, error) {

	row, err := db.Query("Select * FROM foods where ID=?", ID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	var food productType
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
func getRowCount(db *sql.DB, ID string) (int, error) {

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
func deleteRecord(db *sql.DB, ID string) {
	m1.Lock()
	defer m1.Unlock()

	// create the sql query to delete with primary key
	// Note deleting a non-existent record is considered as deleted, so will always passed

	//query := fmt.Sprintf("DELETE FROM foods WHERE ID='%s'", ID)
	//row, err := db.Query(query)
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
func editRecord(db *sql.DB, newID string, f productType, oldID string) error {
	m1.Lock()
	defer m1.Unlock()

	if len(newID) == 0 {
		return errInvalidID
	}

	if _, ok := foodMap[newID]; ok && newID != oldID {
		fmt.Printf("Error JSON new ID Error - %+v is already used\n", newID)
		return fmt.Errorf("New ID (%v) is already in Used", newID)
	}

	if _, ok := foodMap[oldID]; !ok {
		fmt.Printf("Error JSON old ID Error - %+v is invalid\n", oldID)
		return fmt.Errorf("ID (%v) is invalid", oldID)
	}

	row, err := db.Query("UPDATE foods SET ID=?, Category=?, Name=?, Weight=?, Energy=?,Protein=?, FatTotal=?, FatSat=?, Fibre=?, Carb=?, Cholesterol=?, Sodium=? WHERE Id=?",
		newID, f.Category, f.Name, f.Weight, f.Energy, f.Protein, f.FatTotal, f.FatSat, f.Fibre, f.Carb, f.Cholesterol, f.Sodium, oldID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// get updated food record from DB
	updatedFood, err := getOneRecordDB(db, newID)
	if err != nil {
		panic(err.Error())
	}

	// Check if the ID is also updated
	if newID != oldID {
		// remove the old key and update with new keys in foodMap
		delete(foodMap, oldID)
	}
	foodMap[newID] = *updatedFood

	fmt.Println("Edit Successful")

	return nil
}

func insertRecord(db *sql.DB, fd productType, ID string) error {
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
func updateRecord(db *sql.DB, food mapInterface, keyRules mapInterface, oldID string) error {
	m1.Lock()
	defer m1.Unlock()

	// Initialise the food record first with original values
	foodTemp, err := getOneRecordDB(db, oldID)
	if err != nil {
		panic(err.Error())
	}

	// Initialise with current ID
	var newId string = oldID
	var row *sql.Rows

	// Check if the New ID is to be updated is already used (except its own ID)
	if v, ok := food["Id"]; ok {
		newId = v.(string)

		// check if new Id is used excluding the current Id
		if _, ok := foodMap[newId]; ok && newId != oldID {
			fmt.Printf("Error ID - %+v is already used\n", v)
			return fmt.Errorf("ID (%v) is already in Used", newId)
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
