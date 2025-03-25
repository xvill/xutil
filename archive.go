package xutil

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/axgle/mahonia"
)

// UncompressGzip 解析 csv 或 gz 文件 输出 []byte
func UncompressGzip(fname string) ([]byte, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return UncompressReader(file, fname)
}

// UncompressReader 处理单个文件
func UncompressReader(file io.Reader, name string) ([]byte, error) {
	if strings.HasSuffix(name, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		file = gzReader
	}
	return io.ReadAll(file)
}

// UncompressTarGzip 处理 tar,tar.gz
func UncompressTarGzip(fname string) (map[string][]byte, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	var tarReader *tar.Reader
	defer file.Close()
	if strings.HasSuffix(strings.ToLower(fname), ".tar.gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else if strings.HasSuffix(strings.ToLower(fname), ".tar") {
		tarReader = tar.NewReader(file)
	}
	result := make(map[string][]byte)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ext := strings.ToLower(filepath.Ext(header.Name))
		switch ext {
		case ".csv", ".txt", ".gz":
			content, err := UncompressReader(tarReader, header.Name)
			if err != nil {
				return nil, err
			}
			result[header.Name] = content
		}
	}

	return result, nil
}

// UncompressZip 处理 zip
func UncompressZip(fname string) (map[string][]byte, error) {
	zipReader, err := zip.OpenReader(fname)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	result := make(map[string][]byte)

	for _, file := range zipReader.File {
		ext := strings.ToLower(filepath.Ext(file.Name))
		switch ext {
		case ".csv", ".txt", ".gz":
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			content, err := UncompressReader(rc, file.Name)
			if err != nil {
				return nil, err
			}
			result[file.Name] = content
		}
	}

	return result, nil
}

// Uncompres 解析 CSV 及相关压缩文件
func UncompresToByte(fname string) (fmap map[string][]byte, err error) {
	ext := strings.ToLower(filepath.Ext(fname))
	fmap = make(map[string][]byte)
	if strings.HasSuffix(strings.ToLower(fname), "tar.gz") {
		return UncompressTarGzip(fname)
	} else {
		switch ext {
		case ".csv", ".txt", ".gz":
			data, err := UncompressGzip(fname)
			if err != nil {
				return nil, err
			}
			fmap[fname] = data
		case ".zip":
			return UncompressZip(fname)
		default:
			return nil, fmt.Errorf("不支持的文件类型: %s", ext)
		}
	}

	return fmap, nil
}

// UncompressCSVBytes 解析 CSV 字节数据
func CSVBytes(bytesData []byte, encoding string) ([][]string, error) {
	var reader io.Reader
	if encoding == "" {
		reader = bytes.NewReader(bytesData)
	} else {
		reader = mahonia.NewDecoder(encoding).NewReader(bytes.NewReader(bytesData))
	}
	csvReader := csv.NewReader(reader)

	// 设置较大的缓冲区大小，减少读取次数
	csvReader.ReuseRecord = true
	csvReader.LazyQuotes = true

	var records [][]string
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// 复制记录以避免重用问题
		recordCopy := make([]string, len(record))
		copy(recordCopy, record)
		records = append(records, recordCopy)

	}

	return records, nil
}

func demo() {
	//支持 csv 组成的tar.gz,gz,zip文件
	//TODO 增加编码GBK
	fnames := []string{
		// "D:/demo.csv",
		// "D:/demo.txt",
		// "D:/demo.csv.gz", // 替换为实际的文件路径
		// "D:/demo-1.zip",
		// "D:/demo.zip",
		// "D:/D.tar.gz",
	}
	for _, fname := range fnames {
		// encoding := "GBK" // 替换为实际的文件编码
		encoding := "" // 替换为实际的文件编码
		reader, err := UncompresToByte(fname)
		fmt.Println("===========================", fname, "========================")
		fmt.Println(err)
		for k, v := range reader {
			fmt.Println("===========================", k)
			rawData, err := CSVBytes(v, encoding)
			fmt.Println(len(rawData))
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			for i, v := range rawData {
				fmt.Println(i, v)
			}

		}
	}

}
