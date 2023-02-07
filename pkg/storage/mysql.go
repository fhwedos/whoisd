// Copyright 2017 Openprovider Authors. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// simplest logger, which initialized during starts of the application
var (
	mylog = log.New(os.Stdout, "[MYSQL]: ", log.Ldate|log.Ltime)
)

// MysqlRecord - standard record (struct) for mysql storage package
type MysqlRecord struct {
	Host     string
	Port     int
	User     string
	Password string
	DataBase string
	Table    string
}

// Search data in the storage
func (mysql *MysqlRecord) Search(name []string, query []string) (map[string][]string, error) {
	result, err := mysql.searchRaw(mysql.Table, name, query)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result[0], nil
	}

	data := make(map[string][]string) // empty result
	return data, nil
}

// SearchRelated - search data in the storage from related type or table
func (mysql *MysqlRecord) SearchRelated(
	typeTable string, name []string, query []string) (map[string][]string, error) {

	result, err := mysql.searchRaw(typeTable, name, query)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result[0], nil
	}

	data := make(map[string][]string)
	return data, nil
}

// SearchMultiple - search multiple records of data in the storage
func (mysql *MysqlRecord) SearchMultiple(
	typeTable string, name []string, query []string) (map[string][]string, error) {

	result, err := mysql.searchRaw(typeTable, name, query)
	if err != nil {
		return nil, err
	}

	data := make(map[string][]string)

	if len(result) > 0 {
		for _, item := range result {
			for key, value := range item {
				data[key] = append(data[key], value...)
			}
		}
		return data, nil
	}

	return data, nil
}

func (mysql *MysqlRecord) searchRaw(typeTable string, name []string, query []string) ([]map[string][]string, error) {
	// Thanks to https://github.com/go-sql-driver/mysql/wiki/Examples#rawbytes
	db, err := sql.Open("mysql", mysql.User+":"+mysql.Password+
		"@tcp("+mysql.Host+":"+strconv.Itoa(mysql.Port)+")/"+
		mysql.DataBase+"?charset=utf8")

	if err != nil {
		return nil, fmt.Errorf("Mysql connection error: %v", err)
	}

	defer db.Close()

	where := []string{}

	for i := 0; i < len(name); i++ {
		// Filter input
		name[i] = filterString(name[i], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
		query[i] = filterString(query[i], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-")

		where = append(where, name[i]+"=\""+query[i]+"\"")
	}

	dbQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s", typeTable, strings.Join(where, " AND "))

	mylog.Println("QUERY:", dbQuery)

	f, err := os.OpenFile("mysql_log_file", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println(dbQuery)

	// Execute the query
	rows, err := db.Query(dbQuery)
	if err != nil {
		return nil, fmt.Errorf("Mysql query error: %v", err)
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("Mysql column name query error: %v", err)
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	var data []map[string][]string
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, fmt.Errorf("Mysql scan error: %v", err)
		}

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		element := make(map[string][]string)
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "n/a"
			} else {
				value = string(col)
			}
			element[columns[i]] = []string{value}
		}
		data = append(data, element)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Mysql row read error: %v", err)
	}

	return data, nil
}

// Based on http://rosettacode.org/wiki/Strip_a_set_of_characters_from_a_string#Go
func filterString(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) >= 0 {
			return r
		}
		return -1
	}, str)
}
