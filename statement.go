package goe

import "strings"

const (
	DML   int16 = 1 //DML as SELECT, INSERT, UPDATE and DELETE
	ATT   int16 = 2 //Attribute
	TABLE int16 = 3
	JOIN  int16 = 4
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
)

type statement struct {
	keyword string
	tip     int16
}

func createStatement(k string, t int16) *statement {
	return &statement{keyword: k, tip: t}
}

type queue struct {
	head *node
	size int
}

type node struct {
	value *statement
	next  *node
}

func createQueue() *queue {
	return &queue{}
}

func (q *queue) add(v *statement) {
	n := &node{value: v}

	if q.head == nil {
		q.head = n
		q.size++
		return
	}

	tail := getTail(q.head, n.value)
	if tail == nil {
		return
	}
	tail.next = n
	q.size++
}

func (q *queue) get() *statement {
	if q.head == nil {
		return nil
	}

	n := q.head
	q.head = q.head.next
	q.size--
	return n.value
}

func getTail(n *node, v *statement) *node {
	if n.value.keyword == v.keyword {
		return nil
	}
	if n.next != nil {
		return getTail(n.next, v)
	}
	return n
}

func buildSelect(sql *strings.Builder, q *queue) {
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
		break
	}
}
