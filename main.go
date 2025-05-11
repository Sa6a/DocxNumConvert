package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func AddNumberingToDocx(inputDocxPath, outputDocxPath string) (bool, error) {
	processor := NewDocxNumberingProcessor()
	return processor.Process(inputDocxPath, outputDocxPath)
}

func getInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func askYesNo(prompt string) bool {
	for {
		answer := strings.ToLower(getInput(prompt + " (да/нет): "))
		if answer == "да" || answer == "д" || answer == "yes" || answer == "y" {
			return true
		}
		if answer == "нет" || answer == "н" || answer == "no" || answer == "n" {
			return false
		}
		fmt.Println("Пожалуйста, ответьте 'да' или 'нет'.")
	}
}

func isPandocInstalled() bool {
	_, err := exec.LookPath("pandoc")
	return err == nil
}

func logErrorAndExit(message string, originalErr error) {
	errorLogFilename := "DocxNumConvert_error.log"
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessageToFile := fmt.Sprintf("[%s] %s", timestamp, message)
	if originalErr != nil {
		logMessageToFile += fmt.Sprintf("\n  Подробности ошибки: %v", originalErr)
	}
	logMessageToFile += "\n\n" // Добавим пустые строки для разделения записей

	// Выводим сообщение в консоль
	log.Println("-----------------------------------------")
	log.Printf("ОШИБКА: %s", message)
	if originalErr != nil {
		log.Printf("  Подробности: %v", originalErr)
	}

	// Пытаемся записать в файл
	f, fileErr := os.OpenFile(errorLogFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		log.Printf("  КРИТИЧЕСКАЯ ПОД-ОШИБКА: Не удалось записать детали в лог-файл '%s': %v", errorLogFilename, fileErr)
	} else {
		defer f.Close()
		if _, writeErr := f.WriteString(logMessageToFile); writeErr != nil {
			log.Printf("  КРИТИЧЕСКАЯ ПОД-ОШИБКА: Не удалось записать детали в лог-файл '%s' (ошибка записи): %v", errorLogFilename, writeErr)
		} else {
			log.Printf("  Подробности этой ошибки сохранены в файле: %s", errorLogFilename)
		}
	}
	log.Println("-----------------------------------------")
	os.Exit(1)
}

func main() {
	fmt.Println("--- Обработчик нумерации DOCX ---")

	var inputDocxPath string
	for {
		inputDocxPath = getInput("Введите полный путь к DOCX файлу для обработки: ")
		if _, err := os.Stat(inputDocxPath); err == nil {
			if strings.ToLower(filepath.Ext(inputDocxPath)) == ".docx" {
				break
			}
			fmt.Println("Файл должен иметь расширение .docx")
		} else if os.IsNotExist(err) {
			fmt.Printf("Файл не найден: %s\n", inputDocxPath)
		} else {
			log.Fatalf("Ошибка при проверке файла: %v", err)
		}
		fmt.Println("Пожалуйста, попробуйте снова.")
	}

	base := filepath.Base(inputDocxPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	outputDocxProcessedPath := filepath.Join(filepath.Dir(inputDocxPath), fmt.Sprintf("%s_numbered%s", nameWithoutExt, ext))

	fmt.Printf("Файл будет обработан и сохранен как: %s\n", outputDocxProcessedPath)

	fmt.Printf("Начинаю обработку файла: %s...\n", inputDocxPath)
	success, err := AddNumberingToDocx(inputDocxPath, outputDocxProcessedPath)
	if err != nil {
		logErrorAndExit(fmt.Sprintf("Ошибка при обработке DOCX файла '%s'", inputDocxPath), err)
	}
	if !success {
		logErrorAndExit(fmt.Sprintf("Не удалось обработать файл '%s'. AddNumberingToDocx вернул false без явной ошибки.", inputDocxPath), nil)
	}
	fmt.Printf("Файл '%s' успешно обработан и сохранен как '%s'\n", inputDocxPath, outputDocxProcessedPath)

	if !askYesNo("Хотите сконвертировать обработанный DOCX файл в другой формат?") {
		fmt.Println("Завершение работы.")
		return
	}

	if !isPandocInstalled() {
		errMsg := "Pandoc не найден. Для конвертации файлов установите Pandoc (https://pandoc.org/installing.html)."
		logErrorAndExit(errMsg, nil)
	}

	outputFormat := ""
	fmt.Println("\nДоступные форматы для конвертации (некоторые могут требовать доп. ПО, например, LaTeX для PDF):")
	fmt.Println(" - markdown (рекомендуется)")
	fmt.Println(" - gfm (GitHub Flavored Markdown)")
	fmt.Println(" - html")
	fmt.Println(" - pdf")
	fmt.Println(" - commonmark")
	fmt.Println("Или введите любой другой формат, поддерживаемый Pandoc.")
	fmt.Println("ПРЕДУПРЕЖДЕНИЕ: Pandoc не умеет работать с некоторыми таблицами и будет удалять их содержимое.")
	for {
		outputFormat = getInput("Введите желаемый формат вывода: ")
		if outputFormat != "" {
			break
		}
		fmt.Println("Формат не может быть пустым.")
	}

	trackChanges := "all"
	fmt.Println("\nРежим отслеживания изменений при конвертации:")
	fmt.Println(" - all (сохранить все изменения и комментарии)")
	fmt.Println(" - accept (принять все изменения)")
	fmt.Println(" - reject (отклонить все изменения)")
	userTrackChanges := getInput(fmt.Sprintf("Введите режим отслеживания изменений (или Enter для '%s'): ", trackChanges))
	if userTrackChanges != "" {
		validOptions := []string{"all", "accept", "reject"}
		isValid := false
		for _, opt := range validOptions {
			if strings.ToLower(userTrackChanges) == opt {
				trackChanges = strings.ToLower(userTrackChanges)
				isValid = true
				break
			}
		}
		if !isValid {
			fmt.Printf("Предупреждение: Введенный режим '%s' не распознан. Будет использован режим '%s'.\n", userTrackChanges, trackChanges)
		}
	}

	fmt.Printf("\nНачинаю конвертацию файла '%s' в формат '%s'...\n", outputDocxProcessedPath, outputFormat)
	convertedFilePath, err := ConvertDocxToFormat(outputDocxProcessedPath, outputFormat, trackChanges)
	if err != nil {
		errMsgForLog := fmt.Sprintf("Ошибка при конвертации файла '%s' в формат '%s'.\nВозможная команда Pandoc: pandoc \"%s\" -f docx -t %s -o ... --track-changes=%s",
			outputDocxProcessedPath, outputFormat, outputDocxProcessedPath, outputFormat, trackChanges)
		logErrorAndExit(errMsgForLog, err)
	} else {
		fmt.Printf("Файл успешно сконвертирован и сохранен как: %s\n", convertedFilePath)
	}

	fmt.Println("Завершение работы.")
}
