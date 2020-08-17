package extractor

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var register = make(map[string]reflect.Value)

func init() {
	Register("ParseNumber", ParseNumber) // 自定义函数
	Register("ExtractNumber", ExtractNumber)
}

// Register you can register custom function to tag
func Register(name string, reg interface{}) {
	register[name] = reflect.ValueOf(reg)
}

// ExtractNumber 通过正则获取数字, 然后解析ParseNumber
func ExtractNumber(num string) ([]float64, error) {
	var ret []float64
	for _, e := range regexp.MustCompile(`[\d,kKmM\.]+`).FindAllString(num, -1) {
		pn, err := ParseNumber(e)
		if err != nil {
			log.Println(err)
		} else {
			ret = append(ret, pn)
		}
	}
	return ret, nil
}

// ParseNumber 解析带字符的数字 10k 10.3k 0.3k 10,000k 1m等
func ParseNumber(snum string) (float64, error) {
	num := strings.Trim(snum, " ")
	num = strings.ReplaceAll(num, ",", "")
	last := num[len(num)-1]
	factor := 1.0
	switch {
	case last == 'k' || last == 'K':
		factor = 1000.0
		num = num[0 : len(num)-1]
	case last == 'm' || last == 'M':
		factor = 1000000.0
		num = num[0 : len(num)-1]
	}

	if len(num) == 0 {
		err := fmt.Errorf("%s is not number type", snum)
		log.Println(err)
		return 0, err
	}

	i, err := strconv.ParseFloat(num, 64)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return i * factor, nil
}
