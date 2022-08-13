// Lab 3 This is the server implementation for REST API
package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var errEmptyRow = errors.New("sql: Empty Row")
var errInvalidID = errors.New("foodMap: Invalid ID")

// SQL read cache for food data, adding this cache greatlyspeeds up HTTP GET operations, since SQL read is skipped.
var foodMap = map[string]productType{}

// GetProductRecordsInit gets all the rows to the Read Cache map
func GetProductRecordsInit(db *sql.DB) error {

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
func GetProductRecords(db *sql.DB) (*map[string]productType, error) {

	// // query to get all rows of table (persons) of my_db
	// rows, err := db.Query("Select * FROM Foods")
	// if err != nil {
	// 	//panic(err.Error())
	// 	fmt.Println("Error reading from SQL Food")
	// 	return nil, err
	// }
	// defer rows.Close()

	// // Declare an empty product map
	// //productList := []map[string]productType{}
	// productMap := make(map[string]interface{})

	// var id string

	// // extract row by row to create slice of productType
	// for rows.Next() {
	// 	// map this type to the record in the table
	// 	var product productType
	// 	err = rows.Scan(&id, &product.Category, &product.Name, &product.Weight, &product.Energy, &product.Protein,
	// 		&product.FatTotal, &product.FatSat, &product.Fibre, &product.Carb, &product.Cholesterol, &product.Sodium)

	// 	if err != nil {
	// 		//panic(err.Error())
	// 		fmt.Println("Error reading rows from SQL Food!!!")
	// 		//return &productList, err
	// 		return &productMap, err
	// 	}

	// 	// create map
	// 	// productMap[id] = product
	// 	foodMap[id] = product
	// 	//productList = append(productList, productMap)
	// }

	// //fmt.Println(productList)
	// fmt.Println(productMap)

	//return &productList, nil
	//	return &productMap, nil
	return &foodMap, nil
}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
func GetOneRecord(db *sql.DB, ID string) (*map[string]productType, error) {

	// // This should not be done this way to avaoid sql injection risk
	// // see https://go.dev/doc/database/sql-injection
	// //	query := fmt.Sprintf("Select * FROM foods where ID='%s'", id)

	// row, err := db.Query("Select * FROM foods where ID=?", ID)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer row.Close()

	// //var course map[string]string
	// foodMap := make(map[string]interface{})
	// var Id string

	// if row.Next() {
	// 	var food productType
	// 	err = row.Scan(&Id, &food.Category, &food.Name, &food.Weight,
	// 		&food.Energy, &food.Protein, &food.FatTotal, &food.FatSat,
	// 		&food.Fibre, &food.Carb, &food.Cholesterol, &food.Sodium)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	foodMap[Id] = food
	// 	fmt.Printf("Food: %+v\n", foodMap)
	// 	return foodMap, nil
	// } else {
	// 	return foodMap, errEmptyRow
	// }

	foodMapX := map[string]productType{}

	// check for validity of ID
	if v, ok := foodMap[ID]; ok {
		foodMapX[ID] = v
		return &foodMapX, nil
	} else {
		return &foodMapX, errInvalidID
	}

}

// Returns the number of rows that match the ID
func GetRowCount(db *sql.DB, ID string) (int, error) {

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
func DeleteRecord(db *sql.DB, ID string) {
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
func EditRecord(db *sql.DB, p *productType, ID string) {
	// create the sql query to update record
	// query := fmt.Sprintf("UPDATE foods SET Title='%s' WHERE ID='%s'", title, ID)
	// row, err := db.Query(query)
	row, err := db.Query("UPDATE foods SET Category=?, Name=?, Weight=?, Energy=?,Protein=?, FatTotal=?, FatSat=?, Fibre=?, Carb=?, Cholesterol=?, Sodium=? WHERE Id=?", p.Category, p.Name, p.Weight, p.Energy, p.Protein, p.FatTotal, p.FatSat, p.Fibre, p.Carb, p.Cholesterol, p.Sodium, ID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Update/Add Read Cache with Edited record
	foodMap[ID] = *p

	fmt.Println("Edit Successful")
}

// InsertRecord instead a row record into the current table based on the primary key and title
func InsertRecord(db *sql.DB, p *productType, ID string) {
	// create the sql query to insert record
	// note the quote for string
	// query := fmt.Sprintf("INSERT INTO foods VALUES ('%s', '%s')", ID, title)
	// _, err := db.Query(query)
	// Id is auto generated by SQL
	row, err := db.Query("INSERT INTO foods VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ID, p.Category, p.Name, p.Weight, p.Energy, p.Protein, p.FatTotal, p.FatSat, p.Fibre, p.Carb, p.Cholesterol, p.Sodium)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Add Read Cache with New record
	foodMap[ID] = *p

	fmt.Println("Insert Successful")
}
