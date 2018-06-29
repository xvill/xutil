package xutil

import (
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
	} else if rows > 0 { // 执行成功但有错误数据,保留log和bad文件
		os.Remove(data)
	} else { //执行失败
		return rows, badrows, errors.New(string(cmdout))
	}

	return rows, badrows, nil
}

//sqlldrLog 从sqlldr的log文件中获取入库记录数
func sqlldrLog(logfile string) (rows, badrows int, err error) {
	rowspat := regexp.MustCompile(`(\d+) Rows? successfully loaded`)
	rowspatbad := regexp.MustCompile(`(\d+) Rows? not loaded due to data errors`)
	src, err := ioutil.ReadFile(logfile)
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
