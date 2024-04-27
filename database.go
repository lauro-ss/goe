package goe

import (
	"database/sql"
	"fmt"
	"reflect"
)

type conn struct {
	*sql.DB
}

type DB struct {
	conn    conn
	addrMap map[string]any
}

func (db *DB) open(name string, uri string) error {
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
			} else {
				stringArgs = append(stringArgs, fmt.Sprintf("%v", v))
			}
		} else {
			//TODO: Add ptr error
		}
	}

	state := createState(db.conn, querySELECT)

	return state.querySelect(stringArgs, db.addrMap)
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
