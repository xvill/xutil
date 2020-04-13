package xutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/panjf2000/ants"
)

type Poi struct {
	Status   int
	Message  string
	Lng      float64
	Lat      float64
	Addr     string
	Province string
	City     string
	District string
	CityCode string
	Level    string
	Info     map[string]string
}

type MapAPI struct {
	AK                string
	LimitCityShanghai bool
	SM                sync.Map
	wg                sync.WaitGroup
}

func NewMapAPI(ak string) MapAPI {
	return MapAPI{AK: ak, LimitCityShanghai: true}
}

//---------------------------------------------------------------------------------------------------------------------

// AmapGeoCode 高德解析地址为经纬度
func (m MapAPI) AmapGeoCode(address string) (poi Poi) {
	type AmapPoi struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		Infocode string `json:"infocode"`
		Count    string `json:"count"`
		Geocodes []struct {
			FormattedAddress string `json:"formatted_address"`
			Country          string `json:"country"`
			Province         string `json:"province"`
			Citycode         string `json:"citycode"`
			City             string `json:"city"`
			District         string `json:"district"`
			Adcode           string `json:"adcode"`
			Location         string `json:"location"`
			Level            string `json:"level"`
		} `json:"geocodes"`
	}

	// https://lbs.amap.com/api/webservice/guide/api/georegeo
	address = strings.Join(strings.Fields(address), "")
	APIURL := ""
	if m.LimitCityShanghai {
		APIURL = fmt.Sprintf("http://restapi.amap.com/v3/geocode/geo?city=上海&citylimit=true&key=%s&address=%s", m.AK, address)
	} else {
		APIURL = fmt.Sprintf("http://restapi.amap.com/v3/geocode/geo?key=%s&address=%s", m.AK, address)
	}
	resp, err := http.Get(APIURL)
	if err != nil {
		poi.Status = -1
		poi.Message = err.Error()
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		poi.Status = -1
		poi.Message = err.Error()
		return
	}

	mpoi := AmapPoi{}
	err = json.Unmarshal(body, &mpoi)
	if err != nil {
		poi.Status = -1
		poi.Message = err.Error()
		return
	}

	if len(mpoi.Geocodes) > 0 {
		poi.Addr = mpoi.Geocodes[0].FormattedAddress
		poi.Province = mpoi.Geocodes[0].Province
		poi.CityCode = mpoi.Geocodes[0].Citycode
		poi.City = mpoi.Geocodes[0].City
		poi.District = mpoi.Geocodes[0].District
		poi.Level = mpoi.Geocodes[0].Level
		loc := strings.Split(mpoi.Geocodes[0].Location, ",")
		lng, err1 := strconv.ParseFloat(loc[0], 64)
		lat, err2 := strconv.ParseFloat(loc[1], 64)
		if err1 == nil && err2 == nil {
			poi.Lng = lng
			poi.Lat = lat
		}
		poi.Status = 0
		poi.Message = ""
	} else {
		poi.Status = 1
		poi.Message = "EmptyData"
	}
	return
}

//_codeAmapAddr 解析地址,多次失败尝试
func (m *MapAPI) _codeAmapAddr(addr interface{}, cnt int) {
	unaddr := addr.(string)
	if cnt > 5 {
		return
	}
	poi := m.AmapGeoCode(unaddr)
	if poi.Status != 0 {
		cnt++
		m._codeAmapAddr(addr, cnt)
	} else {
		m.SM.Store(unaddr, poi)
	}
}

// AmapGeoCodeALL 高德解析地址为经纬度
func (m *MapAPI) AmapGeoCodeALL(addrsMap map[string]string, poolsize int) (addrsAll map[string]map[string]string) {
	p, _ := ants.NewPoolWithFunc(poolsize, func(addr interface{}) {
		m._codeAmapAddr(addr, 0)
		m.wg.Done()
	})
	defer p.Release()

	for k := range addrsMap {
		m.wg.Add(1)
		p.Invoke(k)
	}
	m.wg.Wait()

	addrsAll = make(map[string]map[string]string, 0)
	m.SM.Range(func(k, v interface{}) bool {
		addr := k.(string)
		geo := v.(map[string]string)
		addrsAll[addr] = geo
		return true
	})
	return addrsAll
}

//---------------------------------------------------------------------------------------------------------------------

// BdmapGeoCode 百度解析地址为经纬度
func (m MapAPI) BdmapGeoCode(address string) (poi Poi) {
	type BdmapPOI struct {
		Status  int    `json:"status"`
		Message string `json:"msg"`
		Result  struct {
			Location struct {
				Lng float64 `json:"lng"`
				Lat float64 `json:"lat"`
			} `json:"location"`
			Precise       int    `json:"precise"`
			Confidence    int    `json:"confidence"`
			Comprehension int    `json:"comprehension"`
			Level         string `json:"level"`
		} `json:"result"`
	}

	address = strings.Join(strings.Fields(address), "")
	url := fmt.Sprintf("http://api.map.baidu.com/geocoder/v2/?output=json&ak=%s&address=%s", m.AK, address)
	resp, err := http.Get(url)
	if err != nil {
		poi.Status = -1
		poi.Message = err.Error()
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		poi.Status = -1
		poi.Message = err.Error()
		return
	}

	mpoi := BdmapPOI{}
	err = json.Unmarshal(body, &mpoi)
	if err != nil {
		poi.Status = -1
		poi.Message = err.Error()
		return
	}
	if mpoi.Status != 0 {
		poi.Status = 1
		poi.Message = mpoi.Message
		return
	}
	// fmt.Println(mpoi)
	poi.Status = mpoi.Status
	poi.Message = mpoi.Message
	poi.Lng = mpoi.Result.Location.Lng
	poi.Lat = mpoi.Result.Location.Lat
	poi.Level = mpoi.Result.Level
	poi.Info = make(map[string]string, 0)
	poi.Info["confidence"] = fmt.Sprintf("%d", mpoi.Result.Confidence)
	poi.Info["comprehension"] = fmt.Sprintf("%d", mpoi.Result.Comprehension)
	poi.Info["precise"] = fmt.Sprintf("%d", mpoi.Result.Precise)
	return
}

//_codeBdmapAddr 解析地址,多次失败尝试
func (m *MapAPI) _codeBdmapAddr(addr interface{}, cnt int) {
	unaddr := addr.(string)
	if cnt > 5 {
		return
	}
	poi := m.BdmapGeoCode(unaddr)
	if poi.Status != 0 {
		cnt++
		m._codeAmapAddr(addr, cnt)
	} else {
		m.SM.Store(unaddr, poi)
	}
}

// BdmapGeoCodeALL 百度解析地址为经纬度
func (m *MapAPI) BdmapGeoCodeALL(addrsMap map[string]string, poolsize int) (addrsAll map[string]map[string]string) {
	p, _ := ants.NewPoolWithFunc(poolsize, func(addr interface{}) {
		m._codeBdmapAddr(addr, 0)
		m.wg.Done()
	})
	defer p.Release()

	for k := range addrsMap {
		m.wg.Add(1)
		p.Invoke(k)
	}
	m.wg.Wait()

	addrsAll = make(map[string]map[string]string, 0)
	m.SM.Range(func(k, v interface{}) bool {
		addr := k.(string)
		geo := v.(map[string]string)
		addrsAll[addr] = geo
		return true
	})
	return addrsAll
}

// mocator := "4|13534914.0122,3645387.5227;13535422.4951,3645834.93158|1-13534914.0122,3645542.22157,13534919.2537,3645555.85957,13535111.9804,3645672.21552,13535163.1497,3645713.32995;"

//BmapMocator  百度墨卡托解析
func (m MapAPI) BmapMocator(mocator string) string {
	if len(mocator) <= 0 {
		return ""
	}
	geos := strings.Split(mocator, "|")
	plm := strings.Split(geos[2], ";")
	geo := ""
	if geos[0] == "4" {
		for i := 0; i < len(plm); i++ {
			geoPaths := strings.Split(plm[i], "-")
			if geoPaths[0] == "1" {
				geo = geoPaths[1]
			}
		}
	}

	// 墨卡托坐标解析
	var sb bytes.Buffer
	sb.WriteString("LINESTRING (")
	geoPolyline := strings.Split(geo, ",")
	for i := 0; i < len(geoPolyline); i = i + 2 {
		x, _ := strconv.ParseFloat(geoPolyline[i], 64)
		y, _ := strconv.ParseFloat(geoPolyline[i+1], 64)
		lng, lat := MercatorToBd09(x, y)
		sb.WriteString(fmt.Sprintf("%g %g,", lng, lat))
	}
	sb.Bytes()[sb.Len()-1] = ')'
	return sb.String()
}
