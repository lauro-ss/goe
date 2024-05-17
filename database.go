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
	config  Config
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

type migrateTable struct {
	pk       *pk
	atts     []*att
	migrated bool
}

func (db *DB) Migrate() {
	tables := make(map[string]*migrateTable, 0)
	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if tables[atr.table] == nil {
				tables[atr.table] = newMigrateTable(db, atr.table)
			}
		}
	}

	for _, t := range tables {
		generateSql(t, tables)
	}
}

func newMigrateTable(db *DB, tableName string) *migrateTable {
	table := new(migrateTable)
	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if atr.table == tableName {
				table.pk = atr
			}
		case *att:
			if atr.pk.table == tableName {
				table.atts = append(table.atts, atr)
			}
		}
	}

	return table
}

func generateSql(mt *migrateTable, tables map[string]*migrateTable) {
	if !mt.migrated {
		fmt.Println("CREATE TABLE " + mt.pk.table)
	}
}

func (db *DB) Select(args ...any) Select {

	stringArgs := getArgs(args...)

	state := createSelectState(db.conn, querySELECT)
	state.addrMap = db.addrMap
	return state.querySelect(stringArgs)
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

func (db *DB) Update(tables ...any) Update {
	stringArgs := getArgs(tables...)

	state := createUpdateState(db.conn, queryUPDATE)

	return state.queryUpdate(stringArgs, db.addrMap)
}

func (db *DB) UpdateBetwent(table1 any, table2 any) Update {
	stringArgs := getArgs(table1, table2)

	state := createUpdateBetwentState(db.conn, queryUPDATE)

	return state.queryUpdateBetwent(stringArgs, db.addrMap)
}

func (db *DB) Delete(table any) Delete {
	stringArgs := getArgs(table)

	state := createDeleteState(db.conn, queryUPDATE)

	return state.queryDelete(stringArgs, db.addrMap)
}

func (db *DB) Equals(arg any, value any) operator {
	addr := fmt.Sprintf("%p", arg)

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
