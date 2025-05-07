package sqlbuilderutil

import (
	"fmt"
	"strings"

	"fknsrs.biz/p/sqlbuilder"
	"fknsrs.biz/p/reflectutil"

	"fknsrs.biz/p/ytmusic/internal/stringutil"
)

type Table struct {
	*sqlbuilder.Table
	nameMap map[string]string
}

func (t *Table) C(name string) *sqlbuilder.BasicColumn {
	if columnName, ok := t.nameMap[name]; ok {
		name = columnName
	}

	return t.Table.C(name)
}

func MakeTable(v interface{}) (*Table, error) {
	s, err := reflectutil.GetDescription(v)
	if err != nil {
		return nil, fmt.Errorf("sqlbuilderx.MakeTable: could not get struct description: %w", err)
	}

	var tableName string
	var columnNames []string

	nameMap := make(map[string]string)

	for _, f := range s.Fields().WithoutTagValue("sql", "-") {
		var columnName string

		sqlTag := f.Tag("sql")

		if sqlTag != nil && sqlTag.Value() != "" {
			columnName = sqlTag.Value()
		} else {
			columnName = stringutil.PascalToSnake(f.Name())
		}

		columnNames = append(columnNames, columnName)

		nameMap[f.Name()] = columnName
		nameMap[strings.ToLower(f.Name())] = columnName
		nameMap[columnName] = columnName

		if sqlTag != nil {
			if tableParameter := sqlTag.Parameter("table"); tableParameter != nil {
				tableName = tableParameter.Value()
			}
		}
	}

	if tableName == "" {
		tableName = stringutil.PascalToSnake(s.Name())
	}

	return &Table{
		Table:   sqlbuilder.NewTable(tableName, columnNames...),
		nameMap: nameMap,
	}, nil
}

func MustMakeTable(v interface{}) *Table {
	t, err := MakeTable(v)
	if err != nil {
		panic(err)
	}
	return t
}
