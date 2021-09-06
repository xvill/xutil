package xutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type XSFtp struct {
	Addr            string
	User            string
	Pwd             string
	FilePattern     []string
	LocalFilePrefix string
	SSH             *ssh.Client
	SFTP            *sftp.Client
}

func (c *XSFtp) Connect() (err error) {
	config := ssh.ClientConfig{
		User:            c.User,
		Auth:            []ssh.AuthMethod{ssh.Password(c.Pwd)},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	c.SSH, err = ssh.Dial("tcp", c.Addr, &config)
	if err != nil {
		return err
	}
	c.SFTP, err = sftp.NewClient(c.SSH)
	if err != nil {
		c.SSH.Close()
		return err
	}
	return nil
}

func (c XSFtp) Logout() error {
	err := c.SFTP.Close()
	if err != nil {
		return err
	}
	err = c.SSH.Close()
	return err
}

// ------------------------------------------------------------------------------

//FileExist 文件是否存在
func (c XSFtp) FileExist(filepath string) (bool, error) {
	if _, err := c.SFTP.Stat(filepath); err != nil {
		return false, err
	}
	return true, nil
}

//IsDir 检查远程是否是个目录
func (c XSFtp) IsDir(path string) bool {
	// 检查远程是文件还是目录
	info, err := c.SFTP.Stat(path)
	if err == nil && info.IsDir() {
		return true
	}
	return false
}

//Size 获取文件大小
func (c XSFtp) Size(path string) int64 {
	info, err := c.SFTP.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

//IsFile 检查远程是否是个文件
func (c XSFtp) IsFile(path string) bool {
	info, err := c.SFTP.Stat(path)
	if err == nil && !info.IsDir() {
		return true
	}
	return false
}
func (c XSFtp) RemoveFile(remoteFile string) error {
	return c.SFTP.Remove(remoteFile)
}

func (c *XSFtp) Cmd(cmd string) (stdout, stderr string) {
	var stdOut, stdErr bytes.Buffer
	session, _ := c.SSH.NewSession()
	defer session.Close()
	session.Stdout = &stdOut
	session.Stderr = &stdErr
	session.Run(cmd)
	return stdOut.String(), stdErr.String()
}

// ------------------------------------------------------------------------------
func (c *XSFtp) NameList() (ftpfiles []string) {
	for _, fpattern := range c.FilePattern {
		fs, _ := c.SFTP.Glob(fpattern)
		ftpfiles = append(ftpfiles, fs...)
	}
	return ftpfiles
}

// func (c *XSFtp) NameList() (ftpfiles []string) {
// 	for _, fpattern := range c.FilePattern {
// 		fdir := filepath.Dir(fpattern)
// 		walker := c.SFTP.Walk(fdir)
// 		for walker.Step() {
// 			if err := walker.Err(); err != nil {
// 				log.Println(err)
// 				continue
// 			}
// 			fname := filepath.Join(fdir, walker.Stat().Name())
// 			if ok, _ := filepath.Match(fpattern, fname); ok {
// 				ftpfiles = append(ftpfiles, fname)
// 			}
// 		}
// 	}

// 	return ftpfiles
// }

func (c *XSFtp) DownloadFiles(files []string) (dat map[string]string, err error) {
	dat = make(map[string]string, 0)
	if len(files) == 0 {
		return
	}
	if c.LocalFilePrefix != "" {
		err = IsDirsExist([]string{c.LocalFilePrefix}, false)
		if err == nil {
			c.LocalFilePrefix = filepath.Dir(c.LocalFilePrefix+string(filepath.Separator)) + string(filepath.Separator)
		} else {
			return dat, err
		}
	}

	fmt.Println("DownloadFiles begin")
	for _, file := range files {
		if c.LocalFilePrefix == "" {
			c.LocalFilePrefix = time.Now().Format("20060102150405") + "_"
		}
		localpath := c.LocalFilePrefix + filepath.Base(file)
		fmt.Println("DownloadFile " + file + " to " + localpath)

		srcFile, err := c.SFTP.Open(file)
		if err != nil {
			fmt.Println("Open", err)
			return dat, err
		}

		raw, err := ioutil.ReadAll(srcFile)
		if err != nil {
			fmt.Println("ReadAll", err)
			return dat, err
		}

		err = ioutil.WriteFile(localpath, raw, 0666)
		if err != nil {
			fmt.Println("WriteFile", err)
			return dat, err
		}
		srcFile.Close()
		dat[file] = localpath
	}
	fmt.Println("DownloadFiles end")
	return dat, nil
}

func (c *XSFtp) UploadFiles(files map[string]string) (retInfo map[string]error) {
	retInfo = make(map[string]error, 0)
	for fname, tname := range files {
		srcFile, err := os.Open(fname)
		retInfo[fname] = err
		if err != nil {
			continue
		}
		dstFile, err := c.SFTP.Create(tname)
		retInfo[fname] = err
		if err != nil {
			continue
		}
		buf := make([]byte, 1024)
		_, err = io.CopyBuffer(dstFile, srcFile, buf)
		srcFile.Close()
		dstFile.Close()
		retInfo[fname] = err
		if err != nil {
			continue
		}
	}
	return
}

func (c *XSFtp) ConnectAndDownload() (files map[string]string, err error) {
	err = c.Connect()
	if err != nil {
		return nil, err
	}
	defer c.Logout()
	files, err = c.DownloadFiles(c.NameList())
	if err != nil {
		return nil, err
	}
	return files, nil
}
