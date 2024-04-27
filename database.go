package goe

import (
	"database/sql"
	"fmt"
	"reflect"
)

// type Config struct {
// 	MigrationsPath string
// }

type conn struct {
	*sql.DB
}

type DB struct {
	conn    conn
	addrMap map[string]any
	//errors []error
	//Config
	//tables map[string]*table
}

func (db *DB) Open(name string, uri string) error {
	if db.conn.DB == nil {
		d, err := sql.Open(name, uri)
		if err == nil {
			db.conn.DB = d
		}
		return err
	}
	return nil
}

func (db *DB) Select(args ...any) Rows {

	stringArgs := make([]string, 0)
	for _, v := range args {
		if reflect.ValueOf(v).Kind() == reflect.Ptr {
			if reflect.ValueOf(v).Elem().Kind() == reflect.Struct {
				for i := 0; i < reflect.ValueOf(v).Elem().NumField(); i++ {
					stringArgs = append(stringArgs, fmt.Sprintf("%v", reflect.ValueOf(v).Elem().Field(i).Addr()))
				}
			}
		}
	}

	builder := createBuilder(querySELECT)
	builder.conn = db.conn
	builder.args = stringArgs

	return builder.buildSelect(db.addrMap)
}

// func (db *DB) Where(b boolean) Rows {
// 	return db
// }

func mapStructQuery(rows *sql.Rows, dest []any, value reflect.Value) (err error) {

	//TODO: add count for slices
	value.Set(reflect.MakeSlice(value.Type(), 10, 10))
	for i := 0; rows.Next(); i++ {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		s := reflect.New(value.Type().Elem()).Elem()
		//Fills the target
		for i, a := range dest {
			setValue(s.Field(i), a)
		}
		value.Index(i).Set(s)
	}
	return err
}

func mapQuery(rows *sql.Rows, dest []any, value reflect.Value) (err error) {
	value.Set(reflect.MakeSlice(value.Type(), 4, 10))

	for i := 0; rows.Next(); i++ {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		s := reflect.New(value.Type().Elem()).Elem()
		for _, a := range dest {
			setValue(s, a)
		}
		value.Index(i).Set(s)
	}
	return err
}

func setValue(v reflect.Value, a any) {
	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(string(*(a.(*sql.RawBytes))))
	}
}

func (db *DB) Equals(arg any, value any) *booleanResult {
	addr := fmt.Sprintf("%p", arg)

	//TODO: Add a return interface

	switch atr := db.addrMap[addr].(type) {
	case *att:
		return createBooleanResult(atr.name, atr.pk, value, EQUALS)
	case *pk:
		return createBooleanResult(atr.name, atr, value, EQUALS)
	}

	return nil
}

// func (db *DB) error(err error) bool {
// 	if err != nil {
// 		//db.errors = append(db.errors, err)
// 		return true
// 	}
// 	return false
// }

// func (db *DB) Errors() []error {
// 	return db.errors
// }
