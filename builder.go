package goe

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type builder struct {
	conn   conn
	sql    *strings.Builder
	queue  *queue
	tables *queue
}

func createBuilder() *builder {
	return &builder{sql: &strings.Builder{}, queue: createQueue(), tables: createQueue()}
}

func (b *builder) Result(target any) {

	//verifiy tables and joins

	//generate query
	b.buildSql()

	// db.errors = nil
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	fmt.Println(b.sql)
	//b.handlerResult(value.Elem())
}

func (b *builder) writeTables() {
	if b.tables.size > 1 {

	}
}

func (b *builder) handlerResult(value reflect.Value) {
	switch value.Kind() {
	case reflect.Slice:
		b.handlerQuery(value)
	case reflect.Struct:
		fmt.Println("struct")
	default:
		fmt.Println("default")
	}
}

func (b *builder) handlerQuery(value reflect.Value) {
	rows, err := b.conn.QueryContext(context.Background(), "SELECT Animal.id, Animal.name, Animal.emoji FROM Animal;")

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

func (b *builder) buildSql() {
	b.queue.buildSql(b.sql)
}
