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

var (
	SELECT = statement{
		keyword: "SELECT",
		tip:     DML,
	}
	FROM = statement{
		keyword: "FROM",
		tip:     DML,
	}
	WHERE = statement{
		keyword: "WHERE",
		tip:     MIDDLE,
	}
)

type statement struct {
	keyword string
	tip     int8
}

func createStatement(k string, t int8) *statement {
	return &statement{keyword: k, tip: t}
}

func buildSelect(sql *strings.Builder, q *statementQueue) {
	if q.head != nil {
		writeSelect(sql, q.head)
		q.head = q.head.next
		q.size--
		buildSelect(sql, q)
		return
	}

	sql.WriteString(";")
}

func writeSelect(sql *strings.Builder, n *node) {
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
