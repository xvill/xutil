package xutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	ftp4go "github.com/shenshouer/ftp4go"
)

type XFtp struct {
	Addr            string
	User            string
	Pwd             string
	PASV            string
	FilePattern     []string
	LocalFilePrefix string
	Conn            *ftp4go.FTP

	FtpFilePats []struct {
		Filepatterns []string `json:"filepatterns"`
		Info         []string `json:"info"`
	}
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

func (c XFtp) MKdir(path string) {
	xdir, xfile := filepath.Split(path)
	fname := filepath.Join(xdir, xfile)
	xdirFiles, _ := c.Conn.Nlst(xdir)
	for _, v := range xdirFiles {
		if v == fname {
			return
		}
	}
	_, err := c.Conn.Mkd(path)
	if err != nil {
		c.MKdir(xdir)
		c.MKdir(fname)
	}
	return
}

func (c XFtp) Size(path string) int64 {
	size, err := c.Conn.Size(path)
	if err != nil {
		return 0
	}
	return int64(size)
}

func (c XFtp) NameList() (ftpfiles []string) {
	return c.FileList("NLST")
}
func (c XFtp) InfoList() (ftpfiles []string) {
	return c.FileList("LIST")
}

func (c XFtp) FileList(CMD string) (ftpfiles []string) {
	for _, fpattern := range c.FilePattern {
		nowfiles := []string{}
		if strings.Contains(filepath.Dir(fpattern), "*") {
			fdirs := []string{}
			fmaps := make(map[string][]string, 0)
			fpaths := strings.Split(fpattern, "/")
			for i, fpath := range fpaths {
				if strings.Contains(fpath, "*") {
					fp := strings.Join(fpaths[0:i+1], "/")
					fmaps[fp] = []string{}
					fdirs = append(fdirs, fp)
				}
			}
			if len(fdirs) == 0 {
				continue
			}

			files, err := c.Conn.Nlst(fdirs[0])
			if err != nil {
				continue
			}
			fmaps[fdirs[0]] = files

			for i, nowpath := range fdirs[1 : len(fdirs)-1] {
				lastpath := fdirs[i]
				xfdir := strings.ReplaceAll(nowpath, lastpath, "")
				for _, fpath := range fmaps[lastpath] {
					xfiles, _ := c.Conn.Nlst(filepath.Join(fpath, xfdir))
					fmaps[nowpath] = append(fmaps[nowpath], xfiles...)
				}
			}
			lastpath := fdirs[len(fdirs)-1]
			lastPathDir := filepath.Dir(lastpath)
			lastPathBase := filepath.Base(lastpath)
			for _, fpath := range fmaps[lastPathDir] {
				nowfiles = append(nowfiles, filepath.Join(fpath, lastPathBase))
			}
		} else {
			nowfiles = []string{fpattern}
		}

		if CMD == "NLST" {
			for _, v := range nowfiles {
				xfiles, _ := c.Conn.Nlst(v)
				xdir := filepath.Dir(v)
				for i := range xfiles {
					xfiles[i] = filepath.Join(xdir, filepath.Base(xfiles[i]))
				}
				ftpfiles = append(ftpfiles, xfiles...)
			}
		} else if CMD == "LIST" {
			for _, v := range nowfiles {
				xdir := filepath.Dir(v)
				xfiles, _ := c.Conn.Dir(v)
				for _, e := range xfiles {
					ls := ParsrLS(e)
					ftpfiles = append(ftpfiles, xdir+"/"+strings.Join(ls, ","))
				}
			}
		}
	}
	return ftpfiles
}

func (c XFtp) DownloadFiles(files []string) (dat map[string]string, err error) {
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
		blockSize := 819200
		f, err := os.OpenFile(localpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		err = c.Conn.GetBytes(ftp4go.RETR_FTP_CMD, f, blockSize, file)
		// err = c.Conn.DownloadFile(file, localpath, false)
		if err != nil {
			fmt.Println(err)
			f.Close()
			os.Remove(localpath) //
			continue
		}
		f.Close()
		dat[file] = localpath
	}
	fmt.Println("DownloadFiles end")
	return dat, nil
}

func (c XFtp) DownloadFilesMap(files map[string]string) (dat map[string]string, err error) {
	dat = make(map[string]string, 0)
	if len(files) == 0 {
		return
	}
	fmt.Println("DownloadFiles begin")
	for ftpfile, localfile := range files {
		fmt.Println("DownloadFile " + ftpfile + " to " + localfile)
		blockSize := 819200
		f, err := os.OpenFile(localfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		err = c.Conn.GetBytes(ftp4go.RETR_FTP_CMD, f, blockSize, ftpfile)
		// err = c.Conn.DownloadFile(file, localpath, false)
		if err != nil {
			return dat, err
		}
		f.Close()
		dat[ftpfile] = localfile
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
	files, err = c.DownloadFiles(c.NameList())
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *XFtp) UploadFiles(files map[string]string, useLineMode bool) (retInfo map[string]error) {
	retInfo = make(map[string]error, 0)
	for localname, remotename := range files {
		retInfo[localname] = c.Conn.UploadFile(remotename, localname, useLineMode, nil)
	}
	return
}

func FtpDownload(ftptype, addr, user, pwd, pasv, localfileprefix string, ftpfiles []string) (rawftpfiles map[string]string, err error) {
	var xftp XFtp
	var xsftp XSFtp

	rawftpfiles = make(map[string]string, 0)

	switch ftptype {
	case "FTP":
		xftp = XFtp{Addr: addr, User: user, Pwd: pwd, PASV: pasv, LocalFilePrefix: localfileprefix}
		err = xftp.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer xftp.Logout()
		rawftpfiles, err = xftp.DownloadFiles(ftpfiles)
	case "SFTP":
		xsftp = XSFtp{Addr: addr, User: user, Pwd: pwd, LocalFilePrefix: localfileprefix}
		err = xsftp.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer xsftp.Logout()
		rawftpfiles, err = xsftp.DownloadFiles(ftpfiles)
	}
	return
}

//GetFTPFiles 获取 FTP/SFTP 匹配的文件
func GetFTPFiles(ftptype, addr, user, pwd, pasv, localfileprefix string, pattern []string, expectfiles []string) (files map[string]string, err error) {

	ftpfiles := make([]string, 0)
	var xftp XFtp
	var xsftp XSFtp
	files = make(map[string]string, 0)

	switch ftptype {
	case "FTP":
		xftp = XFtp{Addr: addr,
			User:            user,
			Pwd:             pwd,
			PASV:            pasv,
			FilePattern:     pattern,
			LocalFilePrefix: localfileprefix}

		err = xftp.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer xftp.Logout()
		ftpfiles = xftp.NameList()
	case "SFTP":
		xsftp = XSFtp{Addr: addr,
			User:            user,
			Pwd:             pwd,
			FilePattern:     pattern,
			LocalFilePrefix: localfileprefix}

		err = xsftp.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer xsftp.Logout()
		ftpfiles = xsftp.NameList()
	}
	// ------------------------------------------------------------------------------
	for i := range ftpfiles {
		ftpfiles[i] = fmt.Sprintf("[%s]%s", addr, ftpfiles[i])
	}

	getftpfiles := StringsMinus(ftpfiles, expectfiles) // 要下载的文件 = FTP文件名 - 已入库的文件
	if len(getftpfiles) == 0 {                         //没有可下载的文件
		return
	}

	getftpfiles = StringsUniq(getftpfiles)
	for i := range getftpfiles {
		getftpfiles[i] = strings.TrimPrefix(getftpfiles[i], "["+addr+"]")
	}
	// ------------------------------------------------------------------------------
	xfiles := make(map[string]string, 0)
	switch ftptype {
	case "FTP":
		xfiles, err = xftp.DownloadFiles(getftpfiles)
	case "SFTP":
		xfiles, err = xsftp.DownloadFiles(getftpfiles)
	}

	for ftpfile, localfile := range xfiles {
		files[fmt.Sprintf("[%s]%s", addr, ftpfile)] = localfile
	}
	return
}

func ParsrLS(s string) (fileInfo []string) {
	//"drwxrwxr-x    5 577      554          4096 May 10  2019 pm",
	//"-rwxrwxrwx    1 501      510       5102081 Oct 09 17:23 pmchk.out",
	// "06-29-22  01:31PM              1383076 NK_Kpi4G_Plmn_202206291315.zip"  --windowsFTP
	var fileName, fileType, fileSize, fileTime string
	arr := strings.Fields(s)
	if len(arr) == 9 {

		fileName, fileSize = arr[8], arr[4]
		fileTime = strings.Join(arr[5:8], " ")
		if strings.Contains(arr[7], ":") {
			t, err := time.Parse("Jan 2 15:04", fileTime)
			if err == nil {
				fileTime = t.AddDate(time.Now().Year(), 0, 0).Format("2006-01-02 15:04")
			}
		} else {
			t, err := time.Parse("Jan 2 2006", fileTime)
			if err == nil {
				fileTime = t.Format("2006-01-02 15:04")
			}
		}
		switch arr[0][0] {
		case '-':
			fileType = "file"
		case 'd':
			fileType = "folder"
		case 'l':
			fileType = "link"
		default:
			fileType = ""
		}
	} else if len(arr) == 4 {
		t, err := time.Parse("01-02-06 15:04PM", strings.Join(arr[0:2], " "))
		if err == nil {
			fileTime = t.Format("2006-01-02 15:04")
		}
		fileType, fileName, fileSize = "file", arr[3], arr[2]
	} else {
		return
	}
	return []string{fileName, fileType, fileSize, fileTime}
}

// FtpFileList 多文件路径匹配 Info->id:文件时间正则 ,结果-> id,文件匹配时间,文件名,文件类型,文件大小,服务器文件时间
func (c *XFtp) FtpFileList() (loadftpFiles []string) {
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
						loadftpFiles = append(loadftpFiles, fmt.Sprintf("%s,%s,%s", id[0], "", ftpfile))
					} else {
						loadftpFiles = append(loadftpFiles, fmt.Sprintf("%s,%s,%s", id[0], dtime.Format("2006-01-02 15:04:05"), ftpfile))
					}
				}
			}
		}
	}
	return
}
