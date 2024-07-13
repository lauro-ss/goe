package goe

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var ErrInvalidArg = errors.New("goe: invalid argument. try sending a pointer as argument")

type DB struct {
	ConnPool ConnectionPool
	addrMap  map[uintptr]field
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

func (db *DB) SelectContext(ctx context.Context, args ...any) *stateSelect {
	stringArgs, err := getArgs(db.addrMap, args...)

	var state *stateSelect
	if err != nil {
		state = createSelectState(nil, err)
		return state.querySelect(nil)
	}

	conn, err := db.ConnPool.Conn(ctx)
	state = createSelectState(conn, err)

	state.addrMap = db.addrMap
	return state.querySelect(stringArgs)
}

func (db *DB) Select(args ...any) *stateSelect {
	return db.SelectContext(context.Background(), args...)
}

func (db *DB) InsertContext(ctx context.Context, table any) *stateInsert {
	stringArgs, err := getArgs(db.addrMap, table)

	var state *stateInsert
	if err != nil {
		state = createInsertState(nil, err)
		return state.queryInsert(nil, nil)
	}

	conn, err := db.ConnPool.Conn(ctx)
	state = createInsertState(conn, err)

	return state.queryInsert(stringArgs, db.addrMap)
}

func (db *DB) Insert(table any) *stateInsert {
	return db.InsertContext(context.Background(), table)
}

func (db *DB) InsertInContext(ctx context.Context, table1 any, table2 any) *stateInsertIn {
	stringArgs, err := getArgsIn(table1, table2)

	var state *stateInsertIn
	if err != nil {
		state = createInsertStateIn(nil, err)
		return state.queryInsertIn(nil, nil)
	}

	conn, err := db.ConnPool.Conn(ctx)
	state = createInsertStateIn(conn, err)

	return state.queryInsertIn(stringArgs, db.addrMap)
}

func (db *DB) InsertIn(table1 any, table2 any) *stateInsertIn {
	return db.InsertInContext(context.Background(), table1, table2)
}

func (db *DB) UpdateContext(ctx context.Context, table any) *stateUpdate {
	stringArgs, err := getArgs(db.addrMap, table)

	var state *stateUpdate
	if err != nil {
		state = createUpdateState(nil, err)
		return state.queryUpdate(nil, nil)
	}
	conn, err := db.ConnPool.Conn(ctx)
	state = createUpdateState(conn, err)

	return state.queryUpdate(stringArgs, db.addrMap)
}

func (db *DB) Update(table any) *stateUpdate {
	return db.UpdateContext(context.Background(), table)
}

func (db *DB) UpdateInContext(ctx context.Context, table1 any, table2 any) *stateUpdateIn {
	stringArgs, err := getArgsIn(table1, table2)

	var state *stateUpdateIn
	if err != nil {
		state = createUpdateInState(nil, err)
		return state.queryUpdateIn(nil, nil)
	}
	conn, err := db.ConnPool.Conn(ctx)
	state = createUpdateInState(conn, err)

	return state.queryUpdateIn(stringArgs, db.addrMap)
}

func (db *DB) UpdateIn(table1 any, table2 any) *stateUpdateIn {
	return db.UpdateInContext(context.Background(), table1, table2)
}

func (db *DB) DeleteContext(ctx context.Context, table any) *stateDelete {
	stringArgs, err := getArgs(db.addrMap, table)

	var state *stateDelete
	if err != nil {
		state = createDeleteState(nil, err)
		return state.queryDelete(nil, nil)
	}
	conn, err := db.ConnPool.Conn(ctx)
	state = createDeleteState(conn, err)

	return state.queryDelete(stringArgs, db.addrMap)
}

func (db *DB) Delete(table any) *stateDelete {
	return db.DeleteContext(context.Background(), table)
}

func (db *DB) DeleteInContext(ctx context.Context, table1 any, table2 any) *stateDeleteIn {
	stringArgs, err := getArgsIn(table1, table2)

	var state *stateDeleteIn
	if err != nil {
		state = createDeleteInState(nil, err)
		return state.queryDeleteIn(nil, nil)
	}
	conn, err := db.ConnPool.Conn(ctx)
	state = createDeleteInState(conn, err)

	return state.queryDeleteIn(stringArgs, db.addrMap)
}

func (db *DB) DeleteIn(table1 any, table2 any) *stateDeleteIn {
	return db.DeleteInContext(context.Background(), table1, table2)
}

func (db *DB) Equals(arg any, value any) operator {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		return nil
	}

	addr := uintptr(v.UnsafePointer())
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

func getArgs(addrMap map[uintptr]field, args ...any) ([]uintptr, error) {
	stringArgs := make([]uintptr, 0)
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
					addr := uintptr(fieldOf.Addr().UnsafePointer())
					if addrMap[addr] != nil {
						stringArgs = append(stringArgs, addr)
					}
				}
			} else {
				stringArgs = append(stringArgs, uintptr(valueOf.Addr().UnsafePointer()))
			}
		} else {
			return nil, ErrInvalidArg
		}
	}
	return stringArgs, nil
}

func getArgsIn(args ...any) ([]uintptr, error) {
	stringArgs := make([]uintptr, 2)
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				stringArgs[i] = uintptr(reflect.ValueOf(args[i]).Elem().Field(0).Addr().UnsafePointer())
			} else {
				stringArgs[i] = uintptr(valueOf.Addr().UnsafePointer())
			}
		} else {
			return nil, ErrInvalidArg
		}
	}
	return stringArgs, nil
}
