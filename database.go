package goe

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

// type Config struct {
// 	MigrationsPath string
// }

type conn struct {
	*sql.DB
	query string
}

type DB struct {
	conn   conn
	errors []error
	//Config
	//tables map[string]*table
}

// func (db *DB) Migrate(s any) {
// 	e := reflect.TypeOf(s).Elem()
// 	db.tables[e.Name()] = setTable(e, db)
// 	fmt.Println(db.tables)
// }

// func setTable(e reflect.Type, db *DB) *table {
// 	var t *table
// 	if db.tables[e.Name()] == nil {
// 		t = &table{}
// 	} else {
// 		t = db.tables[e.Name()]
// 	}
// 	t.name = e.Name()
// 	if t.attributes == nil {
// 		t.attributes = make(map[string]*attribute, e.NumField())
// 	}
// 	for i := 0; i < e.NumField(); i++ {
// 		switch e.Field(i).Type.Kind() {
// 		case reflect.Slice:
// 			if db.tables[e.Field(i).Type.Elem().Name()] == nil {
// 				db.tables[e.Field(i).Type.Elem().Name()] = &table{}
// 				setFk(db.tables[e.Field(i).Type.Elem().Name()], e)
// 			} else {
// 				manyToMany(t, e, e.Field(i).Type.Elem(), db)
// 				//Check if there  is a fk for the same table
// 			}
// 			//t.attributes[e.Field(i).Name] = setAttribute(e.Field(i), db, e.Field(i).Name)
// 			break
// 		}
// 		//t.attributes[e.Field(i).Name] = setAttribute(e.Field(i), db, e.Field(i).Name)
// 	}

// 	return t
// }

// // Set foreign key
// func setFk(t *table, e reflect.Type) {
// 	t.attributes = make(map[string]*attribute)
// 	t.attributes[e.Name()] = &attribute{name: "Id" + e.Name()}
// }

// // foreign key
// func manyToMany(t *table, t1 reflect.Type, t2 reflect.Type, db *DB) {
// 	if checkManyToMany(t1, t2) {
// 		db.tables[t1.Name()+t2.Name()] = &table{}
// 		t.attributes[t2.Name()] = nil
// 	}
// 	//set fk
// }

// func checkManyToMany(t1 reflect.Type, t2 reflect.Type) bool {
// 	for i := 0; i < t2.NumField(); i++ {
// 		if t2.Field(i).Type.Kind() == reflect.Slice {
// 			//fmt.Println(t2.Field(i).Type.Elem().Name(), t1.Name())
// 			if t2.Field(i).Type.Elem().Name() == t1.Name() {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func setAttribute(a reflect.StructField, db *DB, fn string) *attribute {
// 	var at *attribute
// 	switch a.Type.Kind() {
// 	case reflect.Slice:
// 		fmt.Println(fn)
// 		if db.tables[a.Type.Elem().Name()] != nil {
// 			fmt.Println("new table " + a.Type.Elem().Name())
// 		} else {
// 			at = &attribute{}
// 			//Check if the foreing key is always true
// 			at.null = false
// 			if db.tables[a.Type.Elem().Name()] == nil {
// 				db.tables[a.Type.Elem().Name()] = &table{}
// 			}
// 			db.tables[a.Type.Elem().Name()].attributes = make(map[string]*attribute, a.Type.Elem().NumField())
// 			db.tables[a.Type.Elem().Name()].attributes["Id"+a.Type.Elem().Name()] = at
// 			fmt.Println(fn)
// 		}
// 		break
// 	case reflect.String:
// 		break
// 	}
// 	// t := a.Tag.Get("goe")
// 	// fmt.Println(t)
// 	// fmt.Println(a.Type.Kind())
// 	return at
// }

// func (db *DB) Select(s any, a ...string) From {
// 	db.attr = a
// 	return db
// }

// func (db *DB) From(table string) Table {
// 	for _, v := range db.attr {
// 		fmt.Println(table + "." + v)
// 	}
// 	return db.tables[table]
// }

// func (db *DB) Select(args ...Attribute) Rows {
// 	for _, v := range args {
// 		elem := reflect.TypeOf(v)
// 		fmt.Println(elem.Name())
// 		// a := v.(*attribute)
// 		// fmt.Println(a.Table, a.Name)
// 	}

// 	return db
// }

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

func (db *DB) Select(args ...Att) Rows {
	//TODO Better Query
	for _, v := range args {
		switch v := v.(type) {
		case *att:
			q := fmt.Sprintf("SELECT %v FROM %v;", v.name, v.table)
			db.conn.query = q
		}
	}

	return db
}

func (db *DB) Where(b boolean) Rows {
	return db
}

func (db *DB) Result(target any) {
	db.errors = nil
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		if db.error(fmt.Errorf("%v: target result needs to be a pointer try &animals", pkg)) {
			return
		}
	}

	//Generate query if dont have
	//db.conn.query = "SELECT Animal.id, Animal.name, Animal.emoji FROM Animal;"
	db.handlerResult(value.Elem())
}

func (db *DB) handlerResult(value reflect.Value) {
	switch value.Kind() {
	case reflect.Slice:
		db.handlerQuery(value)
	case reflect.Struct:
		fmt.Println("struct")
	default:
		fmt.Println("default")
	}
}

func (db *DB) handlerQuery(value reflect.Value) {
	rows, err := db.conn.QueryContext(context.Background(), db.conn.query)
	if db.error(err) {
		return
	}
	defer rows.Close()

	//Prepare dest for query
	c, err := rows.Columns()
	db.error(err)

	dest := make([]any, len(c))
	for i := range dest {
		dest[i] = new(sql.RawBytes)
	}

	//Check the result target
	switch value.Type().Elem().Kind() {
	case reflect.Struct:
		err = mapStructQuery(rows, dest, value)
		db.error(err)
	default:
		err := mapQuery(rows, dest, value)
		db.error(err)
	}
}

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

func (db *DB) error(err error) bool {
	if err != nil {
		db.errors = append(db.errors, err)
		return true
	}
	return false
}

func (db *DB) Errors() []error {
	return db.errors
}
