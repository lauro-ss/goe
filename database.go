package goe

import (
	"context"
	"fmt"
	"reflect"
)

type DB struct {
	ConnPool ConnectionPool
	addrMap  map[string]any
	driver   Driver
}

func (db *DB) Migrate(m *Migrator) {
	c, err := db.ConnPool.Conn(context.Background())
	if err != nil {
		//TODO: Add Error handler
		fmt.Println(err)
		return
	}
	db.driver.Migrate(m, c)
}

func (db *DB) Select(args ...any) *stateSelect {

	stringArgs := getArgs(args...)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createSelectState(conn, querySELECT)

	state.addrMap = db.addrMap
	return state.querySelect(stringArgs)
}

func (db *DB) Insert(table any) *stateInsert {
	stringArgs := getArgs(table)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createInsertState(conn, queryINSERT)

	return state.queryInsert(stringArgs, db.addrMap)
}

func (db *DB) InsertBetwent(table1 any, table2 any) *stateInsert {
	stringArgs := getArgs(table1, table2)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createInsertState(conn, queryINSERT)

	return state.queryInsertBetwent(stringArgs, db.addrMap)
}

func (db *DB) Update(tables ...any) *stateUpdate {
	stringArgs := getArgs(tables...)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createUpdateState(conn, queryUPDATE)

	return state.queryUpdate(stringArgs, db.addrMap)
}

func (db *DB) UpdateBetwent(table1 any, table2 any) *stateUpdateBetwent {
	stringArgs := getArgs(table1, table2)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createUpdateBetwentState(conn, queryUPDATE)

	return state.queryUpdateBetwent(stringArgs, db.addrMap)
}

func (db *DB) Delete(table any) *stateDelete {
	stringArgs := getArgs(table)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createDeleteState(conn, queryUPDATE)

	return state.queryDelete(stringArgs, db.addrMap)
}

func (db *DB) DeleteIn(table1 any, table2 any) *stateDeleteIn {
	stringArgs := getArgs(table1, table2)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createDeleteInState(conn, queryUPDATE)

	return state.queryDeleteIn(stringArgs, db.addrMap)
}

func (db *DB) Equals(arg any, value any) operator {
	addr := fmt.Sprintf("%p", arg)
	//TODO: add pointer validate
	switch atr := db.addrMap[addr].(type) {
	case *att:
		return createComplexOperator(atr.selectName, "=", value, atr.pk)
	case *pk:
		return createComplexOperator(atr.selectName, "=", value, atr)
	}

	return nil
}

func (db *DB) And() operator {
	return simpleOperator{operator: "AND"}
}

func (db *DB) Or() operator {
	return simpleOperator{operator: "OR"}
}

func getArgs(args ...any) []string {
	stringArgs := make([]string, 0)
	for _, v := range args {
		if reflect.ValueOf(v).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(v).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				for i := 0; i < reflect.ValueOf(v).Elem().NumField(); i++ {
					stringArgs = append(stringArgs, fmt.Sprintf("%p", reflect.ValueOf(v).Elem().Field(i).Addr().Interface()))
				}
			} else {
				stringArgs = append(stringArgs, fmt.Sprintf("%p", v))
			}
		} else {
			//TODO: Add ptr error
		}
	}
	return stringArgs
}
