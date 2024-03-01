package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	JSON_WHITESPACE string = " "
	JSON_SYNTAX     string = "{},:[]"
	JSON_NEWLINE    string = "\n"
)

func isValidSyntax(value string, nextValue string) bool {
	var syntaxCombination bytes.Buffer
	syntaxCombination.WriteString(value)
	syntaxCombination.WriteString(nextValue)
	invalidCombinations := []string{
		"{:", "}:", ":}", ",:", ":,",
		"{,", ",}", "[:", ":]", "]:",
		",]", "{]", "]{", "{{",
	}
	return !slices.Contains(invalidCombinations, syntaxCombination.String())
}

func parser(json []byte) (map[string]interface{}, error) {
	if len(json) == 0 {
		return map[string]interface{}{}, fmt.Errorf("Invalid Json")
	}

	listOfTokens := make([]interface{}, 0)
	listOfTokens, err := lexer(json, listOfTokens)
	if err != nil {
		return map[string]interface{}{}, err
	}
	firstValue, ok := listOfTokens[0].(string)
	if !ok || firstValue != "{" {
		return map[string]interface{}{}, fmt.Errorf("Invalid Json")
	}
	lastValue, ok := listOfTokens[len(listOfTokens)-1].(string)
	if !ok || lastValue != "}" {
		return map[string]interface{}{}, fmt.Errorf("Invalid Json")
	}

	result := make(map[string]interface{}, 100)
	parsedObject, _, err := parseObject(listOfTokens, result)

	return parsedObject, err
}

func parseObject(
	listOfTokens []interface{},
	result map[string]interface{},
) (map[string]interface{}, int, error) {

	var lastKeyInt int // keep track of number of elements iterated over

	// starting from 1 to avoid errors when retrieving n-1 element in the slice.
	for keyInt := 1; keyInt < len(listOfTokens)-1; keyInt++ {
		currentValue, okC := listOfTokens[keyInt].(string)
		previousValue, _ := listOfTokens[keyInt-1].(string)
		nextValue, okN := listOfTokens[keyInt+1].(string)

		// check for syntax
		if okC && okN {
			if !isValidSyntax(currentValue, nextValue) {
				return map[string]interface{}{},
					0,
					fmt.Errorf("Invalid Syntax %q followed by %q", currentValue, nextValue)
			}
		}

		// append if no nested structures
		if currentValue == ":" && nextValue != "{" && nextValue != "[" {
			result[previousValue] = listOfTokens[keyInt+1]
		} else if currentValue == ":" && nextValue == "{" { // unpack nested objects
			newResult := make(map[string]interface{}, 100)
			newValue, newKeyInt, err := parseObject(listOfTokens[keyInt+1:], newResult)
			if err != nil {
				return newValue, 0, err
			}
			result[previousValue] = newValue
			keyInt += newKeyInt + 2
		} else if currentValue == ":" && nextValue == "[" { // unpack arrays
			sliceResult := make([]interface{}, 0)
			newValue, newKeyInt, err := parseArray(listOfTokens[keyInt+1:], sliceResult)
			if err != nil {
				return map[string]interface{}{}, 0, err
			}
			result[previousValue] = newValue
			keyInt += newKeyInt + 2
		} else if currentValue == "}" { // end iteration
			lastKeyInt = keyInt
			break
		}
	}

	return result, lastKeyInt, nil
}

func parseArray(
	listOfTokens []interface{},
	result []interface{},
) ([]interface{}, int, error) {

	var lastKeyInt int

	// starting from 1 to avoid infinite loops and append directly from the first element
	for keyInt := 1; keyInt < len(listOfTokens)-1; keyInt++ {
		currentValue, okC := listOfTokens[keyInt].(string)
		nextValue, okN := listOfTokens[keyInt+1].(string)

		// check for syntax
		if okC && okN {
			if !isValidSyntax(currentValue, nextValue) {
				return []interface{}{},
					0,
					fmt.Errorf("Invalid Syntax %q followed by %q", currentValue, nextValue)
			}
		}

		// append if no nested structures and no json syntax
		if !okC || !strings.Contains(JSON_SYNTAX, currentValue) {
			result = append(result, listOfTokens[keyInt])
		} else if currentValue == "[" { // unpack nested arrays
			sliceResult := make([]interface{}, 0)
			newValue, newKeyInt, err := parseArray(listOfTokens[keyInt:], sliceResult)
			if err != nil {
				return []interface{}{}, 0, err
			}
			result = append(result, newValue)
			keyInt += newKeyInt + 1
		} else if currentValue == "{" { // unpack objects
			newResult := make(map[string]interface{}, 100)
			newValue, newKeyInt, err := parseObject(listOfTokens[keyInt+1:], newResult)
			if err != nil {
				return []interface{}{}, 0, err
			}
			result = append(result, newValue)
			keyInt += newKeyInt + 2
		} else if currentValue == "]" { // end iteration
			lastKeyInt = keyInt
			break
		}
	}

	return result, lastKeyInt, nil
}

func lexer(json []byte, tokens []interface{}) ([]interface{}, error) {
	for i := 0; i < len(json); i++ {
		parsedString, len := lexString(json)
		if len > 0 {
			tokens = append(tokens, parsedString)
			json = json[(len + 2):]
			continue
		}
		parsedInt, len := lexInt(json)
		if len > 0 {
			tokens = append(tokens, parsedInt)
			json = json[len:]
			continue
		}
		parsedFloat, len := lexFloat(json)
		if len > 0 {
			tokens = append(tokens, parsedFloat)
			json = json[len:]
			continue
		}
		parsedBool, len := lexBool(json)
		if len > 0 {
			tokens = append(tokens, parsedBool)
			json = json[len:]
			continue
		}
		_, len = lexNil(json)
		if len > 0 {
			tokens = append(tokens, nil)
			json = json[len:]
			continue
		}

		firstChar := string(json[0])
		switch {
		case strings.Contains(JSON_WHITESPACE, firstChar),
			strings.Contains(JSON_NEWLINE, firstChar):
			return lexer(json[1:], tokens)
		case strings.Contains(JSON_SYNTAX, firstChar):
			tokens = append(tokens, firstChar)
			return lexer(json[1:], tokens)
		default:
			return []interface{}{}, fmt.Errorf("unexpected character: %q", firstChar)
		}
	}

	return tokens, nil
}

func lexString(json []byte) (string, int) {
	var buffer bytes.Buffer
	if json[0] == '"' {
		json = json[1:]
	} else {
		return "", 0
	}
	for _, char := range json {
		if char == '"' {
			break
		} else {
			buffer.WriteString(string(char))
		}
	}
	if len(buffer.String()) == len(json) {
		panic("Expected end-of-string quote")
	}
	return buffer.String(), len(buffer.String())
}

func lexInt(json []byte) (int, int) {
	var buffer bytes.Buffer
	compatibleChars := []string{"-", "e", "."}
	for i := 0; i < 10; i++ {
		compatibleChars = append(compatibleChars, strconv.Itoa(i))
	}
	for _, char := range json {
		if slices.Contains(compatibleChars, string(char)) {
			buffer.WriteString(string(char))
		} else {
			break
		}
	}
	res, err := strconv.Atoi(buffer.String())
	if err != nil {
		return 0, 0
	}
	return res, len(buffer.String())
}

func lexFloat(json []byte) (float64, int) {
	var buffer bytes.Buffer
	compatibleChars := []string{"-", "e", "."}
	for i := 0; i < 10; i++ {
		compatibleChars = append(compatibleChars, strconv.Itoa(i))
	}
	for _, char := range json {
		if slices.Contains(compatibleChars, string(char)) {
			buffer.WriteString(string(char))
		} else {
			break
		}
	}
	res, err := strconv.ParseFloat(buffer.String(), 64)
	if err != nil {
		return float64(0.0), 0
	}
	return res, len(buffer.String())
}

func lexBool(json []byte) (bool, int) {
	var buffer bytes.Buffer
	for i, char := range json {
		buffer.WriteString(string(char))
		if i == 3 && buffer.String() == "true" {
			res, err := strconv.ParseBool(buffer.String())
			if err != nil {
				return false, 0
			}
			return res, len(buffer.String())
		}
		if i == 4 && buffer.String() == "false" {
			res, err := strconv.ParseBool(buffer.String())
			if err != nil {
				return false, 0
			}
			return res, len(buffer.String())
		}
		if i > 4 {
			break
		}
	}
	return false, 0
}

func lexNil(json []byte) (string, int) {
	var buffer bytes.Buffer
	for i, char := range json {
		buffer.WriteString(string(char))
		if i == 3 && buffer.String() == "null" {
			return buffer.String(), 4
		}
		if i > 3 {
			break
		}
	}
	return "", 0
}
