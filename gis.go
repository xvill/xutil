package tools

import "math"

/*
	WGS-84：国际标准,GPS坐标（GoogleEarth、或GPS模块）
	GCJ-02：火星坐标,中国坐标偏移标准(GoogleMap、高德、腾讯)
	BD-09：百度坐标偏移标准(BaiduMap)

	http://www.gpsspg.com/maps.htm 坐标拾取
	http://api.map.baidu.com/lbsapi/getpoint/index.html 百度坐标拾取
*/
const (
	_pi  = 3.14159265358979324    //圆周率
	_a   = 6378245.0              //卫星椭球坐标投影到平面地图坐标系的投影因子
	_ee  = 0.00669342162296594323 //椭球的偏心率
	_xpi = _pi * 3000.0 / 180.0   //圆周率转换量
)

// Wgs2gcj WGS坐标系 ----> GCJ坐标系
func Wgs2gcj(lat, lon float64) (float64, float64) {
	x, y := lon-105.0, lat-35.0
	dLon := 300 + x + 2*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x)) +
		(20*math.Sin(6*x*_pi)+20*math.Sin(2*x*_pi))*2/3 +
		(20*math.Sin(x*_pi)+40*math.Sin(x/3*_pi))*2/3 +
		(150*math.Sin(x/12*_pi)+300*math.Sin(x/30*_pi))*2/3
	dLat := -100 + 2*x + 3*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x)) +
		(20*math.Sin(6*x*_pi)+20*math.Sin(2*x*_pi))*2/3 +
		(20*math.Sin(y*_pi)+40*math.Sin(y/3*_pi))*2/3 +
		(160*math.Sin(y/12*_pi)+320*math.Sin(y*_pi/30))*2/3

	radLat := lat / 180.0 * _pi
	magic := math.Sin(radLat)
	magic = 1 - _ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((_a * (1 - _ee)) / (magic * sqrtMagic) * _pi)
	dLon = (dLon * 180.0) / (_a / sqrtMagic * math.Cos(radLat) * _pi)
	mgLat := lat + dLat
	mgLon := lon + dLon
	return mgLat, mgLon
}

// Gcj2bd  火星(GCJ-02)坐标系 ----> 百度(BD-09)坐标系
func Gcj2bd(lat, lon float64) (float64, float64) {
	x, y := lon, lat
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*_xpi)
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*_xpi)
	bdLon := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return bdLat, bdLon
}

// Bd2gcj  百度(BD-09)坐标系 ----> 火星(GCJ-02)坐标系
func Bd2gcj(lat, lon float64) (float64, float64) {
	x, y := lon-0.0065, lat-0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*_xpi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*_xpi)
	ggLon := z * math.Cos(theta)
	ggLat := z * math.Sin(theta)
	return ggLat, ggLon
}

// Wgs2bd WGS坐标系 ----> 百度坐标系
func Wgs2bd(lat, lon float64) (float64, float64) {
	x, y := Wgs2gcj(lat, lon)
	lat, lng := Gcj2bd(x, y)
	return math.Trunc(lat*1000000) / 1000000, math.Trunc(lng*1000000) / 1000000
}

// EarthDistance 两经纬度距离
func EarthDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const EarthRadius = 6378.137 // 地球半径 km
	const Rad = math.Pi / 180.0  // 计算弧度
	lat1, lng1 = lat1*Rad, lng1*Rad
	lat2, lng2 = lat2*Rad, lng2*Rad
	theta := lng2 - lng1
	return EarthRadius *
		math.Acos(math.Sin(lat1)*math.Sin(lat2)+
			math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
}

// func main() {
// 	lat, lon := 31.2355502882, 121.5012091398  //上海中心大厦gps
// 	fmt.Println(Wgs2bd(lat, lon)) //31.239186 121.512245
// }
