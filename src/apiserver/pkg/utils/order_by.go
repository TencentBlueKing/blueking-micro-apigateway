package utils

import (
	"strings"

	"gorm.io/gen/field"
)

// ParseOrderByExprList parses comma-separated field:direction pairs into order expressions.
func ParseOrderByExprList(
	ascFieldMap map[string]field.Expr,
	descFieldMap map[string]field.Expr,
	orderBy string,
) []field.Expr {
	var orderByExprs []field.Expr

	sortConditions := strings.SplitSeq(orderBy, ",")
	for condition := range sortConditions {
		parts := strings.Split(condition, ":")
		if len(parts) != 2 {
			continue
		}

		fieldName := parts[0]
		direction := strings.ToLower(parts[1])

		switch direction {
		case "asc":
			if _, ok := ascFieldMap[fieldName]; ok {
				orderByExprs = append(orderByExprs, ascFieldMap[fieldName])
			}
		case "desc":
			if _, ok := descFieldMap[fieldName]; ok {
				orderByExprs = append(orderByExprs, descFieldMap[fieldName])
			}
		}
	}

	return orderByExprs
}
