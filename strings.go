package xutil

import (
	"strconv"
	"strings"
)

//StringsContains
func IntsIndex(s1 []int, s int) int {
	for i, v := range s1 {
		if v == s {
			return i
		}
	}
	return -1
}

//StringsIndex
func StringsIndex(s1 []string, s string) int {
	for i, v := range s1 {
		if v == s {
			return i
		}
	}
	return -1
}

//StringsLower
func StringsLower(s1 []string) []string {
	s := make([]string, len(s1))
	for i, v := range s1 {
		s[i] = strings.ToLower(v)
	}
	return s
}

//StringsUpper
func StringsUpper(s1 []string) []string {
	s := make([]string, len(s1))
	for i, v := range s1 {
		s[i] = strings.ToUpper(v)
	}
	return s
}

//StringsMinus 差集s1-s2
func StringsMinus(s1, s2 []string) []string {
	smap := make(map[string]int, 0)
	s := make([]string, 0)
	for _, v := range s2 {
		smap[v] = 0
	}

	for _, v := range s1 {
		if _, exists := smap[v]; !exists {
			s = append(s, v)
		}
	}
	return s
}

//StringsIntersect 交集 s1∩s2 去重
func StringsIntersect(s1, s2 []string) []string {
	smap := make(map[string]int, 0)
	s := make([]string, 0)
	for _, v := range s2 {
		smap[v] = 0
	}

	for _, v := range s1 {
		if _, exists := smap[v]; exists {
			s = append(s, v)
		}
	}
	return s
}

//StringsUnion 并集 s1∪s2 去重
func StringsUnion(s1, s2 []string) []string {
	smap := make(map[string]int, 0)
	s := make([]string, 0)
	for _, v := range s2 {
		smap[v] = 0
	}

	for _, v := range s1 {
		if _, exists := smap[v]; !exists {
			smap[v] = 0
		}
	}
	for k := range smap {
		s = append(s, k)
	}
	return s
}

func SubString(str string, begin, length int) (substr string) {
	rs := []rune(str)
	lth := len(rs)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}
	return string(rs[begin:end])
}

//StringsUniq 去重
func StringsUniq(s1 []string) []string {
	s := make([]string, 0)
	smap := make(map[string]struct{}, 0)
	for _, v := range s1 {
		if _, exists := smap[v]; !exists {
			s = append(s, v)
			smap[v] = struct{}{}
		}
	}
	return s
}

// StringsReverse reverses an array of string
func StringsReverse(s []string) []string {
	for i, j := 0, len(s)-1; i < len(s)/2; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// StringReverse reverses a string
func StringReverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func StringsToInt(arr []string) (dat []int) {
	dat = make([]int, len(arr))
	for i := 0; i < len(arr); i++ {
		f, err := strconv.Atoi(arr[i])
		if err != nil {
			continue
		}
		dat[i] = f
	}
	return
}

func StringsToFloat64(arr []string) (dat []float64) {
	dat = make([]float64, len(arr))
	for i := 0; i < len(arr); i++ {
		f, err := strconv.ParseFloat(arr[i], 64)
		if err != nil {
			// fmt.Println(raw, err.Error())
			continue
		}
		dat[i] = f
	}
	return
}

func StringsToFloat64NoEmpty(arr []string) (dat []float64) {
	for i := 0; i < len(arr); i++ {
		f, err := strconv.ParseFloat(arr[i], 64)
		if err != nil {
			// fmt.Println(raw, err.Error())
			continue
		}
		dat = append(dat, f)
	}
	return
}
