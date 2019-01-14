package xutil

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
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
