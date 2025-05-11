package main

import (
	"fmt"
	"strings"
)

type romanPair struct {
	Value  int
	Symbol string
}

var romanNumerals = []romanPair{
	{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
	{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
	{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
}

func formatNumber(number int, formatType string) string {
	switch formatType {
	case "decimal":
		return fmt.Sprintf("%d", number)
	case "upperRoman":
		return toRoman(number)
	case "lowerRoman":
		return strings.ToLower(toRoman(number))
	case "upperLetter":
		if number >= 1 && number <= 26 {
			return string('A' + number - 1)
		}
	case "lowerLetter":
		if number >= 1 && number <= 26 {
			return string('a' + number - 1)
		}
	}
	return fmt.Sprintf("%d", number)
}

func toRoman(number int) string {
	if number <= 0 {
		return ""
	}
	var result strings.Builder
	for _, pair := range romanNumerals {
		for number >= pair.Value {
			result.WriteString(pair.Symbol)
			number -= pair.Value
		}
	}
	return result.String()
}
