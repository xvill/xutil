package xutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

// ToFixed 浮点数保留
func ToFixed(f float64, n int) float64 {
	shift := math.Pow(10, float64(n))
	fv := 0.0000000001 + f //对浮点数产生.xxx999999999 计算不准进行处理
	return math.Floor(fv*shift+.5) / shift
}

//===============================================================================
/*
	WGS-84：国际标准,GPS坐标（GoogleEarth、或GPS模块）
	GCJ-02：火星坐标,中国坐标偏移标准(GoogleMap、高德、腾讯)
	BD-09：百度坐标偏移标准(BaiduMap)

	http://www.gpsspg.com/maps.htm                      坐标拾取
	http://lbs.amap.com/console/show/picker             高德坐标拾取
	http://api.map.baidu.com/lbsapi/getpoint/index.html 百度坐标拾取
	https://github.com/wandergis/coordtransform         javascript版本
	https://github.com/wandergis/coordTransform_py      python版本
	https://github.com/FreeGIS/postgis_LayerTransform   postgre版本
*/
const (
	_pi  = 3.14159265358979324    //圆周率
	_a   = 6378245.0              //卫星椭球坐标投影到平面地图坐标系的投影因子
	_ee  = 0.00669342162296594323 //椭球的偏心率
	_xpi = _pi * 3000.0 / 180.0   //圆周率转换量
)

func _transformlon(lon, lat float64) float64 {
	dlon := 300 + lon + 2*lat + 0.1*lon*lon + 0.1*lon*lat + 0.1*math.Sqrt(math.Abs(lon)) +
		(20*math.Sin(6*lon*_pi)+20*math.Sin(2*lon*_pi))*2/3 +
		(20*math.Sin(lon*_pi)+40*math.Sin(lon/3*_pi))*2/3 +
		(150*math.Sin(lon/12*_pi)+300*math.Sin(lon/30*_pi))*2/3
	return dlon
}

func _transformlat(lon, lat float64) float64 {
	dLat := -100 + 2*lon + 3*lat + 0.2*lat*lat + 0.1*lon*lat + 0.2*math.Sqrt(math.Abs(lon)) +
		(20*math.Sin(6*lon*_pi)+20*math.Sin(2*lon*_pi))*2/3 +
		(20*math.Sin(lat*_pi)+40*math.Sin(lat/3*_pi))*2/3 +
		(160*math.Sin(lat/12*_pi)+320*math.Sin(lat*_pi/30))*2/3
	return dLat
}

func _offset(lon, lat float64) (float64, float64) {
	dlat := _transformlat(lon-105.0, lat-35.0)
	dlon := _transformlon(lon-105.0, lat-35.0)
	radLat := lat / 180.0 * _pi
	magic := math.Sin(radLat)
	magic = 1 - _ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((_a * (1 - _ee)) / (magic * sqrtMagic) * _pi)
	dlon = (dlon * 180.0) / (_a / sqrtMagic * math.Cos(radLat) * _pi)
	return dlon, dlat
}

// Wgs2gcj WGS坐标系 ----> GCJ坐标系
func Wgs2gcj(lon, lat float64) (float64, float64) {
	dlon, dlat := _offset(lon, lat)
	mglat := lat + dlat
	mglon := lon + dlon
	return ToFixed(mglon, 7), ToFixed(mglat, 7)
}

// Gcj2Wgs  GCJ坐标系 ----> WGS坐标系
func Gcj2Wgs(lon, lat float64) (float64, float64) {
	dlon, dlat := _offset(lon, lat)
	mglat := lat - dlat
	mglon := lon - dlon
	return ToFixed(mglon, 7), ToFixed(mglat, 7)
}

// Gcj2bd  火星(GCJ-02)坐标系 ----> 百度(BD-09)坐标系
func Gcj2bd(lon, lat float64) (float64, float64) {
	x, y := lon, lat
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*_xpi)
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*_xpi)
	bdLon := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return ToFixed(bdLon, 7), ToFixed(bdLat, 7)
}

// Bd2gcj  百度(BD-09)坐标系 ----> 火星(GCJ-02)坐标系
func Bd2gcj(lon, lat float64) (float64, float64) {
	x, y := lon-0.0065, lat-0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*_xpi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*_xpi)
	ggLon := z * math.Cos(theta)
	ggLat := z * math.Sin(theta)
	return ToFixed(ggLon, 7), ToFixed(ggLat, 7)
}

// Wgs2bd WGS坐标系 ----> 百度坐标系
func Wgs2bd(lon, lat float64) (float64, float64) {
	x, y := Wgs2gcj(lon, lat)
	lng, lat := Gcj2bd(x, y)
	return lng, lat
}

// Bd2Wgs 百度坐标系 ----> WGS坐标系
func Bd2Wgs(lon, lat float64) (float64, float64) {
	x, y := Bd2gcj(lon, lat)
	lng, lat := Gcj2Wgs(x, y)
	return lng, lat
}

//===============================================================================

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y string) (p Point, err error) {
	m, err := strconv.ParseFloat(x, 64)
	if err != nil {
		return p, err
	}
	n, err := strconv.ParseFloat(y, 64)
	if err != nil {
		return p, err
	}
	return Point{X: m, Y: n}, nil
}

func (p *Point) ReverseXY() {
	p.X, p.Y = p.Y, p.X
}

func (p *Point) Wgs2gcj() {
	p.X, p.Y = Wgs2gcj(p.X, p.Y)
}

func (p *Point) Wgs2bd() {
	p.X, p.Y = Wgs2bd(p.X, p.Y)
}

func (p *Point) Gcj2bd() {
	p.X, p.Y = Gcj2bd(p.X, p.Y)
}

func (p Point) String() string {
	return fmt.Sprintf("%g,%g", p.X, p.Y)
}

//===============================================================================

/***
http://www.movable-type.co.uk/scripts/latlong.html
http://en.wikipedia.org/wiki/Haversine_formula	Haversine
http://en.wikipedia.org/wiki/Spherical_law_of_cosines	Law of Cosines
http://en.wikipedia.org/wiki/Vincenty's_formulae	Vincenty
Speed: Law of Cosines > Haversine > Vincenty
***/

func Radians(r float64) float64 {
	return r * math.Pi / 180.0
}

func Degrees(d float64) float64 {
	return d * 180.0 / math.Pi
}

func MidPoint(lon1, lat1, lon2, lat2 float64) (float64, float64) {
	λ1 := Radians(lon1)
	λ2 := Radians(lon2)
	φ1 := Radians(lat1)
	φ2 := Radians(lat2)
	Bx := math.Cos(φ2) * math.Cos(λ2-λ1)
	By := math.Cos(φ2) * math.Sin(λ2-λ1)
	φ3 := math.Atan2(math.Sin(φ1)+math.Sin(φ2), math.Sqrt((math.Cos(φ1)+Bx)*(math.Cos(φ1)+Bx)+By*By))
	λ3 := λ1 + math.Atan2(By, math.Cos(φ1)+Bx)

	return Degrees(λ3), Degrees(φ3)
}

// Azimuth  bearing between the two GPS points
func Azimuth(lon1, lat1, lon2, lat2 float64) float64 {
	rad := math.Pi / 180.0
	lat1 = lat1 * rad
	lat2 = lat2 * rad
	lon1 = lon1 * rad
	lon2 = lon2 * rad
	dLon := lon2 - lon1
	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	a := math.Atan2(y, x)
	if dLon < 0 {
		a = a + 2*math.Pi
	}
	return a * 180 / math.Pi
}

// Distance (in meter) Spherical_law_of_cosines
func Distance(lon1, lat1, lon2, lat2 float64) float64 {
	r, rad := 6371000.0, math.Pi/180.0
	lat1 = lat1 * rad
	lat2 = lat2 * rad
	lon1 = lon1 * rad
	lon2 = lon2 * rad
	theta := lon2 - lon1
	return r * math.Acos(math.Sin(lat1)*math.Sin(lat2)+
		math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
}

// DistanceHaversine (in meter) Haversine_formula
func DistanceHaversine(lon1, lat1, lon2, lat2 float64) float64 {
	r, rad := 6371000.0, math.Pi/180.0
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	lat1 = lat1 * rad
	lat2 = lat2 * rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

//DistancePoint   Destination point given distance and bearing from start point
func DistancePoint(lon1, lat1, dist, azimuth float64) (float64, float64) {
	/**
	Destination point given distance and bearing from start point
	http://www.movable-type.co.uk/scripts/latlong.html
	Formula:	φ2 = asin( sin φ1 ⋅ cos δ + cos φ1 ⋅ sin δ ⋅ cos θ )
				λ2 = λ1 + atan2( sin θ ⋅ sin δ ⋅ cos φ1, cos δ − sin φ1 ⋅ sin φ2 )
	where	φ is latitude, λ is longitude,
			θ is the bearing (clockwise from north),
			δ is the angular distance d/R;
			d being the distance travelled,R the earth’s radius
	*/

	φ1 := Radians(lat1)
	λ1 := Radians(lon1)
	θ := Radians(azimuth)
	δ := dist / _a // normalize linear distance to radian angle

	φ2 := math.Asin(math.Sin(φ1)*math.Cos(δ) + math.Cos(φ1)*math.Sin(δ)*math.Cos(θ))
	λ2 := λ1 + math.Atan2(math.Sin(θ)*math.Sin(δ)*math.Cos(φ1), math.Cos(δ)-math.Sin(φ1)*math.Sin(φ2))

	if λ2 < 0 {
		λ2 = λ2 + 2*math.Pi
	}
	// λ2_harmonised := (λ2+3.0*math.Pi)%(2.0*math.Pi) - math.Pi // normalise to −180..+180°
	// return Degrees(λ2_harmonised), Degrees(φ2)
	return Degrees(λ2), Degrees(φ2)
}

// AmapGeocode 高德解析地址为经纬度
func AmapGeocode(ak, address string) (poi map[string]string, err error) {
	url := fmt.Sprintf("http://restapi.amap.com/v3/geocode/geo?key=%s&address=%s", ak, address)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	info := jsoniter.Get(body, "info").ToString()
	if info != "OK" {
		return poi, errors.New(info)
	}
	poi = make(map[string]string, 6)
	poi["formatted_address"] = jsoniter.Get(body, "geocodes", 0, "formatted_address").ToString()
	poi["province"] = jsoniter.Get(body, "geocodes", 0, "province").ToString()
	poi["citycode"] = jsoniter.Get(body, "geocodes", 0, "citycode").ToString()
	poi["city"] = jsoniter.Get(body, "geocodes", 0, "city").ToString()
	poi["district"] = jsoniter.Get(body, "geocodes", 0, "district").ToString()
	poi["location"] = jsoniter.Get(body, "geocodes", 0, "location").ToString()
	return poi, nil
}

// BdmapGeocode 百度解析地址为经纬度
func BdmapGeocode(ak, address string) (poi map[string]string, err error) {
	url := fmt.Sprintf("http://api.map.baidu.com/geocoder/v2/?output=json&ak=%s&address=%s", ak, address)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	info := jsoniter.Get(body, "status").ToString()
	if info != "0" {
		msg := jsoniter.Get(body, "message").ToString()
		return poi, errors.New(msg)
	}
	poi = make(map[string]string, 6)
	poi["lng"] = jsoniter.Get(body, "result", "location", "lng").ToString()
	poi["lat"] = jsoniter.Get(body, "result", "location", "lat").ToString()
	return poi, nil
}

//===============================================================================

// func demo() {
// 	lat, lon := 31.2355502882, 121.5012091398  //上海中心大厦gps
// 	fmt.Println(Wgs2bd(lon, lat)) //31.239186 121.512245
// }

//===============================================================================
