package models

import (
	"errors"
	"fmt"
	u "p3/utils"
	"regexp"
	"strings"
	"time"
)

func ComplexFilterToMap(complexFilter string) (map[string]any, error) {
	// Split the input string into individual filter expressions
	chars := []string{"(", ")", "&", "|"}
	for _, char := range chars {
		complexFilter = strings.ReplaceAll(complexFilter, char, " "+char+" ")
	}
	return complexExpressionToMap(strings.Fields(complexFilter))
}

func complexExpressionToMap(expressions []string) (map[string]any, error) {
	// Find the rightmost operator (AND, OR) outside of parentheses
	parenCount := 0
	for i := len(expressions) - 1; i >= 0; i-- {
		switch expressions[i] {
		case "(":
			parenCount++
		case ")":
			parenCount--
		case "&":
			if parenCount == 0 {
				first, _ := complexExpressionToMap(expressions[:i])
				second, _ := complexExpressionToMap(expressions[i+1:])
				return map[string]any{"$and": []map[string]any{
					first,
					second,
				}}, nil
			}
		case "|":
			if parenCount == 0 {
				first, _ := complexExpressionToMap(expressions[:i])
				second, _ := complexExpressionToMap(expressions[i+1:])
				return map[string]any{"$or": []map[string]any{
					first,
					second,
				}}, nil
			}
		}
	}

	// If there are no operators outside of parentheses, look for the innermost pair of parentheses
	for i := 0; i < len(expressions); i++ {
		if expressions[i] == "(" {
			start, end := i+1, i+1
			for parenCount := 1; end < len(expressions) && parenCount > 0; end++ {
				switch expressions[end] {
				case "(":
					parenCount++
				case ")":
					parenCount--
				}
			}
			return complexExpressionToMap(append(expressions[:start-1], expressions[start:end-1]...))
		}
	}

	// Base case: single filter expression
	return singleExpressionToMap(expressions)
}

func singleExpressionToMap(expressions []string) (map[string]any, error) {
	re := regexp.MustCompile(`^([\w-.]+)\s*(<=|>=|<|>|!=|=)\s*((\[)*[\w-,.*]+(\])*)$`)
	ops := map[string]string{"<=": "$lte", ">=": "$gte", "<": "$lt", ">": "$gt", "!=": "$not"}

	if len(expressions) <= 3 {
		expression := strings.Join(expressions[:], "")
		if match := re.FindStringSubmatch(expression); match != nil {
			// convert filter value to proper type
			filterName := match[1]                   // e.g. category
			filterOp := match[2]                     // e.g. !=
			filterValue := u.ConvertString(match[3]) // e.g. device
			switch filterName {
			case "startDate":
				return map[string]any{"lastUpdated": map[string]any{"$gte": filterValue}}, nil
			case "endDate":
				return map[string]any{"lastUpdated": map[string]any{"$lte": filterValue}}, nil
			case "id", "name", "category", "description", "domain", "createdDate", "lastUpdated", "slug":
				if filterOp == "=" {
					return map[string]any{filterName: filterValue}, nil
				}
				return map[string]any{filterName: map[string]any{ops[filterOp]: filterValue}}, nil
			default:
				if filterOp == "=" {
					return map[string]any{"attributes." + filterName: filterValue}, nil
				}
				return map[string]any{"attributes." + filterName: map[string]any{ops[filterOp]: filterValue}}, nil
			}
		}
	}

	fmt.Println("Error: Invalid filter expression")
	return nil, errors.New("invalid filter expression")
}

func getDatesFromComplexFilters(req map[string]any) error {
	for k, v := range req {
		if k == "$and" || k == "$or" {
			for _, complexFilter := range v.([]map[string]any) {
				err := getDatesFromComplexFilters(complexFilter)
				if err != nil {
					return err
				}
			}
		} else if k == "lastUpdated" {
			for op, date := range v.(map[string]any) {
				parsedDate, err := time.Parse("2006-01-02", date.(string))
				if err != nil {
					return err
				}
				if op == "$lte" {
					parsedDate = parsedDate.Add(time.Hour * 24)
				}
				req[k] = map[string]any{op: parsedDate}
			}
		}
	}
	return nil
}
