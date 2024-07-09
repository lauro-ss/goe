package goe

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

func handlerValues(conn Connection, sqlQuery string, args []any) {
	defer conn.Close()

	_, err := conn.ExecContext(context.Background(), sqlQuery, args...)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) {
	defer conn.Close()

	row := conn.QueryRowContext(context.Background(), sqlQuery, args...)

	err := row.Scan(value.FieldByName(idName).Addr().Interface())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handlerResult(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string) {
	defer conn.Close()

	switch value.Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Struct {
			handlerStructQuery(conn, sqlQuery, value, args, structColumns)
			return
		}
		handlerQuery(conn, sqlQuery, value, args)
	case reflect.Struct:
		handlerStructQueryRow(conn, sqlQuery, value, args, structColumns)
	default:
		handlerQueryRow(conn, sqlQuery, value, args)
	}
}

func handlerQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any) {
	dest := make([]any, 1)
	for i := range dest {
		dest[i] = value.Addr().Interface()
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil {
		fmt.Println(err)
		return
	}
	value.Set(reflect.ValueOf(dest[0]).Elem())
}

func handlerStructQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, columns []string) {
	dest := make([]any, len(columns))
	for i := range dest {
		t, _ := value.Type().FieldByName(columns[i])
		dest[i] = reflect.New(t.Type).Interface()
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, a := range dest {
		value.FieldByName(columns[i]).Set(reflect.ValueOf(a).Elem())
	}
}

func handlerQuery(conn Connection, sqlQuery string, value reflect.Value, args []any) {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	dest := make([]any, 1)
	for i := range dest {
		dest[i] = reflect.New(value.Type().Elem()).Interface()
	}

	err = mapQuery(rows, dest, value)

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
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
			s.Set(reflect.ValueOf(a).Elem())
		}
		value.Set(reflect.Append(value, s))
	}
	return err
}

func handlerStructQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string) {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	dest := make([]any, len(structColumns))
	for i := range dest {
		if t, ok := value.Type().Elem().FieldByName(structColumns[i]); ok {
			dest[i] = reflect.New(t.Type).Interface()
		}
	}

	err = mapStructQuery(rows, dest, value, structColumns)

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
}

func mapStructQuery(rows *sql.Rows, dest []any, value reflect.Value, columns []string) (err error) {
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
			s.FieldByName(columns[i]).Set(reflect.ValueOf(a).Elem())
		}
		value.Set(reflect.Append(value, s))
	}
	return err
}
