# xutil

## install
go get -u github.com/xvill/xutil


## usage
```go
package main

import (
	"fmt"

	"github.com/xvill/xutil"
)

func main() {
	g, _ := xtools.FromWKT("POINT(121.44528145 30.96964209)")
	g.Wgs2gcj()
        g.ReserveLngLat()
	fmt.Println(g)

	wktstr := []string{
		"POINT(1 2)",
		"LINESTRING(3 4,10 50,20 25)",
		"POLYGON((1 1,5 1,5 5,1 5,1 1),(2 2, 3 2, 3 3, 2 3,2 2))",
		"MULTIPOINT(3.5 5.6,4.8 10.5)",
		"MULTILINESTRING((3 4,10 50,20 25),(-5 -8,-10 -8,-15 -4))",
		"MULTIPOLYGON(((1 1,5 1,5 5,1 5,1 1),(2 2, 3 2, 3 3, 2 3,2 2)),((3 3,6 2,6 4,3 3)))",
	}
	for _, s := range wktstr {
		g, _ := xtools.FromWKT(s)
		fmt.Println(g.GeoJSON())
	}
}
```
## function

```go
func CsvWriteALL(data [][]string, wfile string, comma rune) error {} // 生成CSV
func Sqlldr(timeflag, userid, data, control, baddir string)(rows, badrows int, err error)  {}    // 执行成功返回入库记录数,失败则保留log和data到baddir
func IsFileExist(path string) (isExist, isDir bool, err error) {}    // 文件是否存在
 
func FromWKT(wkt string) (Geo, error){}  // 解析WKT为Geo
func (g Geo) ToWKT() (wkt string) {} // 生成WKT
func (g Geo) GeoJSON() (s string, err error) {}  // 生成GeoJSON
func (g Geo) ReserveLngLat() {}  // 转换Lat,Lng 位置
func (g Geo) Wgs2gcj(){} // 经纬度坐标系转换 wgs-> gcj
func (g Geo) Gcj2bd() {} // 经纬度坐标系转换 gcj->BD09
func (g Geo) Wgs2bd() {} // 经纬度坐标系转换 wgs->BD09
func (g Geo) Box() []float64 {}  // 方框边界 minx, miny, maxx, maxy 
 
func Wgs2gcj(lon, lat float64) (float64, float64){}  // WGS坐标系 ----> GCJ坐标系
func Gcj2bd(lon, lat float64) (float64, float64){}   //  火星(GCJ-02)坐标系 ----> 百度(BD-09)坐标系
func Bd2gcj(lon, lat float64) (float64, float64) {}  //  百度(BD-09)坐标系 ----> 火星(GCJ-02)坐标系
func Wgs2bd(lon, lat float64) (float64, float64) {}  // WGS坐标系 ----> 百度坐标系

func Distance(lon1, lat1, lon2, lat2 float64) float64 {} // 两经纬度距离
func DistanceHaversine(lon1, lat1, lon2, lat2 float64) float64 {} // 两经纬度距离
func ToFixed(f float64, n int) float64 {}    // 浮点数保留

身份证信息
行政区信息获取
http://www.mca.gov.cn/article/sj/xzqh/2018/  2018年行政区划代码
http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/    统计用区划和城乡划分代码

func IDsumY(id string) string {} 	// IDsumY 计算身份证的第十八位校验码
func ID15to18(id string) string {} 	// ID15to18 将15位身份证转换为18位的
func IDisValid(id string) bool {} 	// IDisValid 校验身份证第18位是否正确
func IDisPattern(id string) bool {} 	// IDisPattern 二代身份证正则表达式
func NewIDCard(id string) (c IDCard, err error) {} 	// NewIDCard  获取身份证信息

```
