# xtools

## install
go get -u github.com/xvill/xtools

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
 
func Wgs2gcj(lat, lon float64) (float64, float64){}  // WGS坐标系 ----> GCJ坐标系
func Gcj2bd(lat, lon float64) (float64, float64){}   //  火星(GCJ-02)坐标系 ----> 百度(BD-09)坐标系
func Bd2gcj(lat, lon float64) (float64, float64) {}  //  百度(BD-09)坐标系 ----> 火星(GCJ-02)坐标系
func Wgs2bd(lat, lon float64) (float64, float64) {}  // WGS坐标系 ----> 百度坐标系
func EarthDistance(lat1, lng1, lat2, lng2 float64) float64{} // 两经纬度距离
func ToFixed(f float64, n int) float64 {}    // 浮点数保留

```
