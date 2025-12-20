package cli

import (
	"fmt"
	"strings"
)

func PrintTable(rows [][]string) {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	colWidths := make([]int, maxCols)
	for _, row := range rows {
		for j, col := range row {
			if len(col) > colWidths[j] {
				colWidths[j] = len(col)
			}
		}
	}

	for i, row := range rows {
		fmt.Print("| ")
		for j, col := range row {
			padChars := colWidths[j] - len(col)
			pad := strings.Repeat(" ", padChars)
			fmt.Printf("%s%s | ", col, pad)
		}
		fmt.Print("\n")

		if i == 0 {
			fmt.Print("| ")
			for _, colWidth := range colWidths {
				pad := strings.Repeat("-", colWidth)
				fmt.Printf("%s | ", pad)
			}
			fmt.Print("\n")
		}
	}
}
