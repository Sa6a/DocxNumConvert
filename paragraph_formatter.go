package main

import (
	"strconv"

	"github.com/beevik/etree"
)

type ParagraphFormatter struct {
	NumberingDefinitions map[string]*NumberingDefinition
	LastActiveLevels     map[string]string
}

func NewParagraphFormatter(numberingDefs map[string]*NumberingDefinition) *ParagraphFormatter {
	return &ParagraphFormatter{
		NumberingDefinitions: numberingDefs,
		LastActiveLevels:     make(map[string]string),
	}
}

func (pf *ParagraphFormatter) FormatParagraph(paragraph *etree.Element) string {
	pPr := findElement(paragraph, ".//w:pPr")
	if pPr == nil {
		return ""
	}

	numPr := findElement(pPr, ".//w:numPr")
	ilvl, numID, found := parseNumberingInfo(numPr)

	numPrefix := ""
	if found && pf.hasValidNumbering(ilvl, numID) {
		numDef := pf.NumberingDefinitions[numID]

		currentLevelIlvl, _ := strconv.Atoi(ilvl)
		lastActiveIlvlStr, activeLevelExists := pf.LastActiveLevels[numID]

		if activeLevelExists {
			lastActiveIlvl, _ := strconv.Atoi(lastActiveIlvlStr)

			if currentLevelIlvl > lastActiveIlvl {
			} else if currentLevelIlvl < lastActiveIlvl {
				numDef.Levels[ilvl].Increment()
				numDef.ResetLevelsBelow(ilvl)
			} else {
				numDef.Levels[ilvl].Increment()
			}
		} else {
		}

		numPrefix = numDef.GetFormattedNumber(ilvl)
		pf.LastActiveLevels[numID] = ilvl
	} else {
		if numID != "" {
			delete(pf.LastActiveLevels, numID)
		}
	}

	return numPrefix
}

func (pf *ParagraphFormatter) hasValidNumbering(ilvl, numID string) bool {
	if ilvl == "" || numID == "" {
		return false
	}
	def, ok := pf.NumberingDefinitions[numID]
	if !ok {
		return false
	}
	if _, ok := def.Levels[ilvl]; !ok {
		return false
	}
	return true
}
