package xutil

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Abs 整数绝对值 https://segmentfault.com/a/1190000013201491
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// ---------------------------------------------------------------------------------------

func ColFloat64(raw, fields, oper string) (ret float64) {
	dat := StringToFloat64(raw, fields)
	if len(dat) == 0 {
		return
	}
	switch oper {
	case "SUM":
		ret = SumFloat64(dat)
	case "MAX":
		ret = MaxFloat64(dat)
	case "MIN":
		ret = MinFloat64(dat)
	case "AVG", "MEAN":
		ret = AvgFloat64(dat)
	case "STDDEV":
		ret = StdDevFloat64(dat)
	}

	return ret
}

func DiffSqrtMeanFloat64(a []float64) []float64 {
	set := make([]float64, len(a))
	meanVal := AvgFloat64(a)
	var d float64
	for i, v := range a {
		d = v - meanVal
		set[i] = d * d
	}
	return set
}

func StdDevFloat64(a []float64) float64 {
	return math.Sqrt(AvgFloat64(DiffSqrtMeanFloat64(a)))
}

func StringToFloat64(raw, fields string) (dat []float64) {
	arr := strings.Split(raw, fields)
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

func SumFloat64(s []float64) (sum float64) {
	for _, v := range s {
		sum += v
	}
	return
}

func MaxFloat64(i []float64) float64 {
	if len(i) == 0 {
		panic("arg is an empty array/slice")
	}
	var max float64
	for idx := 0; idx < len(i); idx++ {
		item := i[idx]
		if idx == 0 {
			max = item
			continue
		}
		if item > max {
			max = item
		}
	}
	return max
}

func MinFloat64(i []float64) float64 {
	if len(i) == 0 {
		panic("arg is an empty array/slice")
	}
	var min float64
	for idx := 0; idx < len(i); idx++ {
		item := i[idx]
		if idx == 0 {
			min = item
			continue
		}
		if item < min {
			min = item
		}
	}
	return min
}

func MeanFloat64(i []float64) float64 {
	return AvgFloat64(i)
}

func AvgFloat64(i []float64) float64 {
	if len(i) == 0 {
		panic("arg is an empty array/slice")
	}
	return SumFloat64(i) / float64(len(i))
}

//=====================任意进制转换=========================================
var tenToAny map[int]string = map[int]string{0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9", 10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f", 16: "g", 17: "h", 18: "i", 19: "j", 20: "k", 21: "l", 22: "m", 23: "n", 24: "o", 25: "p", 26: "q", 27: "r", 28: "s", 29: "t", 30: "u", 31: "v", 32: "w", 33: "x", 34: "y", 35: "z", 36: ":", 37: ";", 38: "<", 39: "=", 40: ">", 41: "?", 42: "@", 43: "[", 44: "]", 45: "^", 46: "_", 47: "{", 48: "|", 49: "}", 50: "A", 51: "B", 52: "C", 53: "D", 54: "E", 55: "F", 56: "G", 57: "H", 58: "I", 59: "J", 60: "K", 61: "L", 62: "M", 63: "N", 64: "O", 65: "P", 66: "Q", 67: "R", 68: "S", 69: "T", 70: "U", 71: "V", 72: "W", 73: "X", 74: "Y", 75: "Z"}

// 10进制转任意进制
func DecimalToAny(num, n int) string {
	new_num_str := ""
	var remainder int
	var remainder_string string
	for num != 0 {
		remainder = num % n
		if 76 > remainder && remainder > 9 {
			remainder_string = tenToAny[remainder]
		} else {
			remainder_string = strconv.Itoa(remainder)
		}
		new_num_str = remainder_string + new_num_str
		num = num / n
	}
	return new_num_str
}

// map根据value找key
func findkey(in string) int {
	result := -1
	for k, v := range tenToAny {
		if in == v {
			result = k
		}
	}
	return result
}

// 任意进制转10进制
func AnyToDecimal(num string, n int) int {
	var new_num float64
	new_num = 0.0
	nNum := len(strings.Split(num, "")) - 1
	for _, value := range strings.Split(num, "") {
		tmp := float64(findkey(value))
		if tmp != -1 {
			new_num = new_num + tmp*math.Pow(float64(n), float64(nNum))
			nNum = nNum - 1
		} else {
			break
		}
	}
	return int(new_num)
}

func demoTo() {
	fmt.Println(DecimalToAny(9999, 76))
	fmt.Println(AnyToDecimal("1F[", 76))
}

//=======================================================================
