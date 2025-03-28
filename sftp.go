package xutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
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

	FtpFilePats []struct {
		Filepatterns []string `json:"filepatterns"`
		Info         []string `json:"info"`
	}
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

func (c *XSFtp) InfoList() (ftpfiles []string) {
	for _, fpattern := range c.FilePattern {
		fdirs := []string{filepath.Dir(fpattern)}
		if strings.Contains(filepath.Dir(fpattern), "*") {
			fdirs, _ = c.SFTP.Glob(filepath.Dir(fpattern))
		}
		for _, fdir := range fdirs {
			files, _ := c.SFTP.ReadDir(fdir)
			for _, fl := range files {
				fname := path.Join(fdir, fl.Name())
				if matched, _ := path.Match(fpattern, fname); matched {
					ftpfiles = append(ftpfiles, fmt.Sprintf("%s,file,%d,%s", fname, fl.Size(), fl.ModTime().Local().Format("2006-01-02 15:04:05")))
				}
			}
		}
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

func (c *XSFtp) DownloadFilesMap(files map[string]string) (dat map[string]string, err error) {
	dat = make(map[string]string, 0)
	if len(files) == 0 {
		return
	}
	fmt.Println("DownloadFiles begin")
	for ftpfile, localfile := range files {
		fmt.Println("DownloadFile " + ftpfile + " to " + localfile)

		srcFile, err := c.SFTP.Open(ftpfile)
		if err != nil {
			fmt.Println("Open", err)
			return dat, err
		}

		raw, err := ioutil.ReadAll(srcFile)
		if err != nil {
			fmt.Println("ReadAll", err)
			return dat, err
		}

		err = ioutil.WriteFile(localfile, raw, 0666)
		if err != nil {
			fmt.Println("WriteFile", err)
			return dat, err
		}
		srcFile.Close()
		dat[ftpfile] = localfile
	}
	fmt.Println("DownloadFiles end")
	return dat, nil
}

func (c *XSFtp) UploadFiles(files map[string]string) (retInfo map[string]error) {
	retInfo = make(map[string]error, 0)
	for localname, remotename := range files {
		tmpname := remotename + ".tmp"
		srcFile, err := os.Open(localname)
		retInfo[localname] = err
		if err != nil {
			continue
		}
		dstFile, err := c.SFTP.Create(tmpname)
		retInfo[localname] = err
		if err != nil {
			continue
		}
		buf := make([]byte, 1024)
		_, err = io.CopyBuffer(dstFile, srcFile, buf)
		srcFile.Close()
		dstFile.Close()
		retInfo[localname] = err
		if err != nil {
			continue
		}
		retInfo[localname] = c.SFTP.Rename(tmpname, remotename)

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

// FtpFileList 多文件路径匹配 Info->id:文件时间正则 ,结果-> id,文件匹配时间,文件名,文件类型,文件大小,服务器文件时间
func (c *XSFtp) FtpFileList() (loadftpFiles []string) {
	err := c.Connect()
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Logout()
	for _, rfile := range c.FtpFilePats {
		c.FilePattern = rfile.Filepatterns
		rawftpfiles := c.InfoList()

		for _, ids := range rfile.Info {
			id := strings.Split(ids, ":")
			if len(id) != 2 {
				continue
			}
			rege := regexp.MustCompile(id[1])
			for _, ftpfile := range rawftpfiles {
				reg := rege.FindStringSubmatch(strings.Split(ftpfile, ",")[0])
				if len(reg) == 2 {
					dtime, err := TimeParse(reg[1])
					if err != nil {
						log.Println("TimeParse", ftpfile, err)
					}
					loadftpFiles = append(loadftpFiles, fmt.Sprintf("%s,%s,%s", id[0], dtime.Format("2006-01-02 15:04:05"), ftpfile))
				}
			}
		}
	}
	return
}
