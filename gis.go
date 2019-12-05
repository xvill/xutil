package xutil

import (
	"fmt"
	"math"

	"github.com/gonum/floats"
)

//===============================================================================

//Round Round
func Round(x float64, prec int) float64 {
	return floats.Round(x, prec)
}

//PointRound6 PointRound6
func PointRound6(x, y float64) (float64, float64) {
	return floats.Round(x, 6), floats.Round(y, 6)
}

//PointRound7 PointRound7
func PointRound7(x, y float64) (float64, float64) {
	return floats.Round(x, 7), floats.Round(y, 7)
}

//PointRound8 PointRound8
func PointRound8(x, y float64) (float64, float64) {
	return floats.Round(x, 8), floats.Round(y, 8)
}

//===============================================================================
/*
	WGS-84: 国际标准,GPS坐标（GoogleEarth、或GPS模块）
	GCJ-02: 火星坐标,中国坐标偏移标准(GoogleMap、高德、腾讯)
	BD-09: 百度坐标偏移标准(BaiduMap)

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

//百度 墨卡托坐标与经纬度坐标转换
var (
	_mcband = []float64{12890594.86, 8362377.87, 5591021, 3481989.83, 1678043.12, 0.0}
	_llband = []float64{75, 60, 45, 30, 15, 0}
	_mc2ll  = [][]float64{
		{1.410526172116255e-8, 0.00000898305509648872, -1.9939833816331, 200.9824383106796, -187.2403703815547, 91.6087516669843, -23.38765649603339, 2.57121317296198, -0.03801003308653, 17337981.2},
		{-7.435856389565537e-9, 0.000008983055097726239, -0.78625201886289, 96.32687599759846, -1.85204757529826, -59.36935905485877, 47.40033549296737, -16.50741931063887, 2.28786674699375, 10260144.86}, {-3.030883460898826e-8, 0.00000898305509983578, 0.30071316287616, 59.74293618442277, 7.357984074871, -25.38371002664745, 13.45380521110908, -3.29883767235584, 0.32710905363475, 6856817.37},
		{-1.981981304930552e-8, 0.000008983055099779535, 0.03278182852591, 40.31678527705744, 0.65659298677277, -4.44255534477492, 0.85341911805263, 0.12923347998204, -0.04625736007561, 4482777.06},
		{3.09191371068437e-9, 0.000008983055096812155, 0.00006995724062, 23.10934304144901, -0.00023663490511, -0.6321817810242, -0.00663494467273, 0.03430082397953, -0.00466043876332, 2555164.4},
		{2.890871144776878e-9, 0.000008983055095805407, -3.068298e-8, 7.47137025468032, -0.00000353937994, -0.02145144861037, -0.00001234426596, 0.00010322952773, -0.00000323890364, 826088.5}}
	_ll2mc = [][]float64{{-0.0015702102444, 111320.7020616939, 1704480524535203, -10338987376042340, 26112667856603880, -35149669176653700, 26595700718403920, -10725012454188240, 1800819912950474, 82.5},
		{0.0008277824516172526, 111320.7020463578, 647795574.6671607, -4082003173.641316, 10774905663.51142, -15171875531.51559, 12053065338.62167, -5124939663.577472, 913311935.9512032, 67.5},
		{0.00337398766765, 111320.7020202162, 4481351.045890365, -23393751.19931662, 79682215.47186455, -115964993.2797253, 97236711.15602145, -43661946.33752821, 8477230.501135234, 52.5},
		{0.00220636496208, 111320.7020209128, 51751.86112841131, 3796837.749470245, 992013.7397791013, -1221952.21711287, 1340652.697009075, -620943.6990984312, 144416.9293806241, 37.5},
		{-0.0003441963504368392, 111320.7020576856, 278.2353980772752, 2485758.690035394, 6070.750963243378, 54821.18345352118, 9540.606633304236, -2710.55326746645, 1405.483844121726, 22.5},
		{-0.0003218135878613132, 111320.7020701615, 0.00369383431289, 823725.6402795718, 0.46104986909093, 2351.343141331292, 1.58060784298199, 8.77738589078284, 0.37238884252424, 7.45}}
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
	return PointRound8(mglon, mglat)
}

// Gcj2Wgs  GCJ坐标系 ----> WGS坐标系
func Gcj2Wgs(lon, lat float64) (float64, float64) {
	dlon, dlat := _offset(lon, lat)
	mglat := lat - dlat
	mglon := lon - dlon
	return PointRound8(mglon, mglat)
}

// Gcj2bd  火星(GCJ-02)坐标系 ----> 百度(BD-09)坐标系
func Gcj2bd(lon, lat float64) (float64, float64) {
	x, y := lon, lat
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*_xpi)
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*_xpi)
	bdLon := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return PointRound8(bdLon, bdLat)
}

// Bd2gcj  百度(BD-09)坐标系 ----> 火星(GCJ-02)坐标系
func Bd2gcj(lon, lat float64) (float64, float64) {
	x, y := lon-0.0065, lat-0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*_xpi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*_xpi)
	ggLon := z * math.Cos(theta)
	ggLat := z * math.Sin(theta)
	return PointRound8(ggLon, ggLat)
}

// Wgs2bd WGS坐标系 ----> 百度坐标系
func Wgs2bd(lon, lat float64) (float64, float64) {
	x, y := Wgs2gcj(lon, lat)
	return Gcj2bd(x, y)
}

// Bd2Wgs 百度坐标系 ----> WGS坐标系
func Bd2Wgs(lon, lat float64) (float64, float64) {
	x, y := Bd2gcj(lon, lat)
	return Gcj2Wgs(x, y)
}

//===============================================================================

/***
http://www.movable-type.co.uk/scripts/latlong.html
http://en.wikipedia.org/wiki/Haversine_formula	Haversine
http://en.wikipedia.org/wiki/Spherical_law_of_cosines	Law of Cosines
http://en.wikipedia.org/wiki/Vincenty's_formulae	Vincenty
Speed: Law of Cosines > Haversine > Vincenty
***/

//Radians Radians
func Radians(r float64) float64 {
	return r * math.Pi / 180.0
}

//Degrees Degrees
func Degrees(d float64) float64 {
	return d * 180.0 / math.Pi
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

// PointDistance (in meter) Spherical_law_of_cosines
func PointDistance(lon1, lat1, lon2, lat2 float64) float64 {
	r, rad := 6371000.0, math.Pi/180.0
	lat1 = lat1 * rad
	lat2 = lat2 * rad
	lon1 = lon1 * rad
	lon2 = lon2 * rad
	theta := lon2 - lon1
	return r * math.Acos(math.Sin(lat1)*math.Sin(lat2)+
		math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
}

// PointDistHaversine (in meter) Haversine_formula
func PointDistHaversine(lon1, lat1, lon2, lat2 float64) float64 {
	r, rad := 6371000.0, math.Pi/180.0
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	lat1 = lat1 * rad
	lat2 = lat2 * rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

// PointMid 两点间的中间点
func PointMid(lon1, lat1, lon2, lat2 float64) (float64, float64) {
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

// PointAt   Destination point given distance and bearing from start point
func PointAt(lon, lat, dist, azimuth float64) (float64, float64) {
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

	φ1 := Radians(lat)
	λ1 := Radians(lon)
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

//===============================================================================
/*** 墨卡托坐标体系
https://en.wikipedia.org/wiki/Web_Mercator
https://en.wikipedia.org/wiki/Tile_Map_Service 瓦片地图服务
https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
https://github.com/CntChen/tile-lnglat-transform 提供了高德、百度、谷歌和腾讯地图的经纬度坐标与瓦片坐标的相互转换
https://lbs.amap.com/api/javascript-api/reference/map/  高德地图层级
http://lbsyun.baidu.com/index.php?title=webapi/guide/changeposition  百度API坐标转换
https://github.com/davvo/mercator
****/

// Wgs2Tile 瓦片:lnglat转XY
func Wgs2Tile(lng, lat float64, z int) (x, y int) {
	x = int(math.Floor((lng + 180.0) / 360.0 * (math.Exp2(float64(z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(lat*math.Pi/180.0)+1.0/math.Cos(lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(z)))))
	return
}

// Tile2Wgs 瓦片:XY转lnglat
func Tile2Wgs(x, y, z int) (lat, lng float64) {
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	lng = float64(x)/math.Exp2(float64(z))*360.0 - 180.0
	return lat, lng
}

// TileImage 经纬度转瓦片像素点位置
func TileImage(lng, lat float64, z int, deg int) (x, y, px, py int) {
	x, y = Wgs2Tile(lng, lat, z)
	x4, y4 := Wgs2Tile(lng, lat, z+int(math.Log(float64(deg))/math.Log(2)))
	px, py = x4-x*deg, y4-y*deg
	return
}

//===================百度经纬度<--->墨卡托============================================================

//Bd09ToTile 百度经纬度转换为瓦片编号
func Bd09ToTile(lng, lat float64, zoom int) (int, int) {
	x, y := Bd09ToMercator(lng, lat)
	cV := math.Pow(2, float64(18-zoom)) * 256
	return int(math.Floor(x / cV)), int(math.Floor(y / cV))
}

// MercatorToBd09 墨卡托坐标转百度经纬度坐标
func MercatorToBd09(x, y float64) (float64, float64) {
	cF := []float64{}
	x = math.Abs(x)
	yTemp := math.Abs(y)
	for cE := 0; cE < len(_mcband); cE++ {
		if yTemp >= _mcband[cE] {
			cF = _mc2ll[cE]
			break
		}
	}
	return yr(x, y, cF)
}

// Bd09ToMercator 百度经纬度坐标转墨卡托坐标
func Bd09ToMercator(lng, lat float64) (float64, float64) {
	getLoop := func(lng float64, min, max float64) float64 {
		for lng > max {
			lng = lng - math.Abs(max-min)
		}
		for lng < min {
			lng = lng + math.Abs(max-min)
		}
		return lng
	}
	getRange := func(lat float64, min, max float64) float64 {
		return math.Min(math.Max(lat, min), max)
	}

	cE := []float64{}
	lng = getLoop(lng, -180.0, 180.0)
	lat = getRange(lat, -74.0, 74.0)
	for i := 0; i < len(_llband); i++ {
		if lat >= _llband[i] {
			cE = _ll2mc[i]
			break
		}
	}
	if len(cE) == 0 {
		for i := len(_llband) - 1; i >= 0; i-- {
			if lat <= -_llband[i] {
				cE = _ll2mc[i]
				break
			}
		}
	}
	return yr(lng, lat, cE)
}

//yr 百度墨卡托解密函数
func yr(x, y float64, cE []float64) (float64, float64) {
	xTemp := cE[0] + cE[1]*math.Abs(x)
	cC := math.Abs(y) / cE[9]
	yTemp := cE[2] + cE[3]*cC + cE[4]*cC*cC + cE[5]*cC*cC*cC + cE[6]*cC*cC*cC*cC + cE[7]*cC*cC*cC*cC*cC + cE[8]*cC*cC*cC*cC*cC*cC
	if x < 0 {
		xTemp = xTemp * -1
	}
	if y < 0 {
		yTemp = yTemp * -1
	}
	return xTemp, yTemp
}

//===============================================================================

func demo() {
	lng, lat := 121.5012091398, 31.2355502882 //上海中心大厦gps
	bdlng, bdlat := Wgs2bd(lng, lat)          //31.239186 121.512245
	// lng, lat := MercatorToBd09(13525446.26, 3639969.64)
	x, y := Bd09ToMercator(bdlng, bdlat)
	fmt.Println(bdlng, bdlat)
	fmt.Println(x, y)
	fmt.Println(Bd09ToTile(x, y, 15))
}

//===============================================================================
