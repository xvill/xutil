# xtools

## install
go get -u github.com/xvill/xtools

## function

``` file.go
// CsvWriteALL 生成CSV
func CsvWriteALL(data [][]string, wfile string, comma rune) error 

// Sqlldr 执行成功返回入库记录数,失败则保留log和data到baddir
func Sqlldr(timeflag, userid, data, control, baddir string) (rows, badrows int, err error) 

// IsFileExist 文件是否存在
func IsFileExist(path string) (isExist, isDir bool, err error) 
```


``` geo.go
// FromWKT 解析WKT为Geo
func FromWKT(wkt string) (Geo, error)

// ToWKT 生成WKT
func (g Geo) ToWKT() (wkt string) 

// GeoJSON 生成GeoJSON
func (g Geo) GeoJSON() (s string, err error) 

// ReserveLngLat 转换Lat,Lng 位置
func (g Geo) ReserveLngLat() 

// Wgs2gcj 经纬度坐标系转换 wgs-> gcj
func (g Geo) Wgs2gcj()

// Gcj2bd 经纬度坐标系转换 gcj->BD09
func (g Geo) Gcj2bd() 

// Wgs2bd 经纬度坐标系转换 wgs->BD09
func (g Geo) Wgs2bd() 

// Box 方框边界 minx, miny, maxx, maxy 
func (g Geo) Box() []float64 
```


``` gis.go
// Wgs2gcj WGS坐标系 ----> GCJ坐标系
func Wgs2gcj(lat, lon float64) (float64, float64)

// Gcj2bd  火星(GCJ-02)坐标系 ----> 百度(BD-09)坐标系
func Gcj2bd(lat, lon float64) (float64, float64)

// Bd2gcj  百度(BD-09)坐标系 ----> 火星(GCJ-02)坐标系
func Bd2gcj(lat, lon float64) (float64, float64) 

// Wgs2bd WGS坐标系 ----> 百度坐标系
func Wgs2bd(lat, lon float64) (float64, float64) 

// EarthDistance 两经纬度距离
func EarthDistance(lat1, lng1, lat2, lng2 float64) float64

```
