package extractor

import (
	"reflect"
	"strconv"
	"strings"
)

var register = make(map[string]reflect.Value)

func init() {
	Register("ParseNumber", ParseNumber)
}

// Register you can register custom function to tag
func Register(name string, reg interface{}) {
	register[name] = reflect.ValueOf(reg)
}

// ParseNumber 解析带字符的数字 10k 10.3k 0.3k 10,000k 1m等
func ParseNumber(num string) (float64, error) {
	num = strings.Trim(num, " ")
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
	i, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, err
	}

	return i * factor, nil
}
