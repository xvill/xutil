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
	"runtime"
	"strconv"
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

// Sqlldr 执行成功返回入库记录数,失败则保留log和data到baddir
func Sqlldr(timeflag, userid, data, control, baddir string) (
	rows, badrows int, err error) {
	if control == "" {
		control = fmt.Sprintf("%s.ctl", data)
	}

	logfile := fmt.Sprintf("%s/%s.%s.log", baddir, path.Base(data), timeflag)
	badfile := fmt.Sprintf("%s/%s.%s.bad", baddir, path.Base(data), timeflag)

	var cmdout []byte
	cmd := fmt.Sprintf("sqlldr userid=%s data=%s control=%s log=%s bad=%s", userid, data, control, logfile, badfile)
	if runtime.GOOS == "windows" {
		cmdout, err = exec.Command("cmd", "/C", cmd).CombinedOutput()
	} else {
		cmdout, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	}
	rows, badrows, _ = sqlldrLog(logfile)

	// 保留入库文件策略
	if err == nil { // 执行成功
		os.Remove(logfile)
		os.Remove(data)
		return rows, badrows, nil
	}
	if badrows > 0 { // 执行成功但有错误数据,保留log和bad文件
		os.Remove(data)
		return rows, badrows, nil
	}
	return rows, badrows, errors.New(string(cmdout)) //执行失败
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
