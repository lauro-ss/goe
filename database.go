package goe

import (
	"context"
	"fmt"
	"reflect"
)

type DB struct {
	ConnPool ConnectionPool
	addrMap  map[string]field
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

	stringArgs := getArgs(db.addrMap, args...)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createSelectState(conn)

	state.addrMap = db.addrMap
	return state.querySelect(stringArgs)
}

func (db *DB) Insert(table any) *stateInsert {
	stringArgs := getArgs(db.addrMap, table)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createInsertState(conn)

	return state.queryInsert(stringArgs, db.addrMap)
}

func (db *DB) InsertIn(table1 any, table2 any) *stateInsertIn {
	stringArgs := getArgsIn(table1, table2)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createInsertStateIn(conn)

	return state.queryInsertIn(stringArgs, db.addrMap)
}

func (db *DB) Update(tables any) *stateUpdate {
	stringArgs := getArgs(db.addrMap, tables)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createUpdateState(conn)

	return state.queryUpdate(stringArgs, db.addrMap)
}

func (db *DB) UpdateIn(table1 any, table2 any) *stateUpdateIn {
	stringArgs := getArgsIn(table1, table2)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createUpdateInState(conn)

	return state.queryUpdateIn(stringArgs, db.addrMap)
}

func (db *DB) Delete(table any) *stateDelete {
	stringArgs := getArgs(db.addrMap, table)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createDeleteState(conn)

	return state.queryDelete(stringArgs, db.addrMap)
}

func (db *DB) DeleteIn(table1 any, table2 any) *stateDeleteIn {
	stringArgs := getArgsIn(table1, table2)

	//TODO: add ctx
	conn, _ := db.ConnPool.Conn(context.Background())
	state := createDeleteInState(conn)

	return state.queryDeleteIn(stringArgs, db.addrMap)
}

func (db *DB) Equals(arg any, value any) operator {
	addr := fmt.Sprintf("%p", arg)
	if db.addrMap[addr] == nil {
		//TODO: Add error
		return nil
	}
	return db.addrMap[addr].buildComplexOperator("=", value)
}

func (db *DB) And() operator {
	return simpleOperator{operator: "AND"}
}

func (db *DB) Or() operator {
	return simpleOperator{operator: "OR"}
}

func getArgs(addrMap map[string]field, args ...any) []string {
	stringArgs := make([]string, 0)
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				var fieldOf reflect.Value
				for i := 0; i < valueOf.NumField(); i++ {
					fieldOf = valueOf.Field(i)
					if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
						continue
					}
					addr := fmt.Sprintf("%p", fieldOf.Addr().Interface())
					if addrMap[addr] != nil {
						stringArgs = append(stringArgs, addr)
					}
				}
			} else {
				stringArgs = append(stringArgs, fmt.Sprintf("%p", args[i]))
			}
		} else {
			//TODO: Add ptr error
		}
	}
	return stringArgs
}

func getArgsIn(args ...any) []string {
	stringArgs := make([]string, 2)
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				stringArgs[i] = fmt.Sprintf("%p", reflect.ValueOf(args[i]).Elem().Field(0).Addr().Interface())
			} else {
				stringArgs[i] = fmt.Sprintf("%p", args[i])
			}
		} else {
			//TODO: Add ptr error
		}
	}
	return stringArgs
}
