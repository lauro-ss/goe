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

func (db *DB) Select(args ...any) Select {

	stringArgs := getArgs(args...)

	state := createSelectState(db.conn, querySELECT)

	return state.querySelect(stringArgs, db.addrMap)
}

func (db *DB) Insert(table any) Insert {
	stringArgs := getArgs(table)

	state := createInsertState(db.conn, queryINSERT)

	return state.queryInsert(stringArgs, db.addrMap)
}

func (db *DB) InsertBetwent(table1 any, table2 any) InsertBetwent {
	stringArgs := getArgs(table1, table2)

	state := createInsertState(db.conn, queryINSERT)

	return state.queryInsertBetwent(stringArgs, db.addrMap)
}

func (db *DB) Update(table any) Update {
	stringArgs := getArgs(table)

	state := createUpdateState(db.conn, queryUPDATE)

	return state.queryUpdate(stringArgs, db.addrMap)
}

func (db *DB) UpdateBetwent(table1 any, table2 any) Update {
	stringArgs := getArgs(table1, table2)

	state := createUpdateBetwentState(db.conn, queryUPDATE)

	return state.queryUpdateBetwent(stringArgs, db.addrMap)
}

func (db *DB) Equals(arg any, value any) *booleanResult {
	addr := fmt.Sprintf("%p", arg)

	//TODO: Add a return interface

	switch atr := db.addrMap[addr].(type) {
	case *att:
		return createBooleanResult(atr.selectName, atr.pk, value, whereEQUALS)
	case *pk:
		return createBooleanResult(atr.selectName, atr, value, whereEQUALS)
	}

	return nil
}

func (db *DB) And() *booleanResult {
	return createBooleanResult("", nil, "", whereAND)
}

func getArgs(args ...any) []string {
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
	return stringArgs
}
