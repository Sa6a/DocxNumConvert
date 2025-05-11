package main

import "github.com/beevik/etree"

func CleanDocument(documentRoot *etree.Element) {

	xpaths := []string{
		"//w:pPrChange//w:numPr",
		"//w:rPrChange//w:numPr",
		"//w:p[@w:rsidDel]//w:numPr",
	}

	for _, xpath := range xpaths {
		elementsToRemove := findAllElements(documentRoot, xpath)
		for _, numPr := range elementsToRemove {
			parent := numPr.Parent()
			if parent != nil {
				parent.RemoveChild(numPr)
			}
		}
	}
}
