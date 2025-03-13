package util

import (
	"fmt"
	"strings"
)

func ExtractPythonCodeFromNotebookFile(path string) (string, error) {
	text, err := ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Failed to read notebook: '%w'.", err)
	}

	notebook, err := JSONMapFromString(text)
	if err != nil {
		return "", fmt.Errorf("Failed to parse notebook as JSON: '%w'.", err)
	}

	result, err := ExtractPythonCodeFromNotebookJSON(notebook)
	if err != nil {
		return "", fmt.Errorf("Failed to extract code from notebook file '%s': '%w'.", path, err)
	}

	return result, nil
}

func ExtractPythonCodeFromNotebookJSON(notebook map[string]any) (string, error) {
	rawCells, ok := notebook["cells"]
	if !ok {
		return "", nil
	}

	cells, ok := rawCells.([]any)
	if !ok {
		return "", fmt.Errorf("Field 'cells is not a list: '%T'.", rawCells)
	}

	cellContents := []string{}
	for i, rawCell := range cells {
		cell, ok := rawCell.(map[string]any)
		if !ok {
			return "", fmt.Errorf("Cell at index %d is not a JSON object: '%T'.", i, rawCell)
		}

		if cell["cell_type"] != "code" {
			continue
		}

		rawLines, ok := cell["source"].([]any)
		if !ok {
			return "", fmt.Errorf("Cell at index %d does not have an array source: '%T'.", i, cell["source"])
		}

		lines := make([]string, 0, len(rawLines))
		for j, rawLine := range rawLines {
			line, ok := rawLine.(string)
			if !ok {
				return "", fmt.Errorf("Cell at index %d has a source line at index %d that is not a string: '%T'.", i, j, rawLine)
			}

			lines = append(lines, line)
		}

		cellContents = append(cellContents, strings.Join(lines, ""))
	}

	return strings.Join(cellContents, "\n\n") + "\n", nil
}
