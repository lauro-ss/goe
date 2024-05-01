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

func (s *state) Where(brs ...*booleanResult) Rows {
	s.builder.brs = brs
	for _, br := range s.builder.brs {
		switch br.tip {
		case EQUALS:
			s.builder.tables.add(createStatement(br.pk.table, writeTABLE))
			flagPk := *br.pk
			flagPk.skipFlag = true
			s.builder.pks.add(&flagPk)
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

func (s *state) querySelect(args []string, addrMap map[string]any) Rows {
	s.builder.args = args
	s.builder.buildSelect(addrMap)
	return s
}
