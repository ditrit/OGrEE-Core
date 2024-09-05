package utils

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

const NAME_CHARACTER_REGEX = `(\w|\-)*`                              // accepted characters to compose ids
const NAME_REGEX = `\w` + NAME_CHARACTER_REGEX                       // accepted regex for names that compose ids
const NAME_RECURSIVE_REGEX = `(\.` + NAME_REGEX + `)*\.`             // accepted regex for names that compose ids in recursive way
const NAME_RECURSIVE_REGEX_WITH_DEPTH = `(\.` + NAME_REGEX + `)$1\.` // accepted regex for names that compose ids in recursive way with min and max depth

var doubleStarWithDepthRegex = regexp.MustCompile(`\\\.\*\*({\d+,\d*})\\\.`) // ** with min and max depth
var doubleStarRegex = regexp.MustCompile(`\\\.\*\*\\\.`)                     // ** without depth
var pointStar = ".*"
var starRegex = regexp.MustCompile(`([^\)]|^)\*+`) // * with something different to ")" before to avoid replacing the * written in the previous steps

func ApplyWildcardsOnComplexFilter(filter map[string]interface{}) {
	for key, val := range filter {
		switch v := val.(type) {
		case string:
			if key == "$not" || (!strings.HasPrefix(key, "$") && strings.Contains(v, "*")) {
				// only for '=' with * and always for '!='
				filter[key] = regexToMongoFilter(applyWildcards(v))
			}
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					ApplyWildcardsOnComplexFilter(m)
				}
			}
		case []map[string]any:
			for _, item := range v {
				ApplyWildcardsOnComplexFilter(item)
			}
		case map[string]interface{}:
			ApplyWildcardsOnComplexFilter(v)
		}
	}
}

func applyWildcards(value string) string {
	value = strings.ReplaceAll(value, ".", `\.`)

	value = doubleStarWithDepthRegex.ReplaceAllString(value, NAME_RECURSIVE_REGEX_WITH_DEPTH)
	value = doubleStarRegex.ReplaceAllString(value, NAME_RECURSIVE_REGEX)
	value = strings.ReplaceAll(value, pointStar, "."+NAME_REGEX)         // .* must start with a \w
	value = starRegex.ReplaceAllString(value, "$1"+NAME_CHARACTER_REGEX) // any other * doesn't need to start with \w

	return value
}

func regexToMongoFilter(regex string) bson.M {
	return bson.M{"$regex": "^" + regex + "$"}
}
