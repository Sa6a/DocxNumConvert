package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertDocxToFormat(inputDocxPath, outputFormat, trackChanges string) (string, error) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		return "", fmt.Errorf("pandoc не найден в PATH. Установите pandoc: %w", err)
	}

	outputFilePath := strings.TrimSuffix(inputDocxPath, filepath.Ext(inputDocxPath)) + "." + outputFormat

	args := []string{
		inputDocxPath,
		"-f", "docx",
		"-t", outputFormat,
		"-o", outputFilePath,
	}

	validTrackChanges := map[string]bool{"accept": true, "reject": true, "all": true}
	if !validTrackChanges[trackChanges] {
		fmt.Printf("Предупреждение: некорректное значение track_changes '%s'. Используется 'all'.\n", trackChanges)
		trackChanges = "all"
	}
	args = append(args, fmt.Sprintf("--track-changes=%s", trackChanges))

	cmd := exec.Command("pandoc", args...)

	fmt.Printf("Выполнение команды: pandoc %s\n", strings.Join(args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ошибка конвертации файла с помощью pandoc: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Файл успешно сконвертирован в формат %s: %s\n", outputFormat, outputFilePath)
	fmt.Printf("Pandoc output: %s\n", string(output))
	return outputFilePath, nil
}
