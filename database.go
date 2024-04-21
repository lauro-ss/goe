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
	conn conn
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

	builder := createBuilder()
	builder.conn = db.conn

	//TODO: Set a drive type to share stm
	builder.queue.add(&SELECT)

	//TODO Better Query
	for _, v := range args {
		switch v := v.(type) {
		case *att:
			builder.queue.add(createStatement(v.name, ATT))
			builder.tables.add(createStatement(v.pk.table, TABLE))
		case *pk:
			builder.queue.add(createStatement(v.name, ATT))
			builder.tables.add(createStatement(v.table, TABLE))
		default:
			fmt.Println("Call a method to check struct")
		}
	}

	return builder
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
