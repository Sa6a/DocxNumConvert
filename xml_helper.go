package main

import (
	"github.com/beevik/etree"
)

const (
	wordNS                 = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	wordProcessingMLPrefix = "w"
)

func findElement(parent *etree.Element, xpath string) *etree.Element {
	return parent.FindElement(xpath)
}

func findAllElements(parent *etree.Element, xpath string) []*etree.Element {
	return parent.FindElements(xpath)
}

func getAttribute(element *etree.Element, attrNameLocal string) (string, bool) {
	if element == nil {
		return "", false
	}
	attr := element.SelectAttr(wordProcessingMLPrefix + ":" + attrNameLocal)
	if attr == nil {
		attr = element.SelectAttr(attrNameLocal)
		if attr == nil {
			return "", false
		}
	}
	return attr.Value, true
}

func parseNumberingInfo(numPrElement *etree.Element) (ilvl string, numID string, found bool) {
	if numPrElement == nil {
		return "", "", false
	}

	ilvlElement := findElement(numPrElement, ".//w:ilvl")
	numIDElement := findElement(numPrElement, ".//w:numId")

	var ilvlVal, numIDVal string
	var ilvlFound, numIDFound bool

	if ilvlElement != nil {
		ilvlVal, ilvlFound = getAttribute(ilvlElement, "val")
	}
	if numIDElement != nil {
		numIDVal, numIDFound = getAttribute(numIDElement, "val")
	}

	if ilvlFound && numIDFound {
		return ilvlVal, numIDVal, true
	}
	return "", "", false
}

func createElement(tagNameLocal string, attributes map[string]string) *etree.Element {
	element := etree.NewElement(wordProcessingMLPrefix + ":" + tagNameLocal)
	if attributes != nil {
		for attrNameLocal, attrValue := range attributes {
			element.CreateAttr(wordProcessingMLPrefix+":"+attrNameLocal, attrValue)
		}
	}
	return element
}
