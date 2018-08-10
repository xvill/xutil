package xutil

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

type Ftp struct {
	Conn *ftp.ServerConn
}

func OpenFtp(host, user, pwd string, timeout time.Duration) (Ftp, error) {
	conn, err := ftp.DialTimeout(host, timeout)
	if err != nil {
		return Ftp{Conn: conn}, err
	}
	err = conn.Login(user, pwd)
	if err != nil {
		return Ftp{Conn: conn}, err
	}
	return Ftp{Conn: conn}, nil
}

func (c Ftp) NameList(pattern string) (ftpfiles []string) {
	files, err := c.Conn.NameList(filepath.Dir(pattern))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, v := range files {
		if ok, _ := filepath.Match(pattern, v); ok {
			ftpfiles = append(ftpfiles, v)
		}
	}
	return ftpfiles
}

func (c Ftp) FilesByPattern(pattern []string) (dat map[string][]byte, err error) {
	files := make([]string, 0)
	for _, par := range pattern {
		files = append(files, c.NameList(par)...)
	}
	return c.Files(files)
}

func (c Ftp) Files(files []string) (dat map[string][]byte, err error) {
	dat = make(map[string][]byte, 0)
	for _, file := range files {
		r, err := c.Conn.Retr(file)
		if err != nil {
			return dat, err
		}
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return dat, err
		}
		dat[file] = buf
		r.Close()
	}
	return dat, nil
}

func (c Ftp) Logout() error {
	return c.Conn.Logout()
}

func FileFTP(host, user, pwd string, pattern []string) (files map[string][]byte, err error) {
	ftp, err := OpenFtp(host, user, pwd, 10*time.Second)
	if err != nil {
		return nil, err
	}
	files, err = ftp.FilesByPattern(pattern)

	if err != nil {
		return nil, err
	}
	ftp.Logout()

	return files, nil
}
