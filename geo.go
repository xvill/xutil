package xtools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Geo struct {
	Type   string
	Coords [][][][]float64
}

func WKTToGeoJSON(wkt string) (_type, _coordinates string) {
	str := strings.NewReplacer("(", " [ ", ")", " ] ", ",", " , ").Replace(wkt)

	bf := bufio.NewScanner(strings.NewReader(str))
	bf.Split(bufio.ScanWords)

	var retstr bytes.Buffer
	flag := false
	bf.Scan()
	_type = bf.Text()
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
	_coordinates = retstr.String()
	// _type = strings.ToUpper(_type)
	_type = strings.NewReplacer("POINT", "Point", "LINESTRING", "LineString", "MULTILINESTRING", "MultiLineString",
		"POLYGON", "Polygon", "MULTIPOLYGON", "MultiPolygon", "MULTIPOINT", "MultiPoint").Replace(strings.ToUpper(_type))

	return
}

// FromWKT 解析WKT为Geo
func FromWKT(wkt string) (Geo, error) {
	_type, _coordinates := WKTToGeoJSON(wkt)
	_coords := ""
	switch _type {
	case "Point", "LineString", "MultiPoint":
		_coords = fmt.Sprintf("[[%s]]", _coordinates)
	case "Polygon", "MultiLineString":
		_coords = fmt.Sprintf("[%s]", _coordinates)
	case "MultiPolygon":
		_coords = _coordinates
	}

	var v [][][][]float64
	err := json.Unmarshal([]byte(_coords), &v)
	return Geo{Type: _type, Coords: v}, err
}

// GeoJSON 生成GeoJSON
func (g Geo) GeoJSON() (s string, err error) {
	var b []byte
	switch g.Type {
	case "Point", "LineString", "MultiPoint":
		b, err = json.Marshal(g.Coords[0][0])
	case "Polygon", "MultiLineString":
		b, err = json.Marshal(g.Coords[0])
	case "MultiPolygon":
		b, err = json.Marshal(g.Coords)
	}
	return fmt.Sprintf(`{"type":"%s","coordinates":"%s"}`, g.Type, string(b)), err

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

// ReserveLngLat 转换Lat,Lng 位置
func (g Geo) ReserveLngLat() {
	coords := g.Coords
	for _, a := range coords {
		for _, b := range a {
			for _, c := range b {
				c[0], c[1] = c[1], c[0]
			}
		}
	}
}

// Wgs2gcj 经纬度坐标系转换 wgs-> gcj
func (g Geo) Wgs2gcj() {
	coords := g.Coords
	for _, a := range coords {
		for _, b := range a {
			for _, c := range b {
				c[0], c[1] = Wgs2gcj(c[0], c[1])
			}
		}
	}
}

// Gcj2bd 经纬度坐标系转换 gcj->BD09
func (g Geo) Gcj2bd() {
	coords := g.Coords
	for _, a := range coords {
		for _, b := range a {
			for _, c := range b {
				c[0], c[1] = Gcj2bd(c[0], c[1])
			}
		}
	}
}

// Wgs2bd 经纬度坐标系转换 wgs->BD09
func (g Geo) Wgs2bd() {
	coords := g.Coords
	for _, a := range coords {
		for _, b := range a {
			for _, c := range b {
				c[0], c[1] = Wgs2bd(c[0], c[1])
			}
		}
	}
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
