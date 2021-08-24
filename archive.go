package xutil

import (
	"archive/zip"
	"compress/gzip"
	"io/ioutil"
	"os"
)

//ZipDeCompress ZIP解压
func ZipDeCompress(fname string) (retRaw map[string][]byte, err error) {
	retRaw = make(map[string][]byte)
	cf, err := zip.OpenReader(fname)
	if err != nil {
		return
	}
	defer cf.Close()
	for _, f := range cf.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		raw, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		retRaw[f.Name] = raw
	}
	return
}

func Unzip(fname, oname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	gzfile, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzfile.Close()

	raw, err := ioutil.ReadAll(gzfile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(oname, raw, 0600)
	return err
}
