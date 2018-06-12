package xutil

import (
	"encoding/csv"
	"os"
)

// CsvWriteALL 生成CSV
func CsvWriteALL(data [][]string, wfile string, comma rune) error {
	wf, err := os.Create(wfile)
	if err != nil {
		return err
	}
	defer wf.Close()
	wcsv := csv.NewWriter(wf)
	wcsv.Comma = comma
	wcsv.WriteAll(data)
	wcsv.Flush()

	return nil
}
