package xutil

import (
	"encoding/csv"
	"os"
)

//IsFileExist 文件是否存在
func IsFileExist(path string) (isExist, isDir bool, err error) {
	fi, err := os.Stat(path)
	if err == nil {
		return true, fi.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, false, errors.New("no such file or dir")
	}
	return false, false, err
}


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
