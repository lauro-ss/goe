package goe

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
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

type migrateAttribute struct {
	attribute  any
	migrated   bool
	index      string
	indexNames []string
}

type migrateTable struct {
	pk       *pk
	atts     map[string]*migrateAttribute
	migrated bool
}

func (db *DB) Migrate() {
	tables := make(map[string]*migrateTable, 0)
	tablesManyToMany := make(map[string]*manyToMany, 0)

	dataMap := map[string]string{
		"string":  "text",
		"int16":   "smallint",
		"int32":   "integer",
		"int64":   "bigint",
		"float32": "real",
		"float64": "double precision",
	}

	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if tables[atr.table] == nil {
				tables[atr.table] = newMigrateTable(db, atr.table)
			}
			for _, fkAny := range atr.fks {
				switch fk := fkAny.(type) {
				case *manyToMany:
					if tablesManyToMany[fk.table] == nil {
						tablesManyToMany[fk.table] = fk
					}
				}
			}
		}
	}

	sql := new(strings.Builder)
	for _, t := range tables {
		checkTableChanges(db, t, tables, dataMap, sql)
		// check for new index
		checkIndex(db, t.atts, t.pk.table, sql)
	}

	var tablesCreate []table
	// create new tables
	for _, t := range tables {
		if !t.migrated {
			tablesCreate = createTable(t, tables, dataMap, tablesCreate)
		}
	}

	for _, t := range tablesCreate {
		createTableSql(t.name, t.createPk, t.createAttrs, sql)
	}

	//TODO: Add sql builder
	for _, t := range tablesManyToMany {
		if table := createManyToManyTable(db, t, dataMap); table != nil {
			createTableSql(table.tableName, table.createPks, table.ids, sql)
		}
	}

	dropTables(db, tables, tablesManyToMany)

	if sql.Len() != 0 {
		if _, err := db.conn.Exec(sql.String()); err != nil {
			fmt.Println(err)
		}
	}
}

func dropTables(db *DB, tables map[string]*migrateTable, tablesManyToMany map[string]*manyToMany) {
	databaseTables := make([]string, 0)
	sql := `SELECT ic.table_name FROM information_schema.columns ic where ic.table_schema = 'public' group by ic.table_name;`
	rows, err := db.conn.Query(sql)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var table string

	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			fmt.Println(err)
			return
		}
		databaseTables = append(databaseTables, table)
	}

	for _, table = range databaseTables {
		ok := false
		for key := range tables {
			if table == key {
				ok = true
				break
			}
		}
		for key := range tablesManyToMany {
			if table == key {
				ok = true
				break
			}
		}
		if !ok {
			fmt.Println("drop table", table)
		}
	}

}

type tableManytoMany struct {
	tableName string
	ids       []string
	createPks string
}

func createTableSql(create, pks string, attributes []string, sql *strings.Builder) {
	sql.WriteString(create)
	for _, a := range attributes {
		sql.WriteString(a)
	}
	sql.WriteString(pks)
	sql.WriteString(");")
}

func createManyToManyTable(db *DB, mtm *manyToMany, dataMap map[string]string) *tableManytoMany {
	sql := `SELECT
	table_name
	FROM information_schema.columns WHERE table_name = $1;`

	rows, err := db.conn.Query(sql, mtm.table)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		return nil
	}
	//TODO: add sql builder
	return newMigrateTableManyToMany(mtm, dataMap)
}

func newMigrateTableManyToMany(fk *manyToMany, dataMap map[string]string) *tableManytoMany {
	table := new(tableManytoMany)
	table.tableName = fmt.Sprintf("CREATE TABLE %v (", fk.table)
	table.ids = make([]string, 0, len(fk.ids))
	table.createPks = "primary key (id_flag, id_flag)"
	for _, attr := range fk.ids {
		attr.dataType = checkDataType(attr.dataType, dataMap)
		table.ids = append(table.ids, fmt.Sprintf("%v %v NOT NULL,", attr.attributeName, attr.dataType))
		table.createPks = strings.Replace(table.createPks, "id_flag", attr.attributeName, 1)
	}
	return table
}

func newMigrateTable(db *DB, tableName string) *migrateTable {
	table := new(migrateTable)
	table.atts = make(map[string]*migrateAttribute)
	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if atr.table == tableName {
				table.pk = atr
				ma := new(migrateAttribute)
				ma.attribute = atr
				table.atts[atr.attributeName] = ma
				for _, fkAny := range atr.fks {
					switch fk := fkAny.(type) {
					case *manyToOne:
						if !fk.hasMany {
							ma := new(migrateAttribute)
							ma.attribute = fk
							table.atts[strings.Split(fk.id, ".")[1]] = ma
						}
					}
				}
			}
		case *att:
			if atr.pk.table == tableName {
				ma := new(migrateAttribute)
				ma.attribute = atr
				ma.index = atr.index
				table.atts[atr.attributeName] = ma
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

func checkTableChanges(db *DB, mt *migrateTable, tables map[string]*migrateTable, dataMap map[string]string, sql *strings.Builder) {
	sqlTableInfos := `SELECT
	column_name, CASE 
	WHEN data_type = 'character varying' 
	THEN CONCAT('varchar','(',character_maximum_length,')')
	WHEN data_type = 'boolean' THEN 'bool'
	when data_type = 'integer' then case WHEN column_default like 'nextval%' THEN 'serial' ELSE data_type end
	when data_type = 'bigint' then case WHEN column_default like 'nextval%' THEN 'bigserial' ELSE data_type end
	ELSE data_type END, 
	CASE WHEN column_default like 'nextval%' THEN True ELSE False end as auto_increment,
	CASE
	WHEN is_nullable = 'YES'
	THEN True
	ELSE False END AS is_nullable
	FROM information_schema.columns WHERE table_name = $1;	
	`

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
			//TODO: add error return
			fmt.Println(err)
			return
		}
		dts = append(dts, dt)
	}
	if len(dts) > 0 {
		//check fileds
		for i := range dts {
			checkFields(dts[i], mt, tables, dataMap, sql)
		}
		checkNewFields(mt, dataMap, sql)
	}
}

func createTable(mt *migrateTable, tables map[string]*migrateTable, dataMap map[string]string, tablesCreate []table) []table {
	t := table{}
	mt.migrated = true
	t.name = fmt.Sprintf("CREATE TABLE %v (", mt.pk.table)
	for _, attrAny := range mt.atts {
		switch attr := attrAny.attribute.(type) {
		case *pk:
			attr.dataType = checkDataType(attr.dataType, dataMap)
			if attr.autoIncrement {
				t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v NOT NULL,", attr.attributeName, checkTypeAutoIncrement(attr.dataType)))
			} else {
				t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v NOT NULL,", attr.attributeName, attr.dataType))
			}
			t.createPk = fmt.Sprintf("primary key (%v)", attr.attributeName)
		case *att:
			attr.dataType = checkDataType(attr.dataType, dataMap)
			t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v %v,", attr.attributeName, attr.dataType, func(n bool) string {
				if n {
					return "NULL"
				} else {
					return "NOT NULL"
				}
			}(attr.nullable)))
		case *manyToOne:
			tableFk := tables[attr.targetTable]
			if tableFk == nil {
				fmt.Printf("goe: table '%v' not mapped\n", attr.targetTable)
				return nil
			}
			if tableFk.migrated {
				t.createAttrs = append(t.createAttrs, foreingManyToOne(attr, tableFk.pk, dataMap))
			} else {
				tablesCreate = append(tablesCreate, createTable(tableFk, tables, dataMap, tablesCreate)...)
				t.createAttrs = append(t.createAttrs, foreingManyToOne(attr, tableFk.pk, dataMap))
			}
		}
	}
	tablesCreate = append(tablesCreate, t)
	return tablesCreate
}

func foreingManyToOne(attr *manyToOne, pk *pk, dataMap map[string]string) string {
	pk.dataType = checkDataType(pk.dataType, dataMap)
	return fmt.Sprintf("%v %v %v REFERENCES %v,", strings.Split(attr.id, ".")[1], pk.dataType, func(n bool) string {
		if n {
			return "NULL"
		} else {
			return "NOT NULL"
		}
	}(attr.nullable), pk.table)
}

type table struct {
	name        string
	createPk    string
	createAttrs []string
}

type databaseIndex struct {
	indexName string
	unique    bool
	table     string
}

func checkIndex(db *DB, attrs map[string]*migrateAttribute, table string, sql *strings.Builder) {
	sqlQuery := `SELECT DISTINCT ci.relname, i.indisunique as is_unique, c.relname FROM pg_index i
	JOIN pg_attribute a ON i.indexrelid = a.attrelid
	JOIN pg_class ci ON ci.oid = i.indexrelid
	JOIN pg_class c ON c.oid = i.indrelid
	where i.indisprimary = false AND c.relname = $1;
	`

	rows, err := db.conn.Query(sqlQuery, table)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	dis := make([]databaseIndex, 0)
	di := databaseIndex{}
	for rows.Next() {
		err = rows.Scan(&di.indexName, &di.unique, &di.table)
		if err != nil {
			fmt.Println(err)
			return
		}
		dis = append(dis, di)
	}

	migrateIndexs := make([]string, 0)
	for _, di = range dis {
		for _, v := range attrs {
			switch attr := v.attribute.(type) {
			case *att:
				if attr.index != "" {
					indexs := strings.Split(attr.index, ",")
					for _, index := range indexs {
						n := fmt.Sprintf("%v_%v", di.table, getIndexValue(index, "n:"))
						if di.indexName == n {
							migrateIndexs = append(migrateIndexs, di.indexName)
							//drop index if uniquenes changes
							if di.unique != strings.Contains(index, "unique") {
								sql.WriteString(fmt.Sprintf("DROP INDEX %v;", di.indexName))
								continue
							}
							v.indexNames = append(v.indexNames, di.indexName)
						}
					}
				}
			}
		}
	}

	//drop no match index
	for _, di = range dis {
		if !slices.Contains(migrateIndexs, di.indexName) {
			sql.WriteString(fmt.Sprintf("DROP INDEX %v;", di.indexName))
		}
	}

	checkNewIndexs(attrs, sql)
}

func checkNewIndexs(attrs map[string]*migrateAttribute, sql *strings.Builder) {
	created := make([]string, 0, len(attrs))
	for _, v := range attrs {
		if v.index != "" {
			switch attr := v.attribute.(type) {
			case *att:
				indexs := strings.Split(attr.index, ",")
				for _, index := range indexs {
					n := getIndexValue(index, "n:")
					if n == "" {
						fmt.Printf(`error: index "%v" without declared name on struct "%v" attribute "%v". to fix add tag "n:"%v`, index, attr.pk.table, attr.attributeName, "\n")
						continue
					}
					n = fmt.Sprintf("%v_%v", attr.pk.table, n)
					// chefk if index already exists on database
					if slices.Contains(v.indexNames, n) {
						continue
					}
					// skips created indexs
					if slices.Contains(created, n) {
						continue
					}
					unique := strings.Contains(index, "unique")
					created = append(created, n)
					f := getIndexValue(index, "f:")
					if c := checkColumnIndex(n, attr.attributeName, attr.pk.table, attrs, unique); c != "" {
						sql.WriteString(createIndexColumns(attr.pk.table, attr.attributeName, c, n, unique, f))
						continue
					}
					sql.WriteString(createIndex(attr.pk.table, n, attr.attributeName, unique, f))
				}
			}
		}
	}
}

func checkColumnIndex(indexName, attrName, table string, attrs map[string]*migrateAttribute, unique bool) string {
	for _, v := range attrs {
		if v.index != "" {
			switch attr := v.attribute.(type) {
			case *att:
				indexs := strings.Split(attr.index, ",")
				for _, i := range indexs {
					if attr.attributeName != attrName && strings.Contains(i, "unique") == unique && indexName == fmt.Sprintf("%v_%v", table, getIndexValue(i, "n:")) {
						if f := getIndexValue(i, "f:"); f != "" {
							f := fmt.Sprintf("%v(%v)", f, attr.attributeName)
							return f
						}
						return attr.attributeName
					}
				}
			}
		}
	}
	return ""
}

func getIndexValue(valueTag string, tag string) string {
	values := strings.Split(valueTag, " ")
	for _, v := range values {
		if _, value, ok := strings.Cut(v, tag); ok {
			return value
		}
	}
	return ""
}

func createIndex(table, name, attribute string, unique bool, function string) string {
	return fmt.Sprintf("CREATE %v %v ON %v (%v);",
		func(u bool) string {
			if unique {
				return "UNIQUE INDEX"
			} else {
				return "INDEX"
			}
		}(unique),
		name,
		table,
		func(a, f string) string {
			if f != "" {
				return fmt.Sprintf("%v(%v)", f, a)
			}
			return a
		}(attribute, function),
	)
}

func createIndexColumns(table, attribute1, attribute2, name string, unique bool, function string) string {
	return fmt.Sprintf("CREATE %v %v ON %v (%v,%v);", func(u bool) string {
		if unique {
			return "UNIQUE INDEX"
		} else {
			return "INDEX"
		}
	}(unique), name, table, func(a, f string) string {
		if f != "" {
			return fmt.Sprintf("%v(%v)", f, a)
		}
		return a
	}(attribute1, function), attribute2)
}

func checkFields(databaseTable databaseTable, mt *migrateTable, tables map[string]*migrateTable, dataMap map[string]string, sql *strings.Builder) {
	attrAny := mt.atts[databaseTable.columnName]
	if attrAny == nil {
		//TODO: Add prompt to drop or rename column
		//fmt.Printf("goe:field '%v'exists on database table but is missed on struct %v\n", databaseTable.columnName, mt.pk.table)
		//prompt a question to drop or set as a rename field
		//attrAny = mt.atts[databaseTable.columnName]
		return
	}
	switch attr := attrAny.attribute.(type) {
	case *pk:
		attr.dataType = checkDataType(attr.dataType, dataMap)
		if attr.autoIncrement {
			attr.dataType = checkTypeAutoIncrement(attr.dataType)
		}
		if databaseTable.dataType != attr.dataType {
			//TODO: change dataType of fks
			sql.WriteString(alterColumn(attr.table, databaseTable.columnName, attr.dataType, dataMap))
		}
		attrAny.migrated = true
		tables[attr.table].migrated = true
	case *att:
		attr.dataType = checkDataType(attr.dataType, dataMap)
		if databaseTable.dataType != attr.dataType {
			sql.WriteString(alterColumn(attr.pk.table, databaseTable.columnName, attr.dataType, dataMap))
		}
		if databaseTable.nullable != attr.nullable {
			sql.WriteString(nullableColumn(attr.pk.table, attr.attributeName, attr.nullable))
		}
		attrAny.migrated = true
		tables[attr.pk.table].migrated = true
	}
}

func checkNewFields(mt *migrateTable, dataMap map[string]string, sql *strings.Builder) {
	for _, v := range mt.atts {
		if !v.migrated {
			switch attr := v.attribute.(type) {
			case *att:
				attr.dataType = checkDataType(attr.dataType, dataMap)
				sql.WriteString(addColumn(mt.pk.table, attr.attributeName, attr.dataType))
			}
		}
	}
}

func addColumn(table, column, dataType string) string {
	return fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v;", table, column, dataType)
}

func checkDataType(structDataType string, dataMap map[string]string) string {
	if structDataType == "int8" || structDataType == "uint8" || structDataType == "uint16" {
		structDataType = "int16"
	} else if structDataType == "int" || structDataType == "uint" || structDataType == "uint32" {
		structDataType = "int32"
	} else if structDataType == "uint64" {
		structDataType = "int64"
	}
	if dataMap[structDataType] != "" {
		structDataType = dataMap[structDataType]
	}
	return structDataType
}

func checkTypeAutoIncrement(structDataType string) string {
	dataMap := map[string]string{
		"integer": "serial",
		"bigint":  "bigserial",
	}
	return dataMap[structDataType]
}

func alterColumn(table, column, dataType string, dataMap map[string]string) string {
	if dataMap[dataType] == "" {
		return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v;", table, column, dataType)
	}
	return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v;", table, column, dataMap[dataType])
}

func nullableColumn(table, columnName string, nullable bool) string {
	if nullable {
		return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v DROP NOT NULL;", table, columnName)
	}
	return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v SET NOT NULL;", table, columnName)
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

func (db *DB) DeleteIn(table1 any, table2 any) DeleteIn {
	stringArgs := getArgs(table1, table2)

	state := createDeleteInState(db.conn, queryUPDATE)

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
