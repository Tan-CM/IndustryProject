// Lab 3 This is the server implementation for REST API
package main

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// var errEmptyRow = errors.New("sql: Empty Row")
// var errInvalidID = errors.New("foodMap: Invalid ID")
// var errIllegalID = errors.New("foodMap: Illegal Zero ID")
// var errInvalidKey = errors.New("foodMap: Invalid Key")
// var errAlreadyUsedID = errors.New("foodMap: ID is already used")
// var errIncompleteMapStruct = errors.New("foodMap:  Incomplete Map structure")

// SQL read cache for food data, adding this cache greatlyspeeds up HTTP GET operations, since SQL read is skipped.
var dietProfMap = map[string]dietProfileType{}

// Mutex for critical section of SQL write to DB and cache
var m3 = sync.Mutex{}

func dietProfCacheInit() {
	db, err := sql.Open("mysql", cfgDietProf.FormatDSN())
	// handle error
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
	err = dietProfGetRecordsInit(db)
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
}

// foodGetRecordsInit gets all the rows to the Read Cache map
func dietProfGetRecordsInit(db *sql.DB) error {

	m3.Lock()
	defer m3.Unlock()

	// query to get all rows of table (persons) of my_db
	rows, err := db.Query("Select * FROM Profile")
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
		var dietProf dietProfileType
		err = rows.Scan(&id, &dietProf.Energy, &dietProf.Protein, &dietProf.FatTotal, &dietProf.FatSat,
			&dietProf.Fibre, &dietProf.Carb, &dietProf.Cholesterol, &dietProf.Sodium)

		if err != nil {
			fmt.Println("Error reading rows from SQL Food!!!")
			//panic(err.Error())
			return err
		}

		// initialise cache map
		dietProfMap[id] = dietProf
	}

	//fmt.Printf("%+v", dietProfMap)
	return nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func dietProfGetRecords() (*map[string]dietProfileType, error) {
	return &dietProfMap, nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func dietProfGetPrefixedRecords(prefix string) (*map[string]dietProfileType, error) {

	// create a food map to be populated to match search
	selectdietProfMap := map[string]dietProfileType{}
	for k, v := range dietProfMap {
		if strings.HasPrefix(k, prefix) {
			selectdietProfMap[k] = v
		}
	}

	if len(selectdietProfMap) == 0 {
		return &selectdietProfMap, fmt.Errorf("invalid prefix ID (%v)", prefix)
	}

	return &selectdietProfMap, nil
}

// GetRecords gets all the rows of the current table and return as a slice of map
func dietProfGetPrefixPostfixRecords(prefix string, postfix string) (*map[string]dietProfileType, error) {

	// create a food map to be populated to match search
	selectdietProfMap := map[string]dietProfileType{}

	// If prefix and postfix are empty return complete list
	if prefix == "" && postfix == "" {
		return &dietProfMap, nil
	}

	// populate value that has prefix
	if prefix == "" {
		for k, v := range dietProfMap {
			selectdietProfMap[k] = v
		}
	} else {
		for k, v := range dietProfMap {
			if strings.HasPrefix(k, prefix) {
				selectdietProfMap[k] = v
			}
		}

		if len(selectdietProfMap) == 0 {
			return &selectdietProfMap, fmt.Errorf("invalid prefix ID (%v)", prefix)
		}
	}

	// remove key, value pair that that not have postfix key
	if postfix == "" {
		return &selectdietProfMap, nil
	} else {
		for k, _ := range selectdietProfMap {
			if !strings.HasSuffix(k, postfix) {
				delete(selectdietProfMap, k)
			}
		}

		if len(selectdietProfMap) == 0 {
			return &selectdietProfMap, fmt.Errorf("invalid postfix ID (%v)", postfix)
		}
	}

	// fmt.Printf("selectdietProfMap : %+v\n", selectdietProfMap)

	return &selectdietProfMap, nil
}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = Either it has no record or the ID is illegal
func dietProfGetOneRecord(ID string) (*dietProfileType, error) {

	if len(ID) == 0 {
		return nil, errIllegalID
	}

	// check for validity of ID
	if v, ok := dietProfMap[ID]; ok {
		return &v, nil
	} else {
		return &v, fmt.Errorf("invalid ID (%v)", ID)
	}

}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
func dietProfGetOneRecordDB(db *sql.DB, ID string) (*dietProfileType, error) {

	row, err := db.Query("Select * FROM Profile where ID=?", ID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	var dietProf dietProfileType
	var dummyId string

	if row.Next() {
		err = row.Scan(&dummyId, &dietProf.Energy, &dietProf.Protein, &dietProf.FatTotal, &dietProf.FatSat,
			&dietProf.Fibre, &dietProf.Carb, &dietProf.Cholesterol, &dietProf.Sodium)
		if err != nil {
			panic(err.Error())
		}
		return &dietProf, nil
	} else {
		return &dietProf, errEmptyRow
	}
}

// Returns the number of rows that match the ID
func dietProfGetRowCount(db *sql.DB, ID string) (int, error) {

	row, err := db.Query("Select count(*) FROM Profile where ID=?", ID)
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
func dietProfDeleteRecord(db *sql.DB, ID string) {
	m3.Lock()
	defer m3.Unlock()

	row, err := db.Query("DELETE FROM Profile WHERE ID=?", ID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Delete Map Key from Read Cache
	delete(dietProfMap, ID)

	fmt.Println("Delete Successful")
}

// EditRecord edits the record of the current table based on the primary key ID with title
func dietProfEditRecord(db *sql.DB, newID string, fp dietProfileType, oldID string) error {
	m3.Lock()
	defer m3.Unlock()

	// Check if the New ID is to be updated is already used (except its own ID)
	//if err := foodValidateId(newID); err == nil {
	if _, err := dietProfGetOneRecord(newID); err == nil {
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
	if _, err := dietProfGetOneRecord(oldID); err != nil {
		return fmt.Errorf("invalid Old ID (%v)", oldID)
	}

	row, err := db.Query("UPDATE Profile SET ID=?, Energy=?,Protein=?, FatTotal=?, FatSat=?, Fibre=?, Carb=?, Cholesterol=?, Sodium=? WHERE Id=?",
		newID, fp.Energy, fp.Protein, fp.FatTotal, fp.FatSat, fp.Fibre, fp.Carb, fp.Cholesterol, fp.Sodium, oldID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Check if the ID is also updated
	if newID != oldID {
		// remove the old key and update with new keys in dietProfMap
		delete(dietProfMap, oldID)
	}
	dietProfMap[newID] = fp

	fmt.Println("Edit Successful")

	return nil
}

func dietProfInsertRecord(db *sql.DB, dp dietProfileType, ID string) error {
	m3.Lock()
	defer m3.Unlock()

	fmt.Println("Diet Profile :", dp)
	row, err := db.Query("INSERT INTO Profile VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ID, dp.Energy, dp.Protein, dp.FatTotal, dp.FatSat,
		dp.Fibre, dp.Carb, dp.Cholesterol, dp.Sodium)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Update Food Cache, Add Read Cache with New record
	dietProfMap[ID] = dp
	fmt.Printf("dietProfMap = %+v", dietProfMap[ID])

	fmt.Println("Insert Successful")
	return nil
}

// updateRecord with dynamic JSON map
func dietProfUpdateRecord(db *sql.DB, food mapInterface, keyRules mapInterface, oldID string) error {
	m3.Lock()
	defer m3.Unlock()

	// Initialise the food record first with original values
	dietProfile, err := dietProfGetOneRecord(oldID)
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
		if _, err := dietProfGetOneRecord(newId); err == nil {
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
	_, err = updateDietProfMapToStruct(dietProfile, food, keyRules)
	if err != nil {
		fmt.Println("Error :", err)
		return err
	}

	fmt.Printf("dietProfile :%+v\n", dietProfile)

	row, err = db.Query("Update Profile SET Id=?, Energy=?, Protein=?, FatTotal=?, FatSat=?, Fibre=?, Carb=?, Cholesterol=?, Sodium=?  where ID=?",
		newId, (*dietProfile).Energy, (*dietProfile).Protein, (*dietProfile).FatTotal, (*dietProfile).FatSat, (*dietProfile).Fibre,
		(*dietProfile).Carb, (*dietProfile).Cholesterol, (*dietProfile).Sodium, oldID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Update food cache, Check if the ID is changed, then dietProfMap of respective Key is to be deleted
	if newId != oldID {
		delete(dietProfMap, oldID)
	}
	dietProfMap[newId] = *dietProfile

	//	fmt.Println("UserMap ", userMap)
	fmt.Println("Update Successful")
	return nil
}
