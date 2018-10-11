package xutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
* 身份证15位编码规则：dddddd yymmdd xx p
* 身份证18位编码规则：dddddd yyyymmdd xx p y
* dddddd: 地区编码
* yymmdd: 出生年(两位年)月日，如：910215
* yyyymmdd: 出生年(四位年)月日，如：19910215
* xx: 顺序编码，系统产生，无法确定
* p: 性别，奇数为男，偶数为女
* y: 校验码，该位数值可通过前17位计算获得
*
* 前17位号码加权因子为 Wi = [ 7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2 ]
* 验证位 Y = [ 1, 0, 10, 9, 8, 7, 6, 5, 4, 3, 2 ]
* 如果验证码恰好是10，为了保证身份证是十八位，那么第十八位将用X来代替
* 校验位计算: Y_P = mod( ∑(Ai×Wi),11 )
* i为身份证号码1...17 位; Y_P为校验码Y所在校验码数组位置
 */
type IDCard struct {
	ID        string
	Gender    string
	Age       string
	Birthdate string
	Province  string
	City      string
	District  string
}

var maddr map[string]string

// init 初始化 加载行政区信息
func init() {
	fname := "idaddr.json"
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(dat, &maddr)
	if err != nil {
		fmt.Println(err)
	}
}

// IDsumY 计算身份证的第十八位校验码
func sumY(id string) string {
	wi := [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}     //前17位号码加权因子
	y := [11]string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"} //身份证第18位校检码
	ai, res := make([]int, 17), 0
	for i := 0; i < 17; i++ {
		ai[i], _ = strconv.Atoi(string(id[i]))
	}
	for i := 0; i < 17; i++ {
		res += ai[i] * wi[i]
	}
	return y[res%11]
}

// ID15to18 将15位身份证转换为18位的
func ID15to18(id string) string {
	newid := id[:6] + "19" + id[6:] //在年份前添加19,变为17位
	return newid + sumY(newid)      //添加最后一位验证码
}

// IDisValid 校验身份证第18位是否正确
func IDisValid(id string) bool {
	id = strings.ToUpper(id)
	if sumY(id) != string(id[17]) {
		return false
	}
	return true
}

// IDisPattern 二代身份证正则表达式
func IDisPattern(id string) bool {
	pat := "^[1-9]\\d{5}[1-9]\\d{3}((0\\d)|(1[0-2]))(([0|1|2]\\d)|3[0-1])\\d{3}([\\d|x|X]{1})$"
	reg := regexp.MustCompile(pat)
	return reg.Match([]byte(id))
}

// NewIDCard  获取身份证信息
func NewIDCard(id string) (c IDCard, err error) {
	if len(id) == 15 {
		id = ID15to18(id)
	}
	c.ID = id

	if !IDisPattern(id) {
		return c, errors.New(" err_pattern")
	}

	if !IDisValid(id) {
		return c, errors.New(" err_validatecode")
	}
	t, err := time.Parse("20060102", string(id[6:14]))
	if err != nil {
		return c, errors.New(" err_birthdate")
	}

	addr := maddr[string(id[0:6])]
	if addr == "" {
		return c, errors.New(" err_address")
	}
	c.District = addr
	c.Province = maddr[string(id[0:2])+"0000"]
	c.City = maddr[string(id[0:4])+"00"]
	//------------------------------------
	now := time.Now()
	years := now.Year() - t.Year()
	if now.Month() < t.Month() || (now.Month() == t.Month() && now.Day() < t.Day()) {
		years--
	}
	c.Age = strconv.Itoa(years)
	c.Birthdate = t.Format("2006-01-02")
	//------------------------------------
	sex, _ := strconv.Atoi(string(id[16]))
	if sex%2 == 0 {
		c.Gender = "F"
	} else {
		c.Gender = "M"
	}
	//------------------------------------
	return c, nil
}

func demo() {
	fmt.Println(maddr["140522"])
	fmt.Println(ID15to18("210212831019104"))
	fmt.Println(NewIDCard("210212831019104"))
	fmt.Println(sumY("210212198310191044"))
	fmt.Println(IDisValid("210212198310191044"))
	fmt.Println(IDisPattern("210212198310191044"))
}
