package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestLexerAndParserEmpty(t *testing.T) {
	testFile := "./testStep1/valid.json"
	file, err := os.Open(testFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	byteValue, _ := io.ReadAll(file)
	expectedResult := map[string]interface{}{}
	result, _ := parser(byteValue)

	if len(result) != len(expectedResult) {
		t.Errorf("result %v, expected %v", len(result), len(expectedResult))
	}
}

func TestLexerAndParserInvalidStrings(t *testing.T) {
	testFiles := []string{
		"./testStep1/invalid.json",
		"./testStep2/invalid.json",
		"./testStep2/invalid2.json",
		"./testStep3/invalid.json",
		"./testStep4/invalid.json",
	}
	for _, testFile := range testFiles {
		file, err := os.Open(testFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		byteValue, _ := io.ReadAll(file)
		result, err := parser(byteValue)
		fmt.Println(result)
		fmt.Println(err)

		if err == nil {
			t.Errorf("Expected error not nil, got nil")
		}
	}
}

func TestLexerAndParserValidStrings(t *testing.T) {
	tests := map[string]map[string]interface{}{
		"./testStep2/valid.json":  {"key": "value"},
		"./testStep2/valid2.json": {"key": "value", "key2": "value"},
		"./testStep3/valid.json": {
			"key1": true, "key2": false, "key3": nil, "key4": "value", "key5": 101,
		},
		"./testStep3/valid2.json": {
			"key1": map[string]interface{}{"nestedKey": 1, "nestedKey2": 2},
			"key2": "2024-02-23",
		},
		"./testStep4/valid.json": {
			"key":   "value",
			"key-n": 101,
			"key-o": map[string]interface{}{},
			"key-l": []interface{}{},
		},
		"./testStep4/valid2.json": {
			"key":   "value",
			"key-n": 101,
			"key-o": map[string]interface{}{"inner key": "inner value"},
			"key-l": []interface{}{"list value"},
		},
		"./testStep4/valid3.json": {
			"key":   "value",
			"key-n": 101,
			"key-o": map[string]interface{}{"inner key": "inner value"},
			"key-l": []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}},
		},
	}
	for testFile, expectedResult := range tests {
		file, err := os.Open(testFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		byteValue, _ := io.ReadAll(file)
		result, err := parser(byteValue)
		if !reflect.DeepEqual(expectedResult, result) {
			fmt.Println(err)
			fmt.Println(result)
			fmt.Println(expectedResult)
			t.Errorf("The expected map and the result are not equal.")
		}
	}
}
