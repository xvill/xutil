package xutil

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
