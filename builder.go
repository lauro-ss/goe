package goe

import (
	"fmt"
	"strings"
)

const (
	querySELECT int8 = 1
	queryINSERT int8 = 2
	queryUPDATE int8 = 3
)

type builder struct {
	sql       *strings.Builder
	args      []string
	brs       []*booleanResult
	queue     *statementQueue
	tables    *statementQueue
	pks       *pkQueue
	queryType int8
}

func createBuilder(qt int8) *builder {
	return &builder{
		sql:       &strings.Builder{},
		queue:     createStatementQueue(),
		tables:    createStatementQueue(),
		queryType: qt,
		pks:       createPkQueue()}
}

func (b *builder) buildSelect(addrMap map[string]any) {
	//TODO: Set a drive type to share stm
	b.queue.add(&SELECT)

	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.queue.add(createStatement(atr.name, ATT))
			b.tables.add(createStatement(atr.pk.table, TABLE))

			//TODO: Add a list pk?
			b.pks.add(atr.pk)
		case *pk:
			b.queue.add(createStatement(atr.name, ATT))
			b.tables.add(createStatement(atr.table, TABLE))

			//TODO: Add a list pk?
			b.pks.add(atr)
		}
	}

	b.queue.add(&FROM)
}

func (b *builder) buildSql() {
	switch b.queryType {
	case querySELECT:
		b.writeTables()
		b.writeWhere()
		buildSelect(b.sql, b.queue)
	case queryINSERT:
		break
	case queryUPDATE:
		break
	}
}

func (b *builder) writeWhere() {
	if len(b.brs) == 0 {
		return
	}
	b.queue.add(&WHERE)
	for _, br := range b.brs {
		switch br.tip {
		case EQUALS:
			b.queue.add(createStatement(fmt.Sprintf("%v = %v", br.arg, br.value), 0))
		}
	}
}

func (b *builder) writeTables() {
	b.queue.add(b.tables.get())
	if b.tables.size >= 1 {
		for table := b.tables.get(); table != nil; {
			writeJoins(table, b.pks, b.queue)
			table = b.tables.get()
		}
	}
}

func writeJoins(table *statement, pks *pkQueue, stQueue *statementQueue) {
	for pk := pks.get(); pk != nil; {
		if pk.table != table.keyword {
			switch fk := pk.fks[table.keyword].(type) {
			case *manyToOne:
				if fk.hasMany {
					stQueue.add(
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pk.name, fk.id), JOIN),
					)
				} else {
					stQueue.add(
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pks.findPk(table.keyword).name, fk.id), JOIN),
					)
				}
			case *manyToMany:
				if !pk.skipFlag {
					stQueue.add(
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", fk.table, pk.name, fk.ids[pk.table]), JOIN),
					)
					stQueue.add(
						createStatement(
							fmt.Sprintf(
								"inner join %v on (%v = %v)",
								table.keyword, fk.ids[table.keyword],
								pks.findPk(table.keyword).name), JOIN,
						),
					)
				}
			}
		}
		pk = pks.get()
	}
}
