package xutil

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	ftp4go "github.com/shenshouer/ftp4go"
)

type XFtp struct {
	Addr            string
	User            string
	Pwd             string
	PASV            string
	FilePattern     string
	LocalFilePrefix string
	Conn            *ftp4go.FTP
}

func (c *XFtp) Connect() (err error) {
	addra := strings.Split(c.Addr, ":")
	host := addra[0]
	port, _ := strconv.Atoi(addra[1])

	c.Conn = ftp4go.NewFTP(0) // 1 for debugging
	_, err = c.Conn.Connect(host, port, "")
	if err != nil {
		return err
	}

	_, err = c.Conn.Login(c.User, c.Pwd, "")
	if err != nil {
		return err
	}
	if c.PASV == "PORT" {
		c.Conn.SetPassive(false)
	}
	return nil
}

func (c XFtp) NameList() (ftpfiles []string) {
	files, err := c.Conn.Nlst(filepath.Dir(c.FilePattern))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, v := range files {
		if ok, _ := filepath.Match(c.FilePattern, v); ok {
			ftpfiles = append(ftpfiles, v)
		}
	}
	return ftpfiles
}


func (c XFtp) DownloadFiles() (dat map[string]string, err error) {
	files:=c.NameList()
	dat = make(map[string]string, 0)
	if c.LocalFilePrefix != ""{
		x,dir,_:= IsFileExist(c.LocalFilePrefix)
		if x&&dir{
			c.LocalFilePrefix = filepath.Dir(c.LocalFilePrefix +string(filepath.Separator)) + string(filepath.Separator)
		}
	}
	fmt.Println("DownloadFiles begin")
	for _, file := range files {
		if c.LocalFilePrefix == "" {
			c.LocalFilePrefix = time.Now().Format("20060102150405") + "_"
		}
		localpath := c.LocalFilePrefix + filepath.Base(file)
		fmt.Println("DownloadFile "+file+" to "+localpath)
		err = c.Conn.DownloadFile(file, localpath, false)
		if err != nil {
			return dat, err
		}
		dat[file] = localpath
	}
	fmt.Println("DownloadFiles end")
	return dat, nil
}

func (c XFtp) Logout() error {
	_, err := c.Conn.Quit()
	return err
}

func (c *XFtp) ConnectAndDownload() (files map[string]string, err error) {
	err = c.Connect()
	if err != nil {
		return nil, err
	}
	defer c.Logout()
	files, err = c.DownloadFiles()
	if err != nil {
		return nil, err
	}
	return files, nil
}
