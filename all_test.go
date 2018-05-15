package xutil

import (
	"fmt"
	"testing"

	"github.com/xvill/xutil"
)

func Test_Wgs2bd(t *testing.T) {
	lat, lon := 31.2355502882, 121.5012091398
	fmt.Println(xutil.Wgs2bd(lat, lon))
	// 121.486245,31.3838164	121.47521,31.37982
}
