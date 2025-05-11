package main

import (
	"strconv"

	"github.com/beevik/etree"
)

type AbstractLvlData struct {
	Format string
	Text   string
	Start  int
}

type NumberingParser struct {
	AbstractNumberingData map[string]map[string]AbstractLvlData
	NumberingDefinitions  map[string]*NumberingDefinition
}

func NewNumberingParser() *NumberingParser {
	return &NumberingParser{
		AbstractNumberingData: make(map[string]map[string]AbstractLvlData),
		NumberingDefinitions:  make(map[string]*NumberingDefinition),
	}
}

func (np *NumberingParser) ParseNumberingXML(numberingXMLContent []byte) error {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(numberingXMLContent); err != nil {
		return err
	}
	numberingRoot := doc.Root()

	np.parseAbstractNumbering(numberingRoot)
	np.parseNumbering(numberingRoot)
	return nil
}

func (np *NumberingParser) parseAbstractNumbering(numberingRoot *etree.Element) {
	for _, abstractNum := range findAllElements(numberingRoot, "//w:abstractNum") {
		abstractNumID, ok := getAttribute(abstractNum, "abstractNumId")
		if !ok || abstractNumID == "" {
			continue
		}

		np.AbstractNumberingData[abstractNumID] = make(map[string]AbstractLvlData)

		for _, lvl := range findAllElements(abstractNum, ".//w:lvl") {
			ilvl, okLvl := getAttribute(lvl, "ilvl")
			if !okLvl || ilvl == "" {
				continue
			}

			numFmt := "decimal"
			if numFmtElement := findElement(lvl, ".//w:numFmt"); numFmtElement != nil {
				if val, okVal := getAttribute(numFmtElement, "val"); okVal && val != "" {
					numFmt = val
				}
			}

			lvlText := "%1."
			if lvlTextElement := findElement(lvl, ".//w:lvlText"); lvlTextElement != nil {
				if val, okVal := getAttribute(lvlTextElement, "val"); okVal && val != "" {
					lvlText = val
				}
			}

			start := 1
			if startElement := findElement(lvl, ".//w:start"); startElement != nil {
				if startStr, okVal := getAttribute(startElement, "val"); okVal && startStr != "" {
					if s, err := strconv.Atoi(startStr); err == nil {
						start = s
					}
				}
			}

			np.AbstractNumberingData[abstractNumID][ilvl] = AbstractLvlData{
				Format: numFmt,
				Text:   lvlText,
				Start:  start,
			}
		}
	}
}

func (np *NumberingParser) parseNumbering(numberingRoot *etree.Element) {
	for _, num := range findAllElements(numberingRoot, "//w:num") {
		numID, ok := getAttribute(num, "numId")
		if !ok || numID == "" {
			continue
		}

		abstractNumIDVal := ""
		if abstractNumIDElement := findElement(num, ".//w:abstractNumId"); abstractNumIDElement != nil {
			if val, okVal := getAttribute(abstractNumIDElement, "val"); okVal {
				abstractNumIDVal = val
			}
		}

		if abstractNumIDVal == "" {
			continue
		}
		abstractData, dataExists := np.AbstractNumberingData[abstractNumIDVal]
		if !dataExists {
			continue
		}

		numDef := NewNumberingDefinition(abstractNumIDVal)

		for lvlID, lvlData := range abstractData {
			level := NewNumberingLevel(lvlData.Format, lvlData.Text, lvlData.Start)
			numDef.AddLevel(lvlID, level)
		}

		for _, lvlOverride := range findAllElements(num, ".//w:lvlOverride") {
			ilvl, okLvl := getAttribute(lvlOverride, "ilvl")
			if !okLvl || ilvl == "" {
				continue
			}

			levelToOverride, levelExists := numDef.Levels[ilvl]
			if !levelExists {
				continue
			}

			if startOverrideElement := findElement(lvlOverride, ".//w:startOverride"); startOverrideElement != nil {
				if newStartStr, okVal := getAttribute(startOverrideElement, "val"); okVal && newStartStr != "" {
					if newStart, err := strconv.Atoi(newStartStr); err == nil {
						levelToOverride.StartValue = newStart
						levelToOverride.CurrentValue = newStart
					}
				}
			}
		}
		np.NumberingDefinitions[numID] = numDef
	}
}
