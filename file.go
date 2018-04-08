package xtools

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
)

// CsvWriteALL 生成CSV
func CsvWriteALL(wfile string, data [][]string, comma rune) error {
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

// Sqlldr 执行成功返回入库记录数,失败则保留log和data到baddir
func Sqlldr(timeflag, userid, data, baddir, loadeddir, control, logfile, badfile string) (
	rows, badrows int, output []byte, err error) {
	if control == "" {
		control = fmt.Sprintf("%s.ctl", data)
	}
	if logfile == "" {
		logfile = fmt.Sprintf("%s/%s.%s.log", baddir, path.Base(data), timeflag)
	}
	if badfile == "" {
		badfile = fmt.Sprintf("%s/%s.%s.bad", baddir, path.Base(data), timeflag)
	}

	cmd := fmt.Sprintf("sqlldr userid=%s data=%s control=%s log=%s bad=%s", userid, data, control, logfile, badfile)
	output, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	rows, badrows, _ = sqlldrLog(logfile)

	// 保留入库文件策略
	if err == nil { // 执行成功
		os.Remove(logfile)
		os.Remove(data)
	} else if badrows > 0 { // 执行成功但有错误数据,保留log和bad文件
		os.Remove(data)
	}
	return
}

//sqlldrLog 从sqlldr的log文件中获取入库记录数
func sqlldrLog(name string) (rows, badrows int, err error) {
	rowspat := regexp.MustCompile(`(\d+) Rows? successfully loaded`)
	rowspatbad := regexp.MustCompile(`(\d+) Rows? not loaded due to data errors`)
	src, err := ioutil.ReadFile(name)
	if err != nil {
		return
	}
	x := rowspat.FindSubmatch(src)
	y := rowspatbad.FindSubmatch(src)
	if len(x) > 1 {
		rows, err = strconv.Atoi(string(x[1]))
		if err != nil {
			return
		}
	}
	if len(y) > 1 {
		badrows, err = strconv.Atoi(string(y[1]))
		if err != nil {
			return
		}
	}
	return
}

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
