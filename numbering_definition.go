package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type NumberingDefinition struct {
	AbstractNumID string
	Levels        map[string]*NumberingLevel
}

func NewNumberingDefinition(abstractNumID string) *NumberingDefinition {
	return &NumberingDefinition{
		AbstractNumID: abstractNumID,
		Levels:        make(map[string]*NumberingLevel),
	}
}

func (nd *NumberingDefinition) AddLevel(levelID string, level *NumberingLevel) {
	nd.Levels[levelID] = level
}

func (nd *NumberingDefinition) ResetLevelsBelow(currentLevelID string) {
	currentLevelInt, err := strconv.Atoi(currentLevelID)
	if err != nil {
		return
	}

	for levelIDStr, level := range nd.Levels {
		levelIDInt, err := strconv.Atoi(levelIDStr)
		if err != nil {
			continue
		}
		if levelIDInt > currentLevelInt {
			level.Reset()
		}
	}
}

func (nd *NumberingDefinition) GetFormattedNumber(levelID string) string {
	level, ok := nd.Levels[levelID]
	if !ok {
		return ""
	}

	text := level.TextTemplate

	var levelIDs []string
	for id := range nd.Levels {
		levelIDs = append(levelIDs, id)
	}
	sort.Slice(levelIDs, func(i, j int) bool {
		id1, _ := strconv.Atoi(levelIDs[i])
		id2, _ := strconv.Atoi(levelIDs[j])
		return id1 < id2
	})

	for _, subLevelIDStr := range levelIDs {
		subLevel, exists := nd.Levels[subLevelIDStr]
		if !exists {
			continue
		}
		subLevelNum, _ := strconv.Atoi(subLevelIDStr)
		placeholder := fmt.Sprintf("%%%d", subLevelNum+1)

		if strings.Contains(text, placeholder) {
			text = strings.ReplaceAll(text, placeholder, subLevel.FormatCurrentValue())
		}
	}
	return text
}
