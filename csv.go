package xutil

import (
	"archive/zip"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/panjf2000/ants"
)

// ============================================================================================================
// CSVTools Csv工具包
type CSVTools struct {
	ZipFname       string
	Fnames         []string
	OnamePrefix    string
	Field          string
	ThreadSize     int
	InputInfo      map[string][]string
	OutputInfo     map[string][]string
	Rows           int
	Cols           []string
	ColsKV         map[string]string
	FileMaxLines   int
	FileHeadKeep   string
	FileField      string
	FileCols       []string                                 // 指定输出列顺序
	XMLToCSV       func(xmlData []byte) ([][]string, error) // XML转换函数
	ValueProcessor func(value string) string                // 值处理器
	ErrorHandler   func(err error, fileName string)         // 错误处理器

}

// 初始化参数
func (c *CSVTools) init() {
	c.InputInfo = make(map[string][]string)
	c.OutputInfo = make(map[string][]string)
	if c.Field == "" {
		c.Field = "," // 默认逗号分隔符
	}
	if c.FileField == "" {
		c.FileField = "," // 默认逗号分隔符
	}
	if c.ThreadSize <= 0 {
		c.ThreadSize = 10
	}
	if c.FileMaxLines <= 0 {
		c.FileMaxLines = 100000 // 默认10万行
	}
	// 初始化默认值
	if c.ErrorHandler == nil {
		c.ErrorHandler = func(err error, fileName string) {
			fmt.Printf("文件 %s 处理失败: %v\n", fileName, err)
		}
	}
	if c.FileHeadKeep != "fase" {
		c.FileHeadKeep = "true"
	}
	if len(c.ColsKV) == 0 && len(c.FileCols) > 0 {
		c.ColsKV = make(map[string]string)
		for _, v := range c.FileCols {
			c.ColsKV[v] = v
		}
	}
	for k, v := range c.ColsKV {
		c.ColsKV[strings.ToLower(k)] = strings.ToLower(v)
	}
}

// csvCombine 合并 ZIP 中或多个 CSV 文件为多个输出文件
func (c *CSVTools) ParseZip() {
	c.init()

	var (
		allData     [][]string
		wg          sync.WaitGroup
		hasHeader   bool
		mutex       sync.Mutex          // 新增互斥锁
		processFile func(i interface{}) // 定义处理单个文件的函数
	)

	processFile = func(i interface{}) {
		defer wg.Done()
		var file io.ReadCloser
		var fileName string
		var fileInfo fs.FileInfo

		if fileHeader, ok := i.(*zip.File); ok {
			var err error
			file, err = fileHeader.Open()
			if err != nil {
				fmt.Printf("无法打开 ZIP 中的文件 %s: %v\n", fileHeader.Name, err)
				return
			}
			fileName = fileHeader.Name
			fileInfo = fileHeader.FileInfo()
		} else if fname, ok := i.(string); ok {
			var err error
			file, err = os.Open(fname)
			if err != nil {
				fmt.Printf("无法打开文件 %s: %v\n", fname, err)
				return
			}
			fileName = fname
			fileInfo, err = os.Stat(fname)
			if err != nil {
				fmt.Printf("无法获取文件 %s 的信息: %v\n", fname, err)
				return
			}
		} else {
			fmt.Println("传入的参数类型错误")
			return
		}
		defer file.Close()

		// 处理压缩
		reader, err := getReader(file, fileName)
		if err != nil {
			c.ErrorHandler(err, fileName)
			return
		}

		// 解析文件内容
		lines, err := c.parseFile(reader, fileName)
		if err != nil {
			c.ErrorHandler(err, fileName)
			return
		}

		fsize := fmt.Sprintf("%d", fileInfo.Size())
		fctime := fileInfo.ModTime().Format("2006-01-02T15:04:05")
		fcnt := fmt.Sprintf("%d", len(lines))
		c.InputInfo[fileName] = []string{fctime, fsize, fcnt} // 记录文件信息

		// 合并数据并处理值
		if len(lines) > 0 {
			mutex.Lock() // 加锁
			if !hasHeader {
				allData = append(allData, lines[0])
				hasHeader = true
			}
			if c.ValueProcessor == nil {
				// 扩容 allData
				allData = append(allData, make([][]string, len(lines[1:]))...)
				// 使用 copy 函数复制数据
				copy(allData[len(allData)-len(lines[1:]):], lines[1:])
			} else {
				for _, line := range lines[1:] {
					processedLine := make([]string, len(line))
					for i, value := range line {
						processedLine[i] = c.ValueProcessor(value)
					}
					allData = append(allData, processedLine)
				}
			}
			mutex.Unlock() // 解锁
		}
	}

	// 初始化协程池
	pool, _ := ants.NewPoolWithFunc(c.ThreadSize, processFile)
	defer pool.Release()

	// 根据是否有 zip 文件或多个 CSV 文件进行处理
	if c.ZipFname != "" { // 处理 ZIP 文件
		zipReader, err := zip.OpenReader(c.ZipFname)
		if err != nil {
			fmt.Printf("无法打开 ZIP 文件 %s: %v\n", c.ZipFname, err)
			return
		}
		defer zipReader.Close()

		for _, fileHeader := range zipReader.File {
			if fileHeader.FileInfo().IsDir() {
				continue
			}
			wg.Add(1)
			_ = pool.Invoke(fileHeader)
		}
	} else if len(c.Fnames) > 0 { // 处理多个 CSV 文件
		for _, fname := range c.Fnames {
			wg.Add(1)
			_ = pool.Invoke(fname)
		}
	} else {
		fmt.Println("没有提供有效的输入文件")
		return
	}

	// 等待所有任务完成
	wg.Wait()

	// 计算输出文件的总行数和列头
	if len(allData) > 1 {
		c.Rows = len(allData) - 1 //实际数据去除列头
		c.Cols = allData[0]
	}

	// 按指定行数分割数据并写入多个文件
	fileIndex := 1
	for i := 0; i < len(allData); i += c.FileMaxLines {
		end := i + c.FileMaxLines
		if end > len(allData) {
			end = len(allData)
		}
		// 生成输出文件名
		oname := fmt.Sprintf("%s_%d.csv", c.OnamePrefix, fileIndex)

		currentData := make([][]string, 0)
		if fileIndex == 1 {
			currentData = allData[i:end]
		} else {
			currentData = append(currentData, allData[0])
			currentData = append(currentData, allData[i:end]...)
		}
		if len(c.FileCols) == 0 {
			c.FileCols = c.Cols
		}
		// fmt.Println(currentData)
		outputFileInfo, err := RowsKVFile(currentData, c.ColsKV, c.FileCols, oname, c.FileField, c.FileHeadKeep)
		if err != nil {
			fmt.Printf("RowsFile : %v\n", err)
		}

		fsize := fmt.Sprintf("%d", outputFileInfo.Size())
		fctime := outputFileInfo.ModTime().Format("2006-01-02T15:04:05")
		fcnt := fmt.Sprintf("%d", len(currentData))
		c.OutputInfo[oname] = []string{fctime, fsize, fcnt} // 记录输出文件信息

		fileIndex++

	}
}

// 获取文件阅读器（处理压缩）
func getReader(file io.ReadCloser, fileName string) (io.Reader, error) {
	reader := io.Reader(file)
	if strings.HasSuffix(fileName, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("解压失败: %w", err)
		}
		reader = gzReader
	}
	return reader, nil
}

// 解析文件内容
func (c *CSVTools) parseFile(reader io.Reader, fileName string) ([][]string, error) {
	if strings.HasSuffix(fileName, ".xml") || strings.HasSuffix(fileName, ".xml.gz") {
		xmlData, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("读取XML失败: %w", err)
		}
		return c.XMLToCSV(xmlData)
	} else if strings.HasSuffix(fileName, ".csv") || strings.HasSuffix(fileName, ".csv.gz") {
		csvReader := csv.NewReader(reader)
		csvReader.Comma = rune(c.Field[0])
		return csvReader.ReadAll()
	}
	return nil, fmt.Errorf("不支持的文件类型: %s", fileName)
}

//=====================================================================================
//RowReOrder 重新排列行数据顺序
func RowReOrder(row []string, outind []int) (newrow []string) {
	newrow = make([]string, len(outind))
	for i, ind := range outind {
		if ind >= 0 {
			newrow[i] = row[ind]
		} else {
			newrow[i] = ""
		}
	}
	return
}

//RowKVind 根据列映射和新列顺序输出新列索引值
func RowKVind(row []string, kv map[string]string, outhead []string) (outind []int) {
	colmap := make(map[string]int, 0)
	for i, col := range row {
		col = strings.ToLower(col)
		if v, exist := kv[col]; exist {
			colmap[strings.ToLower(v)] = i
		} else {
			colmap[strings.ToLower(col)] = i
		}
	}
	fmt.Printf("colmap %v\n", colmap)

	for _, newcol := range outhead {
		if ind, exist := colmap[strings.ToLower(newcol)]; exist {
			outind = append(outind, ind)
		} else {
			outind = append(outind, -1)
		}
	}
	return outind
}

func RowsKVFile(rawdat [][]string, kv map[string]string, outhead []string, oname, field, outheadKeep string) (fInfo fs.FileInfo, err error) {
	var sameHeaders bool
	var outind []int
	if len(outhead) == 0 {
		sameHeaders = true
		outhead = rawdat[0]
	} else {
		sameHeaders = len(outhead) == len(rawdat[0])
		if sameHeaders {
			for i := range outhead {
				if strings.ToLower(outhead[i]) != strings.ToLower(rawdat[0][i]) {
					sameHeaders = false
					break
				}
			}
		}
	}
	if !sameHeaders {
		outind = RowKVind(rawdat[0], kv, outhead)
	}
	fmt.Println(rawdat[0], kv, outhead)
	fmt.Println(outind)

	// 写入文件
	f, err := os.Create(oname)
	if err != nil {
		fmt.Printf("无法创建输出文件 %s: %v\n", oname, err)
		return
	}
	defer f.Close()

	// var headerToWrite []string
	writer := csv.NewWriter(f)
	writer.Comma = rune(field[0])

	if outheadKeep == "true" {
		// 写入列头
		if err := writer.Write(outhead); err != nil {
			fmt.Printf("无法写入列头到输出文件 %s: %v\n", oname, err)
		}
	}

	if sameHeaders {
		// 批量写入数据
		for _, row := range rawdat[1:] {
			if err := writer.Write(row); err != nil {
				fmt.Printf("无法写入输出文件 %s: %v\n", oname, err)
				break
			}
		}
	} else {
		// 批量写入数据
		for _, row := range rawdat[1:] {
			newRow := RowReOrder(row, outind)
			if err := writer.Write(newRow); err != nil {
				fmt.Printf("无法写入输出文件 %s: %v\n", oname, err)
				break
			}
		}
	}
	writer.Flush()
	// 获取输出文件信息
	fInfo, err = f.Stat()
	if err != nil {
		fmt.Printf("无法获取输出文件 %s 的信息: %v\n", oname, err)
		return
	}
	return
}
