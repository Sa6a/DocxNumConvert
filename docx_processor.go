package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

type DocxNumberingProcessor struct {
	NumberingParser *NumberingParser
}

func NewDocxNumberingProcessor() *DocxNumberingProcessor {
	return &DocxNumberingProcessor{
		NumberingParser: NewNumberingParser(),
	}
}

func (dnp *DocxNumberingProcessor) Process(inputDocxPath, outputDocxPath string) (bool, error) {
	tempDir := inputDocxPath + "_temp"

	if _, err := os.Stat(tempDir); err == nil {
		if err := os.RemoveAll(tempDir); err != nil {
			return false, fmt.Errorf("ошибка удаления временной директории: %w", err)
		}
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return false, fmt.Errorf("ошибка создания временной директории: %w", err)
	}

	defer os.RemoveAll(tempDir)

	if err := unzip(inputDocxPath, tempDir); err != nil {
		return false, fmt.Errorf("ошибка распаковки DOCX: %w", err)
	}

	if err := dnp.processFiles(tempDir); err != nil {
		return false, fmt.Errorf("ошибка обработки файлов: %w", err)
	}

	if err := zipSource(tempDir, outputDocxPath); err != nil {
		return false, fmt.Errorf("ошибка создания DOCX: %w", err)
	}

	return true, nil
}

func (dnp *DocxNumberingProcessor) processFiles(tempDir string) error {
	numberingPath := filepath.Join(tempDir, "word", "numbering.xml")
	if _, err := os.Stat(numberingPath); err == nil {
		content, err := os.ReadFile(numberingPath)
		if err != nil {
			return fmt.Errorf("ошибка чтения numbering.xml: %w", err)
		}
		if err := dnp.NumberingParser.ParseNumberingXML(content); err != nil {
			return fmt.Errorf("ошибка парсинга numbering.xml: %w", err)
		}
	}

	documentPath := filepath.Join(tempDir, "word", "document.xml")
	if _, err := os.Stat(documentPath); err == nil {
		content, err := os.ReadFile(documentPath)
		if err != nil {
			return fmt.Errorf("ошибка чтения document.xml: %w", err)
		}

		modifiedDocument, err := dnp.processDocument(content)
		if err != nil {
			return fmt.Errorf("ошибка обработки document.xml: %w", err)
		}

		if err := os.WriteFile(documentPath, modifiedDocument, 0644); err != nil {
			return fmt.Errorf("ошибка записи document.xml: %w", err)
		}
	}
	return nil
}

func (dnp *DocxNumberingProcessor) processDocument(documentContent []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(documentContent); err != nil {
		return nil, err
	}
	documentRoot := doc.Root()

	CleanDocument(documentRoot)

	paragraphFormatter := NewParagraphFormatter(dnp.NumberingParser.NumberingDefinitions)

	for _, paragraph := range findAllElements(documentRoot, "//w:p") {
		numPrefix := paragraphFormatter.FormatParagraph(paragraph)

		if numPrefix != "" {
			dnp.addNumberingToParagraph(paragraph, numPrefix)
		}
		dnp.removeNumPrTags(paragraph)
	}
	doc.Indent(2)
	return doc.WriteToBytes()
}

func (dnp *DocxNumberingProcessor) addNumberingToParagraph(paragraph *etree.Element, numPrefix string) {

	var firstT *etree.Element
	if firstRElement := findElement(paragraph, ".//w:r"); firstRElement != nil {
		firstT = findElement(firstRElement, ".//w:t")
	}

	if firstT != nil {
		currentText := firstT.Text()
		space := " "

		firstT.SetText(fmt.Sprintf("%s%s%s", numPrefix, space, currentText))
	} else {

		rElement := createElement("r", nil)
		tElement := createElement("t", nil)

		tElement.CreateAttr("xml:space", "preserve")
		tElement.SetText(numPrefix + " ")

		rElement.AddChild(tElement)

		existingChild := paragraph.ChildElements()
		if len(existingChild) > 0 {
			paragraph.InsertChild(existingChild[0], rElement)
		} else {
			paragraph.AddChild(rElement)
		}
	}
}

func (dnp *DocxNumberingProcessor) removeNumPrTags(paragraph *etree.Element) {

	for _, numPr := range findAllElements(paragraph, "./w:pPr/w:numPr") {
		parent := numPr.Parent()
		if parent != nil {
			parent.RemoveChild(numPr)
		}
	}
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("недопустимый путь к файлу: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func zipSource(source, target string) error {
	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	zipWriter := zip.NewWriter(targetFile)
	defer zipWriter.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == source {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		_, err = io.Copy(writer, fileToZip)
		return err
	})
}
