package xutil

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	ftp "github.com/jlaffaye/ftp"
)

type Ftp struct {
	Host        string
	User        string
	Pwd         string
	Conn        *ftp.ServerConn
	DisableEPSV bool
	Timeout     time.Duration
}

func (c *Ftp) OpenFTP() (err error) {
	if c.Timeout == 0 {
		c.Timeout = 10*time.Second
	}
	c.Conn, err = ftp.Dial(c.Host, ftp.DialWithDisabledEPSV(c.DisableEPSV), ftp.DialWithTimeout(c.Timeout))
	if err != nil {
		return err
	}
	err = c.Conn.Login(c.User, c.Pwd)
	if err != nil {
		return err
	}
	return nil
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

func (c Ftp) FileFTP(pattern []string) (files map[string][]byte, err error) {
	err = c.OpenFTP()
	if err != nil {
		return nil, err
	}
	files, err = c.FilesByPattern(pattern)

	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c Ftp) Logout() error {
	return c.Conn.Logout()
}
