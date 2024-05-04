package goe

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

func handlerValues(conn conn, sqlQuery string, value reflect.Value, args []any, idName string) {
	row := conn.QueryRowContext(context.Background(), sqlQuery, args...)
	id := returnTarget(value, idName)
	err := row.Scan(id)
	if err != nil {
		fmt.Println(err)
	}

	value.FieldByName(idName).Set(reflect.ValueOf(id).Elem())
}

func returnTarget(value reflect.Value, idName string) any {
	switch value.FieldByName(idName).Kind() {
	case reflect.Int:
		return new(int)
	case reflect.String:
		return new(string)
	default:
		return new(any)
	}
}

func handlerResult(conn conn, sqlQuery string, value reflect.Value, args []any) {
	switch value.Kind() {
	case reflect.Slice:
		handlerQuery(conn, sqlQuery, value, args)
	case reflect.Struct:
		//handlerQueryRow(conn, sqlQuery, value, args)
		fmt.Println("struct")
	default:
		fmt.Println("default")
	}
}

func handlerQuery(conn conn, sqlQuery string, value reflect.Value, args []any) {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	//Prepare dest for query
	c, err := rows.Columns()

	if err != nil {
		fmt.Println(err)
		return
	}

	dest := make([]any, len(c))
	for i := range dest {
		dest[i] = new(sql.RawBytes)
	}

	//Check the result target
	switch value.Type().Elem().Kind() {
	case reflect.Struct:
		err = mapStructQuery(rows, dest, value)

		//TODO: Better error
		if err != nil {
			fmt.Println(err)
			return
		}
	default:
		err = mapQuery(rows, dest, value)

		//TODO: Better error
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func mapStructQuery(rows *sql.Rows, dest []any, value reflect.Value) (err error) {
	//TODO: add count for slices
	value.Set(reflect.MakeSlice(value.Type(), 0, 0))
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
		value.Set(reflect.Append(value, s))
	}
	return err
}

func mapQuery(rows *sql.Rows, dest []any, value reflect.Value) (err error) {
	//TODO: Len of slice be the size of the query
	value.Set(reflect.MakeSlice(value.Type(), 0, 0))

	for i := 0; rows.Next(); i++ {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		s := reflect.New(value.Type().Elem()).Elem()
		for _, a := range dest {
			setValue(s, a)
		}
		value.Set(reflect.Append(value, s))
	}
	return err
}

func setValue(v reflect.Value, a any) {
	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(string(*(a.(*sql.RawBytes))))
	}
}
