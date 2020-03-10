package xutil

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"encoding/csv"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//FileLinesCount 文件行数计算
func FileLinesCount(filename string, delim byte) int {
	count := 0
	file, err := os.Open(filename)
	if err != nil {
		return count
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	line := []byte{}
	for {
		line, err = reader.ReadBytes(delim)
		if err != nil {
			break
		}
		count++
	}
	if len(line) > 0 {
		count++
	}
	return count
}

//FilePatternLinesCount 匹配文件行数计算
func FilePatternLinesCount(fexp string, delim byte) (count int, detail map[string]int) {
	detail = make(map[string]int, 0)
	files, err := filepath.Glob(fexp)
	if err != nil {
		return count, nil
	}
	for _, fname := range files {
		n := FileLinesCount(fname, delim)
		detail[fname] = n
		count = count + n
	}
	return count, detail
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

//--------------------------------------------------------------------------------------------------------------------

type SSftp struct {
	Ssh  *ssh.Client
	Sftp *sftp.Client
}

func NewSSftp(user, passwd, hostport string) (s SSftp, err error) {
	config := ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	conn, err := ssh.Dial("tcp", hostport, &config)
	if err != nil {
		return s, err
	}
	c, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return s, err
	}
	s.Ssh = conn
	s.Sftp = c
	return s, nil
}

func (s SSftp) Put(scpfiles [][2]string) error {
	for _, v := range scpfiles {
		srcFile, err := os.Open(v[0])
		if err != nil {
			return err
		}
		dstFile, err := s.Sftp.Create(v[1])
		if err != nil {
			return err
		}
		buf := make([]byte, 1024)
		_, err = io.CopyBuffer(dstFile, srcFile, buf)
		srcFile.Close()
		dstFile.Close()
	}
	return nil
}

func (s SSftp) Get(scpfiles [][2]string) error {
	for _, v := range scpfiles {
		srcFile, err := s.Sftp.Open(v[0])
		if err != nil {
			return err
		}
		dstFile, err := os.Create(v[1])
		if err != nil {
			return err
		}
		_, err = srcFile.WriteTo(dstFile)
		if err != nil {
			return err
		}
		srcFile.Close()
		dstFile.Close()
	}
	return nil
}

func (s SSftp) Close() error {
	err := s.Sftp.Close()
	if err != nil {
		return err
	}
	err = s.Ssh.Close()
	return err
}

//---------------------------------------------------------------------

func FileSize(fname string) int64 {
	file, err := os.Stat(fname)

	//if file can't be found create an err message and close program
	if os.IsNotExist(err) {
		log.Fatal("File does not exsist at: ", fname)
	}
	//if any filesystem error occurs close the program with err details
	if err != nil {
		log.Fatal(err)
	}

	// get and return the size of the file
	return file.Size()
}

// FileHash Hash
func FileHash(htype string, fname string) ([]byte, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var h hash.Hash
	switch htype {
	case "SHA1":
		h = sha1.New()

	case "MD5":
		h = md5.New()
	default:
		h = md5.New()
	}

	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// FileSHA1 SHA1
func FileSHA1(fname string) ([]byte, error) {
	return FileHash("SHA1", fname)
}

// FileMD5  MD5
func FileMD5(fname string) ([]byte, error) {
	return FileHash("MD5", fname)
}

// FileCopy 文件/文件夹/链接复制
func FileCopy(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	return _copy(src, dest, info)
}

func _copy(src, dest string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return linkcopy(src, dest, info)
	}
	if info.IsDir() {
		return dircopy(src, dest, info)
	}
	return filecopy(src, dest, info)
}

func filecopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dircopy(srcdir, destdir string, info os.FileInfo) error {
	originalMode := info.Mode()
	if err := os.MkdirAll(destdir, 0755); err != nil {
		return err
	}
	// Recover dir mode with original one.
	defer os.Chmod(destdir, originalMode)
	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := _copy(cs, cd, content); err != nil {
			// If any error, exit immediately
			return err
		}
	}

	return nil
}

func linkcopy(src, dest string, info os.FileInfo) error {
	src, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

//---------------------------------------------------------------------
