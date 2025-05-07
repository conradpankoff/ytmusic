package odata

import (
	"fmt"
	"strings"
)

func ParseFilter(filterString string) (*Filter, error) {
	var conditions []Condition

	for _, s := range splitFilter(filterString) {
		condition, err := parseCondition(s)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, *condition)
	}

	return &Filter{Conditions: conditions}, nil
}

// Split a filter string into individual conditions
func splitFilter(filterString string) []string {
	var splitFilter []string

	// Keep track of the number of open parentheses
	parens := 0

	// Split the filter string by 'and' and 'or', ignoring those inside parentheses
	for _, char := range filterString {
		switch char {
		case '(':
			parens++
		case ')':
			parens--
		case ' ':
			if parens == 0 {
				splitFilter = append(splitFilter, "")
				continue
			}
		}

		if len(splitFilter) == 0 {
			splitFilter = append(splitFilter, "")
		}
		splitFilter[len(splitFilter)-1] += string(char)
	}

	return splitFilter
}

// Parse a single condition
func parseCondition(conditionString string) (*Condition, error) {
	// Remove leading and trailing whitespace
	conditionString = strings.TrimSpace(conditionString)

	// Check for 'and' and 'or' operators
	var operator Operator
	if strings.HasPrefix(conditionString, "and ") {
		operator = AndOperator
		conditionString = strings.TrimSpace(conditionString[4:])
	} else if strings.HasPrefix(conditionString, "or ") {
		operator = OrOperator
		conditionString = strings.TrimSpace(conditionString[3:])
	}

	// Split the condition into property, operator, and value
	parts := strings.Split(conditionString, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid condition: %s", conditionString)
	}

	property := parts[0]
	operatorString := parts[1]
	value := parts[2]

	// Convert the operator string into an Operator enum
	operator, ok := operatorStrings[operatorString]
	if !ok {
		return nil, fmt.Errorf("invalid operator: %s", operatorString)
	}

	return &Condition{Property: property, Operator: operator, Value: value}, nil
}
