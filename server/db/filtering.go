package db

import (
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type (
	LinkOperatorType  string
	OperatorValueType string
)

const (
	LinkOperatorTypeAnd = "and"
	LinkOperatorTypeOr  = "or"

	OperatorValueTypeContains   = "contains"
	OperatorValueTypeStartsWith = "startsWith"
	OperatorValueTypeEndsWith   = "endsWith"
	OperatorValueTypeEquals     = "equals"

	OperatorValueTypeIsNull    = "isnull"
	OperatorValueTypeIsNotNull = "isnotnull"

	// Dates
	OperatorValueTypeIs           = "is"
	OperatorValueTypeIsNot        = "not"
	OperatorValueTypeIsAfter      = "after"
	OperatorValueTypeIsOnOrAfter  = "onOrAfter"
	OperatorValueTypeIsBefore     = "before"
	OperatorValueTypeIsOnOrBefore = "onOrBefore"

	// Numbers
	OperatorValueTypeNumberEquals    = "="
	OperatorValueTypeNumberNotEquals = "!="
	OperatorValueTypeGreaterThan     = ">"
	OperatorValueTypeGreaterOrEqual  = ">="
	OperatorValueTypeLessThan        = "<"
	OperatorValueTypeLessOrEqual     = "<="
)

// ListFilterRequest contains filter data commonly used in list requests
type ListFilterRequest struct {
	LinkOperator LinkOperatorType         `json:"link_operator"`
	Items        []*ListFilterRequestItem `json:"items"`
}

// ListFilterRequestItem contains instructions on filtering
type ListFilterRequestItem struct {
	Table    string            `json:"table"`
	Column   string            `json:"column"`
	Operator OperatorValueType `json:"operator"`
	Value    string            `json:"value"`
}

func GenerateListFilterQueryMod(table, column, value string, operator OperatorValueType, index int, linkOperator LinkOperatorType) qm.QueryMod {
	checkValue := value
	checkColumn := column
	if table != "" {
		checkColumn = fmt.Sprintf("%s.%s", table, column)
	}

	condition := fmt.Sprintf("%s %s ?", checkColumn, operator)

	switch operator {
	case OperatorValueTypeContains, OperatorValueTypeStartsWith, OperatorValueTypeEndsWith:
		switch operator {
		case OperatorValueTypeContains:
			checkValue = "%" + value + "%"
		case OperatorValueTypeStartsWith:
			checkValue = value + "%"
		case OperatorValueTypeEndsWith:
			checkValue = "%" + value
		}
		break
	}

	switch operator {
	case OperatorValueTypeIsNull:
		condition = fmt.Sprintf("%s IS NULL", checkColumn)
		break
	case OperatorValueTypeIsNotNull:
		condition = fmt.Sprintf("%s IS NOT NULL", checkColumn)
		break
	case OperatorValueTypeEquals, OperatorValueTypeIs, OperatorValueTypeNumberEquals:
		condition = fmt.Sprintf("%s = ?", checkColumn)
		break
	case OperatorValueTypeIsNot, OperatorValueTypeNumberNotEquals:
		condition = fmt.Sprintf("%s <> ?", checkColumn)
		break
	case OperatorValueTypeIsAfter, OperatorValueTypeGreaterThan:
		condition = fmt.Sprintf("%s > ?", checkColumn)
		break
	case OperatorValueTypeIsOnOrAfter, OperatorValueTypeGreaterOrEqual:
		condition = fmt.Sprintf("%s >= ?", checkColumn)
		break
	case OperatorValueTypeIsBefore, OperatorValueTypeLessThan:
		condition = fmt.Sprintf("%s < ?", checkColumn)
		break
	case OperatorValueTypeIsOnOrBefore, OperatorValueTypeLessOrEqual:
		condition = fmt.Sprintf("%s <= ?", checkColumn)
		break
	case OperatorValueTypeContains, OperatorValueTypeStartsWith, OperatorValueTypeEndsWith:
		condition = fmt.Sprintf("%s ILIKE ?", checkColumn)
	}

	if index == 0 {
		return qm.Where(condition, checkValue)
	}
	if linkOperator == LinkOperatorTypeOr {
		return qm.Or(condition, checkValue)
	}
	return qm.And(condition, checkValue)
}

// ColumnFilter generates SQL for filtering a column
func GenerateListFilterSQL(column string, value string, operator OperatorValueType, index int) (string, string) {
	checkValue := value
	condition := ""
	indexStr := fmt.Sprintf("$%d", index)
	column = fmt.Sprintf(`"%s"`, column)

	switch operator {
	case OperatorValueTypeContains, OperatorValueTypeStartsWith, OperatorValueTypeEndsWith:
		// Strings
		condition = fmt.Sprintf("%s ILIKE $%d", column, index)
		switch operator {
		case OperatorValueTypeContains:
			checkValue = "%" + value + "%"
		case OperatorValueTypeStartsWith:
			checkValue = value + "%"
		case OperatorValueTypeEndsWith:
			checkValue = "%" + value
		}

	case OperatorValueTypeIs, OperatorValueTypeIsNot, OperatorValueTypeIsAfter, OperatorValueTypeIsOnOrAfter, OperatorValueTypeIsBefore, OperatorValueTypeIsOnOrBefore:
		// Dates (convert column to date to compare by day)
		// column += "::date"
		if checkValue == "" {
			return "", checkValue // don't filter if no value is set
		}

	case OperatorValueTypeNumberEquals, OperatorValueTypeNumberNotEquals, OperatorValueTypeGreaterThan, OperatorValueTypeGreaterOrEqual, OperatorValueTypeLessThan, OperatorValueTypeLessOrEqual:
		// Numbers
		if checkValue == "" {
			checkValue = "0"
		}
	}

	switch operator {
	case OperatorValueTypeIsNull:
		condition = fmt.Sprintf("%s IS NULL", column)
	case OperatorValueTypeIsNotNull:
		condition = fmt.Sprintf("%s IS NOT NULL", column)
	case OperatorValueTypeEquals, OperatorValueTypeIs, OperatorValueTypeNumberEquals:
		condition = fmt.Sprintf("%s = %s", column, indexStr)
	case OperatorValueTypeIsNot, OperatorValueTypeNumberNotEquals:
		condition = fmt.Sprintf("%s <> %s", column, indexStr)
	case OperatorValueTypeIsAfter, OperatorValueTypeGreaterThan:
		condition = fmt.Sprintf("%s > %s", column, indexStr)
	case OperatorValueTypeIsOnOrAfter, OperatorValueTypeGreaterOrEqual:
		condition = fmt.Sprintf("%s >= %s", column, indexStr)
	case OperatorValueTypeIsBefore, OperatorValueTypeLessThan:
		condition = fmt.Sprintf("%s < %s", column, indexStr)
	case OperatorValueTypeIsOnOrBefore, OperatorValueTypeLessOrEqual:
		condition = fmt.Sprintf("%s <= %s", column, indexStr)
	}

	return condition, checkValue
}

// AttributeFilterRequest contains attribute-specific filter data commonly used in list requests
type AttributeFilterRequest struct {
	LinkOperator LinkOperatorType              `json:"linkOperator"`
	Items        []*AttributeFilterRequestItem `json:"items"`
}

// AttributeFilterRequestItem contains instructions on filtering
type AttributeFilterRequestItem struct {
	Trait         string            `json:"trait"`
	Value         string            `json:"value"`
	OperatorValue OperatorValueType `json:"operatorValue"`
}

// GenerateAttributeFilterSQL generates SQL for filtering a column
func GenerateAttributeFilterSQL(trait string, value string, operator OperatorValueType, index int, tableName string) (*string, error) {
	condition := fmt.Sprintf(`
	%[1]s.attributes @> '[{"trait_type": "%[2]s", "value": "%[3]s"}]' `, tableName, trait, value)
	return &condition, nil
}

// GenerateDataFilterSQL generates SQL for filtering a data column
func GenerateDataFilterSQL(trait string, value string, index int, tableName string) string {
	condition := fmt.Sprintf(`%[1]s."data"::text ILIKE '%%"%[2]s": "%[3]s%%'`, tableName, trait, value)
	return condition
}

// GenerateDataFilterSQL generates SQL for filtering a data column
func GenerateDataSearchSQL(trait string, search string, index int, tableName string) (string, string) {
	indexStr := fmt.Sprintf("$%d", index)
	condition := fmt.Sprintf(`(%[1]s."data"::json -> 'mech' -> '%[2]s')::text ILIKE %[3]s`, tableName, trait, indexStr)
	return search, condition
}

// GenerateDataFilterSQL generates SQL for filtering a data column
func GenerateDataSearchStoreItemsSQL(trait string, search string, index int, tableName string) (string, string) {
	indexStr := fmt.Sprintf("$%d", index)
	condition := fmt.Sprintf(`(%[1]s."data"::json -> 'template' -> '%[2]s')::text ILIKE %[3]s`, tableName, trait, indexStr)
	return search, condition
}
