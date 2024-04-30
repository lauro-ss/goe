package goe

import (
	"strings"
)

const (
	DML    int8 = 1 //DML as SELECT, INSERT, UPDATE and DELETE
	ATT    int8 = 2 //Attribute
	TABLE  int8 = 3
	JOIN   int8 = 4
	MIDDLE int8 = 6
)

func writeSelect(sql *strings.Builder, q *statementQueue) {
	if q.head != nil {
		writeStatement(sql, q.head)
		q.head = q.head.next
		q.size--
		writeSelect(sql, q)
		return
	}

	sql.WriteString(";")
}

func writeStatement(sql *strings.Builder, n *node) {
	switch n.value.tip {
	case ATT:
		// next node is a attribute
		if n.next != nil && n.next.value.tip == ATT {
			sql.WriteString(n.value.keyword)
			sql.WriteRune(',')
			return
		}
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
	case DML:
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
	case TABLE:
		sql.WriteString(n.value.keyword)
	case JOIN:
		sql.WriteRune('\n')
		sql.WriteString(n.value.keyword)
	case MIDDLE:
		sql.WriteRune('\n')
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
	default:
		sql.WriteString(n.value.keyword)
	}
}
