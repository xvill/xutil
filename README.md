# xtools


## function

```
CsvWriteALL 生成CSV
func CsvWriteALL(data [][]string, wfile string, comma rune) error 

Sqlldr 执行成功返回入库记录数,失败则保留log和data到baddir
func Sqlldr(timeflag, userid, data, control, baddir string) (
	rows, badrows int, err error) 

//IsFileExist 文件是否存在
func IsFileExist(path string) (isExist, isDir bool, err error) 
```
