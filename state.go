package goe

import (
	"fmt"
	"reflect"
)

type state struct {
	conn    conn
	builder *builder
}

func createState(conn conn, qt int8) *state {
	return &state{conn: conn, builder: createBuilder(qt)}
}

func (s *state) Where(brs ...*booleanResult) StateSelect {
	s.builder.brs = brs
	for _, br := range s.builder.brs {
		switch br.tip {
		case EQUALS:
			s.builder.tables.add(createStatement(br.pk.table, writeTABLE))
			s.builder.pks.add(br.pk)
		}
	}
	return s
}

func (s *state) Result(target any) {
	// db.errors = nil
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerResult(s.conn, s.builder.sql.String(), value.Elem())
}

func (s *state) querySelect(args []string, addrMap map[string]any) StateSelect {
	s.builder.args = args
	s.builder.buildSelect(addrMap)
	return s
}

func (s *state) Values(target any) {
	// db.errors = nil
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Struct {
		fmt.Printf("%v: target value needs to be a struct\n", pkg)
		return
	}

	s.builder.buildValues(value)

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	//handlerResult(s.conn, s.builder.sql.String(), value.Elem())
}

func (s *state) queryInsert(args []string, addrMap map[string]any) StateInsert {
	s.builder.args = args
	s.builder.buildInsert(addrMap)
	return s
}
