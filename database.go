package goe

import (
	"fmt"
	"reflect"
)

type Config struct {
	MigrationsPath string
}

type database struct {
	Config
	tables map[string]*table
}

func (db *database) Migrate(s any) {
	e := reflect.TypeOf(s).Elem()
	db.tables[e.Name()] = setTable(e, db)
	fmt.Println(db.tables)
}

func setTable(e reflect.Type, db *database) *table {
	var t *table
	if db.tables[e.Name()] == nil {
		t = &table{}
	} else {
		t = db.tables[e.Name()]
	}
	t.name = e.Name()
	if t.attributes == nil {
		t.attributes = make(map[string]*attribute, e.NumField())
	}
	for i := 0; i < e.NumField(); i++ {
		switch e.Field(i).Type.Kind() {
		case reflect.Slice:
			if db.tables[e.Field(i).Type.Elem().Name()] == nil {
				db.tables[e.Field(i).Type.Elem().Name()] = &table{}
				setFk(db.tables[e.Field(i).Type.Elem().Name()], e)
			} else {
				manyToMany(t, e, e.Field(i).Type.Elem(), db)
				//Check if there  is a fk for the same table
			}
			//t.attributes[e.Field(i).Name] = setAttribute(e.Field(i), db, e.Field(i).Name)
			break
		}
		//t.attributes[e.Field(i).Name] = setAttribute(e.Field(i), db, e.Field(i).Name)
	}

	return t
}

// Set foreign key
func setFk(t *table, e reflect.Type) {
	t.attributes = make(map[string]*attribute)
	t.attributes[e.Name()] = &attribute{name: "Id" + e.Name()}
}

// foreign key
func manyToMany(t *table, t1 reflect.Type, t2 reflect.Type, db *database) {
	if checkManyToMany(t1, t2) {
		db.tables[t1.Name()+t2.Name()] = &table{}
		t.attributes[t2.Name()] = nil
	}
	//set fk
}

func checkManyToMany(t1 reflect.Type, t2 reflect.Type) bool {
	for i := 0; i < t2.NumField(); i++ {
		if t2.Field(i).Type.Kind() == reflect.Slice {
			//fmt.Println(t2.Field(i).Type.Elem().Name(), t1.Name())
			if t2.Field(i).Type.Elem().Name() == t1.Name() {
				return true
			}
		}
	}
	return false
}

// func setAttribute(a reflect.StructField, db *database, fn string) *attribute {
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

// func (db *database) Select(s any, a ...string) From {
// 	db.attr = a
// 	return db
// }

// func (db *database) From(table string) Table {
// 	for _, v := range db.attr {
// 		fmt.Println(table + "." + v)
// 	}
// 	return db.tables[table]
// }
