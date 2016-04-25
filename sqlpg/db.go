package sqlpg

import (
	"database/sql"
	"dna"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"reflect"
)

var ErrNoRows = sql.ErrNoRows
var ErrTxDone = sql.ErrTxDone

// DB is a wrapper of sql.DB but some custom methods are added to enhance its functionalties.
type DB struct {
	*sql.DB
}

// Connect returns database of connected server.
// Returns an error if cannot connect to to server
func Connect(cf *SQLConfig) (*DB, error) {
	connectionString := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", cf.Username, cf.Password, cf.Host, cf.Post, cf.Database, cf.SSLMode)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	} else {
		pingErr := db.Ping()
		if pingErr != nil {
			return nil, errors.New(fmt.Sprintf("Cannot connect to database. Error: %s", pingErr.Error()))
		} else {
			return &DB{db}, nil
		}
	}
}

// QueryRecords is a reimplementation of sql.Db.Query().
// But it returns *Rows not *sql.Rows and error and takes query param as dna.String type
func (db *DB) Query(query dna.String, args ...interface{}) (*Rows, error) {
	rows, err := db.DB.Query(query.String(), args...)
	return &Rows{rows}, err
}

func (db *DB) QueryRow(query dna.String, args ...interface{}) *Row {
	row := db.DB.QueryRow(query.String(), args...)
	return &Row{row}
}

// Insert inserts custom struct to a table.
// The table's name depends on the type's name and its package's name.
// Ex: Any instance of type Song from package ns will be inserted into table nssongs.
// Insert returns an error if the struct fails.
//
// The error format is:
// 	Error description - $$$error$$$SQL_QUERY$$$error$$$
// Sql query is enclosed by `$$$error$$$`
func (db *DB) Insert(structValue interface{}) error {
	db.SetMaxIdleConns(-1)
	tbName := GetTableName(structValue)
	insertQuery := GetInsertStatement(tbName, structValue, false)
	_, err := db.Exec(insertQuery.String())
	if err != nil {
		str := dna.Sprintf("%s - $$$error$$$%v$$$error$$$", err.Error(), insertQuery).String()
		return errors.New(str)
	} else {
		return nil
	}

}

// InsertIgnore runs exactly the same as Insert.
// However if the insert value has already existed in a table,
// it does not return any error.
// It only returns an error if and only if other errors occur.
//
// The error format is the same as the one of sqlpg.DB.Insert()
func (db *DB) InsertIgnore(structValue interface{}) error {
	err := db.Insert(structValue)
	if err != nil {
		if dna.String(err.Error()).Contains(`duplicate key value violates unique constraint`) {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

// Update updates statement from GetUpdateStatment and returns error if available
//
// 	* structValue : A struct-typed value being scanned. Its fields have to be dna basic type or time.Time.
// 	* conditionColumn : A snake-case column name in the condition, usually it's an id
// 	* columns : A list of args of column names in the table being updated.
// 	* Returns an update statement.
// The error format is:
// 	Error description - $$$error$$$SQL_QUERY$$$error$$$
// Sql query is enclosed by `$$$error$$$`
func (db *DB) Update(structValue interface{}, conditionColumn dna.String, columns ...dna.String) error {
	tbName := GetTableName(structValue)
	updateQuery, err0 := GetUpdateStatement(tbName, structValue, conditionColumn, columns...)
	if err0 != nil {
		str := dna.Sprintf("%s - $$$error$$$%v$$$error$$$", err0.Error(), updateQuery).String()
		return errors.New(str)
	} else {
		_, err := db.Exec(updateQuery.String())
		if err != nil {
			str := dna.Sprintf("%s - $$$error$$$%v$$$error$$$", err.Error(), updateQuery).String()
			return errors.New(str)
		} else {
			return nil
		}
	}
}

// Select runs an arbitrary SQL query,
// binding the columns in the result to fields on the struct specified by structValue.
//
//	* structValue: A struct value being binded. It has to be a pointer to a slice
//	* query: A query statement
//	* args: The args are for any placeholder parameters in the query.
//	* Returns error if available.
//
// It supports only custom struct, basic dna types(dna.Int,dna.String..) and time.Time
//
// Basic usage : (See example below for exact implementation):
//
// 	// Custone Struct
// 	songs := &[]ns.Song{}
// 	err := db.Select(songs, "SELECT * FROM nssongs ORDER BY id ASC LIMIT 10")
// 	// Basic dna types
// 	ids := &[]Int{}
//	err := db.Select(ids, "SELECT id FROM nssongs ORDER BY id ASC LIMIT 10")
func (db *DB) Select(structValue interface{}, query dna.String, args ...interface{}) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			ptrStruct := reflect.ValueOf(structValue)
			if ptrStruct.Kind() == reflect.Ptr {
				realStruct := reflect.Indirect(ptrStruct)
				if realStruct.Kind() != reflect.Slice {
					return errors.New("Select() Method only accepts slice")
				} else {
					val := reflect.New(reflect.TypeOf(structValue).Elem().Elem())
					if reflect.Indirect(val).Kind() == reflect.Struct {
						rows.StructScan(val.Interface())
					} else {
						rows.Scan(val.Interface())
					}
					// Log(val.Interface())
					realStruct.Set(reflect.Append(realStruct, reflect.Indirect(val)))
				}
			} else {
				return errors.New("Select() Method only accepts pointer")
			}

		}
	}

	if rows != nil {
		rows.Close()
	}
	return nil
}
