package goe

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
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
	atts     map[string]any
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
		generateSql(db, t, tables)
	}
}

func newMigrateTable(db *DB, tableName string) *migrateTable {
	table := new(migrateTable)
	table.atts = make(map[string]any)
	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if atr.table == tableName {
				table.pk = atr
				for _, fkAny := range atr.fks {
					switch fk := fkAny.(type) {
					case *manyToOne:
						table.atts[strings.Split(fk.id, ".")[1]] = fk
					}
				}
			}
		case *att:
			if atr.pk.table == tableName {
				table.atts[atr.attributeName] = atr
			}
		}
	}

	return table
}

type databaseTable struct {
	columnName   string
	dataType     string
	defaultValue *string
	nullable     bool
}

func generateSql(db *DB, mt *migrateTable, tables map[string]*migrateTable) {
	sqlTableInfos := `SELECT
	column_name, CASE 
	WHEN data_type = 'character varying' 
	THEN CONCAT('varchar','(',character_maximum_length,')')
	WHEN data_type = 'text'
	THEN 'string'
	WHEN data_type = 'boolean'
	THEN 'bool'
	ELSE data_type END, 
	column_default, 
	CASE
	WHEN is_nullable = 'YES'
	THEN True
	ELSE False END AS is_nullable
	FROM information_schema.columns WHERE table_name = $1;	
	`

	if !mt.migrated {
		rows, err := db.conn.Query(sqlTableInfos, mt.pk.table)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rows.Close()

		dts := make([]databaseTable, 0)
		dt := databaseTable{}
		for rows.Next() {
			err = rows.Scan(&dt.columnName, &dt.dataType, &dt.defaultValue, &dt.nullable)
			if err != nil {
				fmt.Println(err)
				return
			}
			dts = append(dts, dt)
		}
		if len(dts) > 0 {
			//check fileds
			for i := range dts {
				checkFields(dts[i], mt, tables)
			}
		} else {
			fmt.Println("CREATE TABLE " + mt.pk.table)
		}
	}
}

func checkFields(databaseTable databaseTable, mt *migrateTable, tables map[string]*migrateTable) {
	if databaseTable.columnName == mt.pk.attributeName {
		// primary key field
		//fmt.Println(databaseTable, mt.pk)
		return
	}
	attrAny := mt.atts[databaseTable.columnName]
	if attrAny == nil {
		//fmt.Printf("goe:field '%v'exists on database table but is missed on struct %v\n", databaseTable.columnName, mt.pk.table)
		//prompt a question to drop or set as a rename field
	}
	switch attr := attrAny.(type) {
	case *att:
		if databaseTable.dataType != attr.dataType {
			//fmt.Println(alterColumn(attr.pk.table, databaseTable.columnName, attr.dataType))
		}
		if databaseTable.nullable != attr.nullable {
			fmt.Println(nullableColumn(attr.pk.table, attr.attributeName, attr.nullable))
		}
	}
}

func alterColumn(table, column, dataType string) string {
	dataMap := map[string]string{
		"string": "text",
	}
	if dataMap[dataType] == "" {
		return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v", table, column, dataType)
	}
	return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v", table, column, dataMap[dataType])
}

func nullableColumn(table, columnName string, nullable bool) string {
	if nullable {
		return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v DROP NOT NULL", table, columnName)
	}
	return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v SET NOT NULL", table, columnName)
}

func renameColumn(table, oldColumnName, newColumnName string) string {
	return fmt.Sprintf("ALTER TABLE %v RENAME COLUMN %v TO %v", table, oldColumnName, newColumnName)
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
