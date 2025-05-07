package godatautil

import (
	"fmt"
	"strings"

	sb "fknsrs.biz/p/sqlbuilder"
	"github.com/gost/godata"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
)

var (
	ErrFieldNotFound = fmt.Errorf("field not found")
)

func MakeCondition(q *godata.GoDataQuery, table *sqlbuilderutil.Table) (sb.AsExpr, error) {
	if q == nil || q.Filter == nil {
		return nil, nil
	}

	expr, err := makeCondition(q.Filter.Tree, table)
	if err != nil {
		return nil, fmt.Errorf("godatautil.MakeCondition: %w", err)
	}

	return expr, nil
}

func makeCondition(n *godata.ParseNode, table *sqlbuilderutil.Table) (sb.AsExpr, error) {
	switch n.Token.Type {
	case godata.FilterTokenLogical:
		var a []sb.AsExpr
		for _, e := range n.Children {
			expr, err := makeCondition(e, table)
			if err != nil {
				return nil, fmt.Errorf("godatautil.makeCondition: %w", err)
			}
			a = append(a, expr)
		}
		switch n.Token.Value {
		case "and", "or":
			return sb.BooleanOperator(n.Token.Value, a...), nil
		default:
			return nil, fmt.Errorf("godatautil.makeCondition: unrecognised logical filter type %q", n.Token.Value)
		}
	case godata.FilterTokenFunc:
		switch n.Token.Value {
		case "substringof":
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("godatautil.makeCondition: substringof must have exactly two arguments; instead had %d", len(n.Children))
			}
			if tokenType := n.Children[0].Token.Type; tokenType != godata.FilterTokenLiteral {
				return nil, fmt.Errorf("godatautil.makeCondition: substringof first argument must be Literal; was instead %s", filterTokenName(tokenType))
			}
			if tokenType := n.Children[1].Token.Type; tokenType != godata.FilterTokenString {
				return nil, fmt.Errorf("godatautil.makeCondition: substringof first argument must be String; was instead %s", filterTokenName(tokenType))
			}

			c := table.C(n.Children[0].Token.Value)
			if c == nil {
				return nil, fmt.Errorf("godatautil.makeCondition: unrecognised field %s", n.Children[0].Token.Value)
			}

			return sb.Ne(
				sb.Func(
					"instr",
					c,
					sb.Bind(unquote(n.Children[1].Token.Value)),
				),
				sb.Literal("0"),
			), nil
		default:
			return nil, fmt.Errorf("godatautil.makeCondition: unrecognised function %s", n.Token.Value)
		}
	default:
		return nil, fmt.Errorf("godatautil.makeCondition: unrecognised token type %d (%s)", n.Token.Type, filterTokenName(n.Token.Type))
	}
}

func filterTokenName(tokenType int) string {
	switch tokenType {
	case godata.FilterTokenOpenParen: // 0
		return "OpenParen"
	case godata.FilterTokenCloseParen: // 1
		return "CloseParen"
	case godata.FilterTokenWhitespace: // 2
		return "Whitespace"
	case godata.FilterTokenNav: // 3
		return "Nav"
	case godata.FilterTokenColon: // 4
		return "Colon"
	case godata.FilterTokenComma: // 5
		return "Comma"
	case godata.FilterTokenLogical: // 6
		return "Logical"
	case godata.FilterTokenOp: // 7
		return "Op"
	case godata.FilterTokenFunc: // 8
		return "Func"
	case godata.FilterTokenLambda: // 9
		return "Lambda"
	case godata.FilterTokenNull: // 10
		return "Null"
	case godata.FilterTokenIt: // 11
		return "It"
	case godata.FilterTokenRoot: // 12
		return "Root"
	case godata.FilterTokenFloat: // 13
		return "Float"
	case godata.FilterTokenInteger: // 14
		return "Integer"
	case godata.FilterTokenString: // 15
		return "String"
	case godata.FilterTokenDate: // 16
		return "Date"
	case godata.FilterTokenTime: // 17
		return "Time"
	case godata.FilterTokenDateTime: // 18
		return "DateTime"
	case godata.FilterTokenBoolean: // 19
		return "Boolean"
	case godata.FilterTokenLiteral: // 20
		return "Literal"
	case godata.FilterTokenGeography: // 21
		return "Geography"
	default:
		return "???" // ??
	}
}

func MakeOrders(q *godata.GoDataQuery, table *sqlbuilderutil.Table, defaultOrders ...sb.AsOrderingTerm) ([]sb.AsOrderingTerm, error) {
	if q == nil || q.OrderBy == nil {
		return defaultOrders, nil
	}

	var a []sb.AsOrderingTerm

	for _, item := range q.OrderBy.OrderByItems {
		c := table.C(item.Field.Value)
		if c == nil {
			return nil, fmt.Errorf("godatautil.MakeOrders: could not find field %q: %w", item.Field.Value, ErrFieldNotFound)
		}

		switch item.Order {
		case "asc":
			a = append(a, sb.OrderAsc(c))
		case "desc":
			a = append(a, sb.OrderDesc(c))
		}
	}

	return a, nil
}

func MakeOffsetLimit(q *godata.GoDataQuery, defaultSkip, defaultTop int) *sb.OffsetLimitClause {
	skip := defaultSkip
	if q == nil || q.Skip != nil {
		skip = int(*q.Skip)
	}

	top := defaultTop
	if q == nil || q.Top != nil {
		top = int(*q.Top)
	}

	return sb.OffsetLimit(sb.Bind(skip), sb.Bind(top))
}

func unquote(s string) string {
	return strings.Replace(s[1:len(s)-1], "''", "'", -1)
}
