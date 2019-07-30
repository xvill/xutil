package xutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/panjf2000/ants"
)

type MapAPI struct {
	AK string
	SM sync.Map
	wg sync.WaitGroup
}

func NewMapAPI(ak string) MapAPI {
	return MapAPI{AK: ak}
}

// AmapGeoCode 高德解析地址为经纬度
func (m MapAPI) AmapGeoCode(address string) (poi map[string]string, err error) {
	// https://lbs.amap.com/api/webservice/guide/api/georegeo
	address = strings.Join(strings.Fields(address), "")
	url := fmt.Sprintf("http://restapi.amap.com/v3/geocode/geo?city=上海&citylimit=true&key=%s&address=%s", m.AK, address)
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

//_codeAmapAddr 解析地址,多次失败尝试
func (m *MapAPI) _codeAmapAddr(addr interface{}, cnt int) {
	unaddr := addr.(string)
	if cnt > 5 {
		return
	}
	geo, _ := m.AmapGeoCode(unaddr)
	gps := geo["location"]
	if gps == "" {
		cnt++
		m._codeAmapAddr(addr, cnt)
	} else {
		m.SM.Store(unaddr, geo)
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

// BdmapGeoCode 百度解析地址为经纬度
func (m MapAPI) BdmapGeoCode(address string) (poi map[string]string, err error) {
	address = strings.Join(strings.Fields(address), "")
	url := fmt.Sprintf("http://api.map.baidu.com/geocoder/v2/?output=json&ak=%s&address=%s", m.AK, address)
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
