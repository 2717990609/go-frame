// Package database 纯 SQL 查询构建（无 Gorm 依赖，供 SQLx 等使用）
package database

import (
	"fmt"
	"strings"
)

// BuildSelectSQL 将 Query 转换为 SELECT SQL 和参数
func BuildSelectSQL(table string, query Query) (sql string, args []any) {
	var parts []string
	if len(query.Select) > 0 {
		parts = append(parts, "SELECT "+strings.Join(query.Select, ", "))
	} else {
		parts = append(parts, "SELECT *")
	}
	parts = append(parts, "FROM "+table)

	if len(query.Where) > 0 {
		var whereParts []string
		for _, c := range query.Where {
			switch strings.ToLower(c.Operator) {
			case "=", "!=", ">", "<", ">=", "<=", "like":
				whereParts = append(whereParts, fmt.Sprintf("%s %s ?", c.Column, strings.ToUpper(c.Operator)))
				args = append(args, c.Value)
			case "in":
				ph, a := expandIn(c.Value)
				whereParts = append(whereParts, c.Column+" IN ("+ph+")")
				args = append(args, a...)
			case "not_in":
				ph, a := expandIn(c.Value)
				whereParts = append(whereParts, c.Column+" NOT IN ("+ph+")")
				args = append(args, a...)
			case "is_null":
				whereParts = append(whereParts, c.Column+" IS NULL")
			case "is_not_null":
				whereParts = append(whereParts, c.Column+" IS NOT NULL")
			default:
				whereParts = append(whereParts, fmt.Sprintf("%s %s ?", c.Column, c.Operator))
				args = append(args, c.Value)
			}
		}
		parts = append(parts, "WHERE "+strings.Join(whereParts, " AND "))
	}
	for _, j := range query.Joins {
		parts = append(parts, j)
	}
	if len(query.GroupBy) > 0 {
		parts = append(parts, "GROUP BY "+strings.Join(query.GroupBy, ", "))
	}
	if len(query.Having) > 0 {
		var havingParts []string
		for _, c := range query.Having {
			havingParts = append(havingParts, fmt.Sprintf("%s %s ?", c.Column, c.Operator))
			args = append(args, c.Value)
		}
		parts = append(parts, "HAVING "+strings.Join(havingParts, " AND "))
	}
	if len(query.Order) > 0 {
		orderParts := make([]string, len(query.Order))
		for i, o := range query.Order {
			orderParts[i] = fmt.Sprintf("%s %s", o.Column, strings.ToUpper(o.Direction))
		}
		parts = append(parts, "ORDER BY "+strings.Join(orderParts, ", "))
	}
	if query.Limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *query.Limit))
	}
	if query.Offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *query.Offset))
	}
	return strings.Join(parts, " "), args
}

func expandIn(v any) (placeholders string, args []any) {
	switch vals := v.(type) {
	case []any:
		// In("col", []int{1,2,3}) 传入时 Value 为 []interface{}{[]int{1,2,3}}
		if len(vals) == 1 {
			if nested := sliceToAny(vals[0]); len(nested) > 1 {
				return expandIn(nested)
			}
		}
		ps := make([]string, len(vals))
		for i := range vals {
			ps[i] = "?"
			args = append(args, vals[i])
		}
		return strings.Join(ps, ","), args
	case []int, []int64, []string:
		return expandIn(sliceToAny(v))
	default:
		return "?", append(args, v)
	}
}

func sliceToAny(v any) []any {
	switch s := v.(type) {
	case []int:
		out := make([]any, len(s))
		for i, x := range s {
			out[i] = x
		}
		return out
	case []int64:
		out := make([]any, len(s))
		for i, x := range s {
			out[i] = x
		}
		return out
	case []string:
		out := make([]any, len(s))
		for i, x := range s {
			out[i] = x
		}
		return out
	}
	return []any{v}
}
