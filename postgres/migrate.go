package postgres

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/lauro-ss/goe"
)

type migrateAttribute struct {
	attribute  any
	migrated   bool
	index      string
	indexNames []string
}

type migrateTable struct {
	pk       *goe.MigratePk
	atts     map[string]*migrateAttribute
	migrated bool
}

func (db *Driver) Migrate(migrator *goe.Migrator, conn goe.Connection) {
	defer conn.Close()

	tables := make(map[string]*migrateTable, 0)
	tablesManyToMany := make(map[string]*goe.MigrateManyToMany, 0)

	dataMap := map[string]string{
		"string":    "text",
		"int16":     "smallint",
		"int32":     "integer",
		"int64":     "bigint",
		"float32":   "real",
		"float64":   "double precision",
		"[]uint8":   "bytea",
		"time.Time": "timestamp",
		"bool":      "boolean",
	}

	for _, v := range migrator.Tables {
		switch atr := v.(type) {
		case *goe.MigratePk:
			if tables[atr.Table] == nil {
				tables[atr.Table] = newMigrateTable(migrator.Tables, atr.Table)
			}
			for _, fkAny := range atr.Fks {
				switch fk := fkAny.(type) {
				case *goe.MigrateManyToMany:
					if tablesManyToMany[fk.Table] == nil {
						tablesManyToMany[fk.Table] = fk
					}
				}
			}
		}
	}

	sql := new(strings.Builder)
	sqlColumns := new(strings.Builder)
	for _, t := range tables {
		checkTableChanges(t, tables, dataMap, sqlColumns, conn)
		// check for new index
		checkIndex(t.atts, t.pk.Table, sqlColumns, conn)
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

	for _, t := range tablesManyToMany {
		if table := createManyToManyTable(t, dataMap, conn); table != nil {
			createTableSql(table.tableName, table.createPks, table.ids, sql)
		}
	}

	dropTables(tables, tablesManyToMany, sql, conn)

	sql.WriteString(sqlColumns.String())

	if sql.Len() != 0 {
		if _, err := conn.ExecContext(context.Background(), sql.String()); err != nil {
			fmt.Println(err)
		}
	}
}

func dropTables(tables map[string]*migrateTable, tablesManyToMany map[string]*goe.MigrateManyToMany, sql *strings.Builder, conn goe.Connection) {
	databaseTables := make([]string, 0)
	sqlQuery := `SELECT ic.table_name FROM information_schema.columns ic where ic.table_schema = 'public' group by ic.table_name;`
	rows, err := conn.QueryContext(context.Background(), sqlQuery)
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
			var c string
			fmt.Printf(`goe:do you want to remove table "%v" from database? (y/n):`, table)
			fmt.Scanln(&c)
			if c == "y" {
				sql.WriteString(fmt.Sprintf("DROP TABLE %v;", table))
			}
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

func createManyToManyTable(mtm *goe.MigrateManyToMany, dataMap map[string]string, conn goe.Connection) *tableManytoMany {
	sql := `SELECT
	table_name
	FROM information_schema.columns WHERE table_name = $1;`

	rows, err := conn.QueryContext(context.Background(), sql, mtm.Table)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		return nil
	}
	return newMigrateTableManyToMany(mtm, dataMap)
}

func newMigrateTableManyToMany(fk *goe.MigrateManyToMany, dataMap map[string]string) *tableManytoMany {
	table := new(tableManytoMany)
	table.tableName = fmt.Sprintf("CREATE TABLE %v (", fk.Table)
	table.ids = make([]string, 0, len(fk.Ids))
	table.createPks = "primary key (id_flag, id_flag)"
	for key, attr := range fk.Ids {
		attr.DataType = checkDataType(attr.DataType, dataMap)
		table.ids = append(table.ids, fmt.Sprintf("%v %v NOT NULL REFERENCES %v,", attr.AttributeName, attr.DataType, key))
		table.createPks = strings.Replace(table.createPks, "id_flag", attr.AttributeName, 1)
	}
	return table
}

func newMigrateTable(tables []any, tableName string) *migrateTable {
	table := new(migrateTable)
	table.atts = make(map[string]*migrateAttribute)
	for _, v := range tables {
		switch atr := v.(type) {
		case *goe.MigratePk:
			if atr.Table == tableName {
				table.pk = atr
				ma := new(migrateAttribute)
				ma.attribute = atr
				table.atts[atr.AttributeName] = ma
				for _, fkAny := range atr.Fks {
					switch fk := fkAny.(type) {
					case *goe.MigrateManyToOne:
						if !fk.HasMany {
							ma := new(migrateAttribute)
							ma.attribute = fk
							table.atts[strings.Split(fk.Id, ".")[1]] = ma
						}
					}
				}
			}
		case *goe.MigrateAtt:
			if atr.Pk.Table == tableName {
				ma := new(migrateAttribute)
				ma.attribute = atr
				ma.index = atr.Index
				table.atts[atr.AttributeName] = ma
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

func checkTableChanges(mt *migrateTable, tables map[string]*migrateTable, dataMap map[string]string, sql *strings.Builder, conn goe.Connection) {
	sqlTableInfos := `SELECT
	column_name, CASE 
	WHEN data_type = 'character varying' 
	THEN CONCAT('varchar','(',character_maximum_length,')')
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

	rows, err := conn.QueryContext(context.Background(), sqlTableInfos, mt.pk.Table)
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
		checkNewFields(mt, dataMap, tables, sql)
	}
}

func createTable(mt *migrateTable, tables map[string]*migrateTable, dataMap map[string]string, tablesCreate []table) []table {
	t := table{}
	mt.migrated = true
	t.name = fmt.Sprintf("CREATE TABLE %v (", mt.pk.Table)
	for _, attrAny := range mt.atts {
		switch attr := attrAny.attribute.(type) {
		case *goe.MigratePk:
			attr.DataType = checkDataType(attr.DataType, dataMap)
			if attr.AutoIncrement {
				t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v NOT NULL,", attr.AttributeName, checkTypeAutoIncrement(attr.DataType)))
			} else {
				t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v NOT NULL,", attr.AttributeName, attr.DataType))
			}
			t.createPk = fmt.Sprintf("primary key (%v)", attr.AttributeName)
		case *goe.MigrateAtt:
			attr.DataType = checkDataType(attr.DataType, dataMap)
			t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v %v,", attr.AttributeName, attr.DataType, func(n bool) string {
				if n {
					return "NULL"
				} else {
					return "NOT NULL"
				}
			}(attr.Nullable)))
		case *goe.MigrateManyToOne:
			tableFk := tables[attr.TargetTable]
			if tableFk == nil {
				fmt.Printf("goe: table '%v' not mapped\n", attr.TargetTable)
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

func foreingManyToOne(attr *goe.MigrateManyToOne, pk *goe.MigratePk, dataMap map[string]string) string {
	pk.DataType = checkDataType(pk.DataType, dataMap)
	return fmt.Sprintf("%v %v %v REFERENCES %v,", strings.Split(attr.Id, ".")[1], pk.DataType, func(n bool) string {
		if n {
			return "NULL"
		} else {
			return "NOT NULL"
		}
	}(attr.Nullable), pk.Table)
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

func checkIndex(attrs map[string]*migrateAttribute, table string, sql *strings.Builder, conn goe.Connection) {
	sqlQuery := `SELECT DISTINCT ci.relname, i.indisunique as is_unique, c.relname FROM pg_index i
	JOIN pg_attribute a ON i.indexrelid = a.attrelid
	JOIN pg_class ci ON ci.oid = i.indexrelid
	JOIN pg_class c ON c.oid = i.indrelid
	where i.indisprimary = false AND c.relname = $1;
	`

	rows, err := conn.QueryContext(context.Background(), sqlQuery, table)
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
			case *goe.MigrateAtt:
				if attr.Index != "" {
					indexs := strings.Split(attr.Index, ",")
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
			case *goe.MigrateAtt:
				indexs := strings.Split(attr.Index, ",")
				for _, index := range indexs {
					n := getIndexValue(index, "n:")
					if n == "" {
						fmt.Printf(`error: index "%v" without declared name on struct "%v" attribute "%v". to fix add tag "n:"%v`, index, attr.Pk.Table, attr.AttributeName, "\n")
						continue
					}
					n = fmt.Sprintf("%v_%v", attr.Pk.Table, n)
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
					if c := checkColumnIndex(n, attr.AttributeName, attr.Pk.Table, attrs, unique); c != "" {
						sql.WriteString(createIndexColumns(attr.Pk.Table, attr.AttributeName, c, n, unique, f))
						continue
					}
					sql.WriteString(createIndex(attr.Pk.Table, n, attr.AttributeName, unique, f))
				}
			}
		}
	}
}

func checkColumnIndex(indexName, attrName, table string, attrs map[string]*migrateAttribute, unique bool) string {
	for _, v := range attrs {
		if v.index != "" {
			switch attr := v.attribute.(type) {
			case *goe.MigrateAtt:
				indexs := strings.Split(attr.Index, ",")
				for _, i := range indexs {
					if attr.AttributeName != attrName && strings.Contains(i, "unique") == unique && indexName == fmt.Sprintf("%v_%v", table, getIndexValue(i, "n:")) {
						if f := getIndexValue(i, "f:"); f != "" {
							f := fmt.Sprintf("%v(%v)", f, attr.AttributeName)
							return f
						}
						return attr.AttributeName
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
		fmt.Printf(`goe:field "%v" exists on database table but is missed on struct "%v"%v`, databaseTable.columnName, mt.pk.Table, "\n")
		fmt.Printf(`goe:do you renamed the field "%v" on "%v"? (leave empty if not):`, databaseTable.columnName, mt.pk.Table)
		var c string
		fmt.Scanln(&c)
		if c == "" {
			fmt.Printf(`goe:do you want to remove the field "%v" on table "%v"? (y/n):`, databaseTable.columnName, mt.pk.Table)
			fmt.Scanln(&c)
			if c == "y" {
				sql.WriteString(dropColumn(mt.pk.Table, databaseTable.columnName))
			}
			return
		}
		if mt.atts[c] != nil {
			sql.WriteString(renameColumn(mt.pk.Table, databaseTable.columnName, mt.atts[c].attribute.(*goe.MigrateAtt).AttributeName))
			mt.atts[c].migrated = true
		}
		return
	}
	switch attr := attrAny.attribute.(type) {
	case *goe.MigratePk:
		dataType := checkDataType(attr.DataType, dataMap)
		if attr.AutoIncrement {
			dataType = checkTypeAutoIncrement(dataType)
		}
		if databaseTable.dataType != dataType {
			sql.WriteString(alterColumn(attr.Table, databaseTable.columnName, dataType, dataMap))
		}
		attrAny.migrated = true
		tables[attr.Table].migrated = true
	case *goe.MigrateAtt:
		dataType := checkDataType(attr.DataType, dataMap)
		if databaseTable.dataType != dataType {
			sql.WriteString(alterColumn(attr.Pk.Table, databaseTable.columnName, dataType, dataMap))
		}
		if databaseTable.nullable != attr.Nullable {
			sql.WriteString(nullableColumn(attr.Pk.Table, attr.AttributeName, attr.Nullable))
		}
		attrAny.migrated = true
		tables[attr.Pk.Table].migrated = true
	case *goe.MigrateManyToOne:
		if databaseTable.nullable != attr.Nullable {
			sql.WriteString(nullableColumn(mt.pk.Table, databaseTable.columnName, attr.Nullable))
		}
		attrAny.migrated = true
	}
}

func checkNewFields(mt *migrateTable, dataMap map[string]string, tables map[string]*migrateTable, sql *strings.Builder) {
	for _, v := range mt.atts {
		if !v.migrated {
			switch attr := v.attribute.(type) {
			case *goe.MigrateAtt:
				sql.WriteString(addColumn(mt.pk.Table, attr.AttributeName, checkDataType(attr.DataType, dataMap), attr.Nullable))
			case *goe.MigrateManyToOne:
				targetTable := tables[attr.TargetTable]
				if targetTable == nil {
					fmt.Printf(`goe:target struct "%v" it's not mapped on Database struct%v`, attr.TargetTable, "\n")
					continue
				}
				table, column, _ := strings.Cut(attr.Id, ".")
				sql.WriteString(addColumn(table, column, checkDataType(targetTable.pk.DataType, dataMap), attr.Nullable))
				sql.WriteString(addFkColumn(table, column, attr.TargetTable))
			}
		}
	}
}

func addColumn(table, column, dataType string, nullable bool) string {
	return fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v %v;", table, column, dataType,
		func(n bool) string {
			if n {
				return "NULL"
			}
			return "NOT NULL"
		}(nullable))
}

func addFkColumn(table, column, targetTable string) string {
	return fmt.Sprintf("ALTER TABLE %v ADD CONSTRAINT %v FOREIGN KEY (%v) REFERENCES %v;", table, fmt.Sprintf("fk_%v_%v", targetTable, column), column, targetTable)
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
		"smallint": "smallserial",
		"integer":  "serial",
		"bigint":   "bigserial",
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
	return fmt.Sprintf("ALTER TABLE %v RENAME COLUMN %v TO %v;", table, oldColumnName, newColumnName)
}

func dropColumn(table, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %v DROP COLUMN %v;", table, columnName)
}
