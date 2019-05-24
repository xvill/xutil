package xutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Geo struct {
	Type   string
	Coords [][][][]float64
}

type Point struct {
	X float64
	Y float64
}

func (p Point) String() string {
	return fmt.Sprintf("%g,%g", p.X, p.Y)
}

type Line struct {
	P1 Point
	P2 Point
}

//===============================================================================

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

//===============================================================================

func (g Geo) Lines() []Line {
	lines := []Line{}
	for _, a := range g.Coords {
		for _, b := range a {
			lines = append(lines, Line{Point{b[0][0], b[0][1]}, Point{b[1][0], b[1][1]}})
		}
	}
	return lines
}
func (g Geo) Points() []Point {
	points := []Point{}
	for _, a := range g.Coords {
		for _, b := range a {
			for _, c := range b {
				points = append(points, Point{c[0], c[1]})
			}
		}
	}
	return points
}

//===============================================================================

// FromWKT 解析WKT为Geo
func FromWKT(wkt string) (g Geo, err error) {
	wkt = strings.NewReplacer("(", " [ ", ")", " ] ", ",", " , ").Replace(wkt)
	bf := bufio.NewScanner(strings.NewReader(wkt))
	bf.Split(bufio.ScanWords)

	var retstr bytes.Buffer
	flag := false
	bf.Scan()
	_type := bf.Text()
	for bf.Scan() {
		w := bf.Text()
		if flag && w != "[" && w != "]" && w != "," {
			retstr.WriteRune(',')
			retstr.WriteString(w)
			retstr.WriteRune(']')
			flag = true
		} else if w != "[" && w != "]" && w != "," {
			retstr.WriteRune('[')
			retstr.WriteString(w)
			flag = true
		} else {
			retstr.WriteString(w)
			flag = false
		}
	}
	_coordinates := retstr.String()
	_type = strings.NewReplacer("POINT", "Point", "LINESTRING", "LineString", "MULTILINESTRING", "MultiLineString",
		"POLYGON", "Polygon", "MULTIPOLYGON", "MultiPolygon", "MULTIPOINT", "MultiPoint").Replace(strings.ToUpper(_type))

	rawjson := fmt.Sprintf(`{"type":"%s","coordinates":%s}`, _type, _coordinates)
	return FromGeoJSON(rawjson)
}

// FromGeoJSON 解析GeoJSON为Geo
func FromGeoJSON(geojson string) (g Geo, err error) {
	type GeoJSON struct {
		Type   string          `json:"type"`
		Coords json.RawMessage `json:"coordinates"`
	}

	var gj GeoJSON
	err = json.Unmarshal([]byte(geojson), &gj)
	if err != nil {
		return g, err
	}

	g.Coords = [][][][]float64{{{{}}}}
	g.Type = strings.NewReplacer("POINT", "Point", "LINESTRING", "LineString", "MULTILINESTRING", "MultiLineString",
		"POLYGON", "Polygon", "MULTIPOLYGON", "MultiPolygon", "MULTIPOINT", "MultiPoint").Replace(strings.ToUpper(gj.Type))

	var v4 [][][][]float64
	err = json.Unmarshal(gj.Coords, &v4)
	if err == nil {
		g.Coords = v4
		return
	}

	var v3 [][][]float64
	err = json.Unmarshal(gj.Coords, &v3)
	if err == nil {
		g.Coords[0] = v3
		return
	}

	var v2 [][]float64
	err = json.Unmarshal(gj.Coords, &v2)
	if err == nil {
		g.Coords[0][0] = v2
		return
	}

	var v1 []float64
	err = json.Unmarshal(gj.Coords, &v1)
	if err == nil {
		g.Coords[0][0][0] = v1
		return
	}

	return

}

// GeoJSON 生成GeoJSON
func (g Geo) GeoJSON() (s string, err error) {
	s, err = g.CoordsJSON()
	s = fmt.Sprintf(`{"type":"%s","coordinates":%s}`, g.Type, s)
	return s, err
}

func (g Geo) CoordsJSON() (s string, err error) {
	var b []byte
	switch g.Type {
	case "Point":
		b, err = json.Marshal(g.Coords[0][0][0])
	case "LineString", "MultiPoint":
		b, err = json.Marshal(g.Coords[0][0])
	case "Polygon", "MultiLineString":
		b, err = json.Marshal(g.Coords[0])
	case "MultiPolygon":
		b, err = json.Marshal(g.Coords)
	}
	return string(b), err
}

func (g Geo) String() (wkt string) {
	return g.ToWKT()
}

// ToWKT 生成WKT
func (g Geo) ToWKT() (wkt string) {
	coords := g.Coords
	var points, polygon, multipolygon []string
	for _, a := range coords {
		polygon = make([]string, 0)
		for _, b := range a {
			points = make([]string, 0)
			for _, c := range b {
				points = append(points, fmt.Sprintf("%g %g", c[0], c[1]))
			}
			polygon = append(polygon, fmt.Sprintf("( %s)", strings.Join(points, ", ")))
		}
		multipolygon = append(multipolygon, fmt.Sprintf("(%s)", strings.Join(polygon, ", ")))
	}

	switch g.Type {
	case "Point":
		wkt = fmt.Sprintf("POINT (%s)", points[0])
	case "MultiPoint":
		wkt = fmt.Sprintf("MULTIPOINT (%s)", strings.Join(points, ","))
	case "LineString":
		wkt = fmt.Sprintf("LINESTRING (%s)", strings.Join(points, ","))
	case "MultiLineString":
		wkt = fmt.Sprintf("MULTILINESTRING (%s)", strings.Join(polygon, ","))
	case "Polygon":
		wkt = fmt.Sprintf("POLYGON (%s)", strings.Join(polygon, ","))
	case "MultiPolygon":
		wkt = fmt.Sprintf("MULTIPOLYGON (%s)", strings.Join(multipolygon, ","))
	}
	return
}

// PointFunc 对所有点应用函数
func (g Geo) PointFunc(f func(lon, lat float64) (float64, float64)) {
	coords := g.Coords
	for _, a := range coords {
		for _, b := range a {
			for _, c := range b {
				c[0], c[1] = f(c[0], c[1])
			}
		}
	}
}

// ReverseLngLat 转换Lat,Lng 位置
func (g Geo) ReverseLngLat() {
	f := func(lon, lat float64) (float64, float64) { return lat, lon }
	g.PointFunc(f)
}

// Wgs2gcj 经纬度坐标系转换 wgs-> gcj
func (g Geo) Wgs2gcj() {
	g.PointFunc(Wgs2gcj)
}

// Gcj2bd 经纬度坐标系转换 gcj->BD09
func (g Geo) Gcj2bd() {
	g.PointFunc(Gcj2bd)
}

// Wgs2bd 经纬度坐标系转换 wgs->BD09
func (g Geo) Wgs2bd() {
	g.PointFunc(Wgs2bd)
}

// Box 方框边界 minx, miny, maxx, maxy
func (g Geo) Box() []float64 {
	coords := g.Coords
	minx, miny, maxx, maxy := coords[0][0][0][0], coords[0][0][0][1], coords[0][0][0][0], coords[0][0][0][1]
	for _, a := range coords {
		for _, b := range a {
			for _, c := range b {
				if c[0] > maxx {
					maxx = c[0]
				}
				if c[0] < minx {
					minx = c[0]
				}
				if c[1] > maxy {
					maxy = c[1]
				}
				if c[1] < miny {
					miny = c[1]
				}
			}
		}
	}
	return []float64{minx, miny, maxx, maxy}
}

// IsClockwise  Green公式判断顺时针
// isClockwise  Green公式判断顺时针
func isClockwise(lnglats [][]float64) bool {
	d := 0.0
	n := len(lnglats)
	for i := 0; i < n-1; i++ {
		d += -0.5 * (lnglats[i][1] + lnglats[i+1][1]) * (lnglats[i+1][0] - lnglats[i][0])
		//d+= －0.5*(y[i+1]+y[i])*(x[i+1]-x[i])
	}
	if d > 0 {
		return false //counter clockwise
	}
	return true // clockwise
}
