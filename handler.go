package goe

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
)

var ErrNotFound = errors.New("goe: not found any element on result set")

func handlerValues(conn Connection, sqlQuery string, args []any, ctx context.Context) error {
	_, err := conn.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string, ctx context.Context) error {
	row := conn.QueryRowContext(ctx, sqlQuery, args...)

	err := row.Scan(value.FieldByName(idName).Addr().Interface())
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturningBatch(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string, ctx context.Context) error {
	rows, err := conn.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

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

func handlerResult(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string, limit uint, ctx context.Context) error {
	switch value.Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Struct && value.Type().Elem().Name() != "Time" {
			return handlerStructQuery(conn, sqlQuery, value, args, structColumns, limit, ctx)
		}
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return handlerQueryRow(conn, sqlQuery, value, args, ctx)
		}
		return handlerQuery(conn, sqlQuery, value, args, limit, ctx)
	case reflect.Struct:
		if value.Type().Name() != "Time" {
			return handlerStructQueryRow(conn, sqlQuery, value, args, structColumns, ctx)
		}
		return handlerQueryRow(conn, sqlQuery, value, args, ctx)
	default:
		return handlerQueryRow(conn, sqlQuery, value, args, ctx)
	}
}

func handlerQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, ctx context.Context) error {
	dest := make([]any, 1)
	for i := range dest {
		dest[i] = value.Addr().Interface()
	}
	err := conn.QueryRowContext(ctx, sqlQuery, args...).Scan(dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	value.Set(reflect.ValueOf(dest[0]).Elem())
	return nil
}

func handlerStructQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, columns []string, ctx context.Context) error {
	dest := make([]any, len(columns))
	for i := range dest {
		t, _ := value.Type().FieldByName(columns[i])
		if t.Type == nil {
			dest[i] = reflect.New(value.Type().Field(i).Type).Interface()
			continue
		}
		dest[i] = reflect.New(t.Type).Interface()
	}
	err := conn.QueryRowContext(ctx, sqlQuery, args...).Scan(dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	var field reflect.Value
	for i, a := range dest {
		field = value.FieldByName(columns[i])
		if !field.CanSet() {
			field = value.Field(i)
		}
		field.Set(reflect.ValueOf(a).Elem())
	}
	return nil
}

func handlerQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, limit uint, ctx context.Context) error {
	rows, err := conn.QueryContext(ctx, sqlQuery, args...)

	if err != nil {
		return err
	}
	defer rows.Close()

	dest := make([]any, 1)
	for i := range dest {
		dest[i] = reflect.New(value.Type().Elem()).Interface()
	}

	err = mapQuery(rows, dest, value, limit)

	if err != nil {
		return err
	}
	return nil
}

func mapQuery(rows *sql.Rows, dest []any, value reflect.Value, limit uint) (err error) {
	if !rows.Next() {
		return ErrNotFound
	}

	value.Set(reflect.MakeSlice(value.Type(), 0, int(limit)))

	err = rows.Scan(dest...)
	if err != nil {
		return err
	}
	value.Set(reflect.Append(value, reflect.ValueOf(dest[0]).Elem()))

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		value.Set(reflect.Append(value, reflect.ValueOf(dest[0]).Elem()))
	}
	return nil
}

func handlerStructQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string, limit uint, ctx context.Context) error {
	rows, err := conn.QueryContext(ctx, sqlQuery, args...)

	if err != nil {
		return err
	}
	defer rows.Close()

	dest := make([]any, len(structColumns))
	for i := range dest {
		if t, ok := value.Type().Elem().FieldByName(structColumns[i]); ok {
			dest[i] = reflect.New(t.Type).Interface()
			continue
		}
		dest[i] = reflect.New(value.Type().Elem().Field(i).Type).Interface()
	}

	err = mapStructQuery(rows, dest, value, structColumns, limit)

	if err != nil {
		return err
	}
	return nil
}

func mapStructQuery(rows *sql.Rows, dest []any, value reflect.Value, columns []string, limit uint) (err error) {
	if !rows.Next() {
		return ErrNotFound
	}

	value.Set(reflect.MakeSlice(value.Type(), 0, int(limit)))

	err = handlerStructRow(rows, dest, value, columns)
	if err != nil {
		return err
	}

	for rows.Next() {
		err = handlerStructRow(rows, dest, value, columns)
		if err != nil {
			return err
		}
	}
	return nil
}

func handlerStructRow(rows *sql.Rows, dest []any, value reflect.Value, columns []string) (err error) {
	err = rows.Scan(dest...)
	if err != nil {
		return err
	}
	s := reflect.New(value.Type().Elem()).Elem()
	var f reflect.Value
	//Fills the target
	for i, a := range dest {
		f = s.FieldByName(columns[i])
		if !f.CanSet() {
			f = s.Field(i)
		}
		f.Set(reflect.ValueOf(a).Elem())
	}
	value.Set(reflect.Append(value, s))
	return nil
}
