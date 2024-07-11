package goe

import (
	"context"
	"database/sql"
	"reflect"
)

func handlerValues(conn Connection, sqlQuery string, args []any) error {
	defer conn.Close()

	_, err := conn.ExecContext(context.Background(), sqlQuery, args...)
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) error {
	defer conn.Close()

	row := conn.QueryRowContext(context.Background(), sqlQuery, args...)

	err := row.Scan(value.FieldByName(idName).Addr().Interface())
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturningBatch(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) error {
	defer conn.Close()

	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)
	if err != nil {
		return err
	}

	i := 0
	for rows.Next() {
		err = rows.Scan(value.Index(i).FieldByName(idName).Addr().Interface())
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func handlerResult(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string) error {
	defer conn.Close()

	switch value.Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Struct {
			return handlerStructQuery(conn, sqlQuery, value, args, structColumns)
		}
		return handlerQuery(conn, sqlQuery, value, args)
	case reflect.Struct:
		return handlerStructQueryRow(conn, sqlQuery, value, args, structColumns)
	default:
		return handlerQueryRow(conn, sqlQuery, value, args)
	}
}

func handlerQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any) error {
	dest := make([]any, 1)
	for i := range dest {
		dest[i] = value.Addr().Interface()
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(dest[0]).Elem())
	return nil
}

func handlerStructQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, columns []string) error {
	dest := make([]any, len(columns))
	for i := range dest {
		t, _ := value.Type().FieldByName(columns[i])
		dest[i] = reflect.New(t.Type).Interface()
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil {
		return err
	}

	for i, a := range dest {
		value.FieldByName(columns[i]).Set(reflect.ValueOf(a).Elem())
	}
	return nil
}

func handlerQuery(conn Connection, sqlQuery string, value reflect.Value, args []any) error {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	if err != nil {
		return err
	}
	defer rows.Close()

	dest := make([]any, 1)
	for i := range dest {
		dest[i] = reflect.New(value.Type().Elem()).Interface()
	}

	err = mapQuery(rows, dest, value)

	if err != nil {
		return err
	}
	return nil
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

func handlerStructQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string) error {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	if err != nil {
		return err
	}
	defer rows.Close()

	dest := make([]any, len(structColumns))
	for i := range dest {
		if t, ok := value.Type().Elem().FieldByName(structColumns[i]); ok {
			dest[i] = reflect.New(t.Type).Interface()
		}
	}

	err = mapStructQuery(rows, dest, value, structColumns)

	if err != nil {
		return err
	}
	return nil
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
