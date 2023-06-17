package extractor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/474420502/extractor/htmlquery"

	"github.com/pkg/errors"
)

// HmtlExtractor 提取器
type HmtlExtractor struct {
	content []byte
	// doc     types.Document
	doc *htmlquery.Node
}

// ExtractHtmlString extractor xml(html)
func ExtractHtmlString(content string) *HmtlExtractor {
	return ExtractHtml([]byte(content))
}

// ExtractHtml extractor xml(html)
func ExtractHtml(content []byte) *HmtlExtractor {
	doc, err := htmlquery.Parse(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}
	e := &HmtlExtractor{}
	e.doc = doc
	e.content = content
	return e
}

// ExtractHtmlReader extractor xml(html)
func ExtractHtmlReader(in io.Reader) *HmtlExtractor {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(in); err != nil {
		panic(errors.Wrap(err, "failed to rea from io.Reader"))
	}
	return ExtractHtml(buf.Bytes())
}

// RegexpBytes multi xpath extractor
func (etor *HmtlExtractor) RegexpBytes(exp string) [][][]byte {
	return regexp.MustCompile(exp).FindAllSubmatch(etor.content, -1)
}

// RegexpString multi xpath extractor
func (etor *HmtlExtractor) RegexpString(exp string) [][]string {
	return regexp.MustCompile(exp).FindAllStringSubmatch(string(etor.content), -1)
}

// GetObjectByTag single xpath extractor
func (etor *HmtlExtractor) GetObjectByTag(obj interface{}) {
	v := reflect.ValueOf(obj)
	vtype := v.Type()
	if v.Kind() != reflect.Ptr {
		log.Panic("obj must ptr")
	}
	vtype = vtype.Elem()
	getInfoByTag(etor.doc, getFieldTags(vtype), v.Elem())
}

// XPaths multi xpath extractor
func (etor *HmtlExtractor) XPath(exp string) (*XPath, error) {
	result, err := etor.doc.QueryAll(exp)
	return newXPath(result...), err
}

// ErrorFlags  忽略错误标志位, 暂时不用
type ErrorFlags int

const (
	// ErrorBreak 遇到错误, 终止执行
	ErrorBreak ErrorFlags = 0
	// ErrorSkip 遇到错误, 忽略继续执行
	ErrorSkip ErrorFlags = 1
)

// XPath for easy extractor data
type XPath struct {
	results    []*htmlquery.Node
	errorFlags ErrorFlags
}

func newXPath(result ...*htmlquery.Node) *XPath {
	xp := &XPath{results: result, errorFlags: ErrorSkip}
	return xp
}

// GetXPathResults Get Current XPath Results
func (xp *XPath) GetXPathResults() []*htmlquery.Node {
	return xp.results
}

// GetStrings Get Current XPath Results
func (xp *XPath) GetStrings() []string {
	var ret []string
	for _, xresult := range xp.results {
		ret = append(ret, xresult.OutputHTML(true))
	}
	return ret
}

// GetAttrValuesByKey Get the Attribute of the Current XPath Results
func (xp *XPath) GetAttrValuesByKey(key string) []string {
	if len(xp.results) == 0 {
		return nil
	}

	var attrs []string
	for _, xpresult := range xp.results {
		if attr, err := xpresult.AttributeValue(key); err == nil {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}

// GetAttrKeysByValue Get the Attribute of the Current XPath Results
func (xp *XPath) GetAttrKeysByValue(value string) []string {
	if len(xp.results) == 0 {
		return nil
	}

	var attrs []string
	for _, xpresult := range xp.results {
		if attr := xpresult.GetAttributeByValue(value); attr != nil {
			attrs = append(attrs, attr.Key)
		}
	}
	return attrs
}

// Attributes Get the Attribute of the Current XPath Results
func (xp *XPath) Attributes(key string) []*htmlquery.Attribute {
	return xp.GetAttributes(key)
}

// GetAttributes Get the Attribute of the Current XPath Results
func (xp *XPath) GetAttributes(key string) []*htmlquery.Attribute {
	if len(xp.results) == 0 {
		return nil
	}

	var attrs []*htmlquery.Attribute
	for _, xpresult := range xp.results {
		if attr := xpresult.Attribute(key); attr != nil {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}

// GetTexts Get the Text of the Current XPath Results
func (xp *XPath) GetTexts() []string {
	if len(xp.results) == 0 {
		return nil
	}

	var txts []string
	for _, xpresult := range xp.results {
		txts = append(txts, xpresult.Text())
	}
	return txts
}

// GetTagNames Get the NodeValue of the Current XPath Results
func (xp *XPath) GetTagNames() []string {
	if len(xp.results) == 0 {
		return nil
	}

	var nvalues []string
	for _, xpresult := range xp.results {
		tn, err := xpresult.TagName()
		if err == nil {
			nvalues = append(nvalues, tn)
		} else {
			log.Println(err)
		}
	}

	return nvalues
}

// GetNodeStrings Get the String of the Current XPath Results

type methodtag struct {
	IsRegister bool            // is register 是否为注册函数
	Method     string          // method name 方法名
	Args       []reflect.Value // Args 参数
}

type fieldtag struct {
	Type   reflect.Type // 参考reflect
	Kind   reflect.Kind // 参考reflect
	VType  string       // Type的字符串形式 eg: String Int64 time.Time...
	VIndex int          // exp results selected index
	MIndex int          // method results selected index
	Index  int          // index
	Exp    string       // expression 表达式
	// Method string
	// Args   []reflect.Value
	Methods []methodtag // multi method 多个方法
}

// DefaultMethod 默认函数 如果tag没写mth(method) 的标识. 默认就是call Text()
var DefaultMethod = "Text"

var methodDict map[string]string

// 方法映射 动态调用过程能映射自定义方法
type nodeMethod string

const (

	// GetAttribute Node.GetAttribute()
	GetAttribute   nodeMethod = "GetAttribute"
	NodeName       nodeMethod = "TagName"
	ParentNode     nodeMethod = "ParentNode"
	AttributeValue nodeMethod = "AttributeValue"
	Text           nodeMethod = "Text"
	String         nodeMethod = "String"
)

func init() {
	methodDict = make(map[string]string)
	methodDict["Attribute"] = string(GetAttribute)
	methodDict["AttrValue"] = string(AttributeValue)
	methodDict["Name"] = string(NodeName)
}

// 获取成员变量的tag信息.
func getFieldTags(otype reflect.Type) []*fieldtag {

	var fieldtags []*fieldtag
	for i := 0; i < otype.NumField(); i++ {

		f := otype.Field(i)
		// 获取表达式 TODO: 转义之类的支持 正则之类的支持. json之类的支持 ...
		if exp, ok := f.Tag.Lookup("exp"); ok {
			ft := &fieldtag{}
			ft.Index = i
			ft.Exp = exp
			ft.Kind = f.Type.Kind()
			ft.Type = otype

			var smethod string
			var ok bool

			// 获取函数信息 method == mth
			for _, mth := range []string{"method", "mth"} {
				if smethod, ok = f.Tag.Lookup(mth); ok {
					for _, method := range strings.Split(smethod, " ") {
						methodAndArgs := strings.Split(method, ",")
						mt := methodtag{}
						mt.Method = methodAndArgs[0]

						// 注册函数的前置标志判断
						mtsp := strings.Split(mt.Method, ":")
						if len(mtsp) == 2 {
							switch mtsp[0] {
							case "r":
								fallthrough
							case "R":
								mt.IsRegister = true
								mt.Method = mtsp[1]
							default:
								panic(fmt.Errorf("flag %s is not exists", mtsp[0]))
							}
						}

						if v, ok := methodDict[mt.Method]; ok {
							mt.Method = v
						}
						// ft.Method = method[0]
						var args []reflect.Value = nil
						for _, arg := range methodAndArgs[1:] {
							args = append(args, reflect.ValueOf(arg))
						}
						mt.Args = args
						ft.Methods = append(ft.Methods, mt)
					}
					break
				}
			}

			if !ok {
				mt := methodtag{}
				mt.Method = DefaultMethod
				if v, ok := methodDict[mt.Method]; ok {
					mt.Method = v
				}
				mt.Args = nil
				ft.Methods = append(ft.Methods, mt)
			}

			ft.VType = ft.Type.Field(ft.Index).Type.String()
			ft.VType = strings.ReplaceAll(ft.VType, "[]", "")
			// 获取index
			if index, ok := f.Tag.Lookup("index"); ok {
				i, err := strconv.Atoi(index)
				if err != nil {
					log.Panic(err)
				}
				ft.VIndex = i
			} else {
				ft.VIndex = -1
			}
			// 获取mindex
			if index, ok := f.Tag.Lookup("mindex"); ok {
				i, err := strconv.Atoi(index)
				if err != nil {
					log.Panic(err)
				}
				ft.MIndex = i
			} else {
				ft.MIndex = -1
			}

			fieldtags = append(fieldtags, ft)
		}
	}
	return fieldtags
}

func autoValueType(vtype string, v interface{}) reflect.Value {
	switch vtype {
	case "int":

		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(rv)
		case int32:
			return reflect.ValueOf(int(rv))
		case int64:
			return reflect.ValueOf(int(rv))
		case uint:
			return reflect.ValueOf(int(rv))
		case uint32:
			return reflect.ValueOf(int(rv))
		case uint64:
			return reflect.ValueOf(int(rv))
		case float32:
			return reflect.ValueOf(int(rv))
		case float64:
			return reflect.ValueOf(int(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}

	case "int32":

		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(int32(rv))
		case int32:
			return reflect.ValueOf(rv)
		case int64:
			return reflect.ValueOf(int32(rv))
		case uint:
			return reflect.ValueOf(int32(rv))
		case uint32:
			return reflect.ValueOf(int32(rv))
		case uint64:
			return reflect.ValueOf(int32(rv))
		case float32:
			return reflect.ValueOf(int32(rv))
		case float64:
			return reflect.ValueOf(int32(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}

	case "int64":

		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(int64(rv))
		case int32:
			return reflect.ValueOf(int64(rv))
		case int64:
			return reflect.ValueOf(rv)
		case uint:
			return reflect.ValueOf(int64(rv))
		case uint32:
			return reflect.ValueOf(int64(rv))
		case uint64:
			return reflect.ValueOf(int64(rv))
		case float32:
			return reflect.ValueOf(int64(rv))
		case float64:
			return reflect.ValueOf(int64(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}

	case "uint":

		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(uint(rv))
		case int32:
			return reflect.ValueOf(uint(rv))
		case int64:
			return reflect.ValueOf(uint(rv))
		case uint:
			return reflect.ValueOf(rv)
		case uint32:
			return reflect.ValueOf(uint(rv))
		case uint64:
			return reflect.ValueOf(uint(rv))
		case float32:
			return reflect.ValueOf(uint(rv))
		case float64:
			return reflect.ValueOf(uint(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}

	case "uint32":

		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(uint32(rv))
		case int32:
			return reflect.ValueOf(uint32(rv))
		case int64:
			return reflect.ValueOf(uint32(rv))
		case uint:
			return reflect.ValueOf(uint32(rv))
		case uint32:
			return reflect.ValueOf(rv)
		case uint64:
			return reflect.ValueOf(uint32(rv))
		case float32:
			return reflect.ValueOf(uint32(rv))
		case float64:
			return reflect.ValueOf(uint32(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}

	case "uint64":
		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(uint64(rv))
		case int32:
			return reflect.ValueOf(uint64(rv))
		case int64:
			return reflect.ValueOf(uint64(rv))
		case uint:
			return reflect.ValueOf(uint64(rv))
		case uint32:
			return reflect.ValueOf(uint64(rv))
		case uint64:
			return reflect.ValueOf(rv)
		case float32:
			return reflect.ValueOf(uint64(rv))
		case float64:
			return reflect.ValueOf(uint64(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}
	case "float32":
		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(float32(rv))
		case int32:
			return reflect.ValueOf(float32(rv))
		case int64:
			return reflect.ValueOf(float32(rv))
		case uint:
			return reflect.ValueOf(float32(rv))
		case uint32:
			return reflect.ValueOf(float32(rv))
		case uint64:
			return reflect.ValueOf(float32(rv))
		case float32:
			return reflect.ValueOf(rv)
		case float64:
			return reflect.ValueOf(float32(rv))
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}
	case "float64":
		switch rv := v.(type) {
		case int:
			return reflect.ValueOf(float64(rv))
		case int32:
			return reflect.ValueOf(float64(rv))
		case int64:
			return reflect.ValueOf(float64(rv))
		case uint:
			return reflect.ValueOf(float64(rv))
		case uint32:
			return reflect.ValueOf(float64(rv))
		case uint64:
			return reflect.ValueOf(float64(rv))
		case float32:
			return reflect.ValueOf(float64(rv))
		case float64:
			return reflect.ValueOf(rv)
		default:
			panic(fmt.Errorf("%s, %s", rv, v))
		}
	case "string":
		panic("type is string")
	default:
		panic(fmt.Errorf("ValueType %s is not exists", vtype))
	}
}

func autoStrToValueByType(ft *fieldtag, fvalue reflect.Value) reflect.Value {
	//log.Println(fvalue.Kind(), reflect.String)
	if fvalue.Kind() != reflect.String {
		if fvalue.Kind() == reflect.Slice {
			var sel = 0
			if ft.MIndex != -1 {
				sel = ft.MIndex
			}
			if fvalue.Len() > 0 {
				return autoValueType(ft.VType, fvalue.Index(sel).Interface())
			}
			return reflect.New(ft.Type.Field(ft.Index).Type).Elem()
		}
		return autoValueType(ft.VType, fvalue.Interface())
	}

	switch ft.VType {
	case "int":
		v, err := strconv.ParseInt(fvalue.Interface().(string), 10, 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(int(v))
	case "int32":
		v, err := strconv.ParseInt(fvalue.Interface().(string), 10, 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(int32(v))
	case "int64":
		v, err := strconv.ParseInt(fvalue.Interface().(string), 10, 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(v)

	case "uint":
		v, err := strconv.ParseUint(fvalue.Interface().(string), 10, 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(uint(v))
	case "uint32":
		v, err := strconv.ParseUint(fvalue.Interface().(string), 10, 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(uint32(v))
	case "uint64":
		v, err := strconv.ParseUint(fvalue.Interface().(string), 10, 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(v)
	case "float32":
		v, err := strconv.ParseFloat(fvalue.Interface().(string), 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(float32(v))
	case "float64":
		v, err := strconv.ParseFloat(fvalue.Interface().(string), 64)
		if err != nil {
			log.Println(err)
		}
		return reflect.ValueOf(v)
	case "string":
		return fvalue
	default:
		log.Panic("ValueType ", ft.VType, "is not exists")
	}
	return fvalue
}

func callMethod(becall reflect.Value, method *methodtag) []reflect.Value {
	var callresult []reflect.Value
	if method.IsRegister { // call register function
		if becall.Kind() != reflect.String {
			becall = becall.MethodByName(DefaultMethod).Call(nil)[0] // call String()
		}
		callresult = []reflect.Value{becall}
		callresult = append(callresult, method.Args...)
		// var retcallresult []reflect.Value

		if mcall, ok := register[method.Method]; ok {
			return mcall.Call(callresult)
		}

		// 提示相似的函数. 防止写错自定义函数名字
		var maxpercent float64 = 0
		var curmehtod string
		for key := range register {
			percent := SimilarText(method.Method, key)
			if percent > maxpercent {
				maxpercent = percent
				curmehtod = key
			}
		}

		panic(fmt.Sprintf("Method name %s is not exists. please check it.\nMethod may be %s. the sim is %f", method.Method, curmehtod, maxpercent))
	}

	// call becall default method
	bymethod := becall.MethodByName(method.Method)
	if bymethod.IsValid() {
		callresult = bymethod.Call(method.Args)
		return callresult
	}

	log.Panicln(method.Method, "is not exists")
	return nil
}

func getInfoByTag(node *htmlquery.Node, fieldtags []*fieldtag, obj reflect.Value) {
	var ft *fieldtag
	defer func() {
		if err := recover(); err != nil {
			log.Panicf("err is %s\n fieldtags is %#v", err, ft)
		}
	}()

	for _, ft = range fieldtags {
		result, err := node.QueryAll(ft.Exp)
		if err == nil {
			if ft.Kind == reflect.Slice { // 如果是Slice 就返回Slice
				var callresults [][]reflect.Value
				for _, n := range result {
					becall := reflect.ValueOf(n)
					var isVaild = true
					var callresult []reflect.Value
					for _, method := range ft.Methods {
						if !becall.IsNil() {
							callresult = callMethod(becall, &method)
							becall = callresult[0]
						} else {
							isVaild = false
							break
						}
					}

					if isVaild {
						callresults = append(callresults, callresult)
					}
				}

				if len(callresults) > 0 {
					fvalue := obj.Field(ft.Index)
					for _, callresult := range callresults {
						fvalue = reflect.Append(fvalue, autoStrToValueByType(ft, callresult[0]))
					}
					obj.Field(ft.Index).Set(fvalue)
				}

			} else {

				if len(result) > 0 {
					var selResult *htmlquery.Node
					if ft.VIndex != -1 {
						selResult = result[ft.VIndex]
					} else {
						selResult = result[0]
					}

					var isVaild = true
					becall := reflect.ValueOf(selResult)
					var callresult []reflect.Value
					for _, method := range ft.Methods {
						if !becall.IsNil() {
							callresult = callMethod(becall, &method)
							becall = callresult[0]
						} else {
							isVaild = false
							break
						}

						if isVaild {
							fvalue := callresult[0]
							obj.Field(ft.Index).Set(autoStrToValueByType(ft, fvalue))
						}
					}
				}
			}
		}
	}
}

// ForEachObjectByTag after every result executing xpath, get the String of all result
func (xp *XPath) ForEachObjectByTag(obj interface{}) {
	// oslice := reflect.ValueOf(obj)
	ov := reflect.ValueOf(obj)
	oslice := ov.Elem()
	otype := oslice.Type().Elem()
	// log.Println(oslice, otype)
	var fieldtags []*fieldtag
	var isTypePtr bool = false
	if otype.Kind() == reflect.Ptr {
		otype = otype.Elem()
		isTypePtr = true
	}

	fieldtags = getFieldTags(otype)
	for _, xpresult := range xp.results {
		o := reflect.New(otype).Elem()
		getInfoByTag(xpresult, fieldtags, o)
		if isTypePtr {
			oslice = reflect.Append(oslice, o.Addr())
		} else {
			oslice = reflect.Append(oslice, o)
		}

	}

	ov.Elem().Set(oslice)
}

// ForEachTagName after every result executing xpath, get the String of all result
func (xp *XPath) ForEachTagName(exp string) (sstr []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		if txt, err := node.TagName(); err == nil {
			return txt
		}
		return nil
	})

	for _, i := range inames {
		sstr = append(sstr, i.(string))
	}

	return sstr, errlist
}

// ForEachString after every result executing xpath, get the String of all result
func (xp *XPath) ForEachString(exp string) (sstr []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		return node.OutputHTML(true)
	})

	for _, i := range inames {
		sstr = append(sstr, i.(string))
	}

	return sstr, errlist
}

// ForEachText after every result executing xpath, get the TextContent of all result
func (xp *XPath) ForEachText(exp string) (texts []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		return node.Text()
	})

	for _, i := range inames {
		texts = append(texts, i.(string))
	}

	return texts, errlist
}

// ForEachAttr after every result executing xpath, get the Attributes of all result
func (xp *XPath) ForEachAttr(exp string) (attributes []*htmlquery.Attribute, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		return node.Attributes()
	})

	for _, i := range inames {
		for _, attr := range i.([]*htmlquery.Attribute) {
			attributes = append(attributes, attr)
		}
	}

	return attributes, errlist
}

// ForEachAttrKeys after every result executing xpath, get the Attribute Key of all result
func (xp *XPath) ForEachAttrKeys(exp string) (keyslist []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		var ir []string
		for _, attr := range node.Attributes() {
			if attr != nil {
				ir = append(ir, attr.GetKey())
			}
		}
		return ir
	})

	for _, i := range inames {
		keyslist = append(keyslist, i.([]string)...)
	}

	return keyslist, errlist
}

// ForEachAttrValue after every result executing xpath, get the Attribute Value of all result
func (xp *XPath) ForEachAttrValue(exp string, attributes ...string) (values []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		var ir []string

		for _, attr := range attributes {
			attribute := node.GetAttributeByKey(attr)
			if attribute != nil {
				ir = append(ir, attribute.GetValue())
			}
		}

		return ir
	})

	for _, i := range inames {
		for _, s := range i.([]string) {
			values = append(values, s)
		}
	}

	return values, errlist
}

// ForEachEx foreach after every result executing xpath do funciton. note: duplicate
func (xp *XPath) ForEachEx(exp string, do func(*htmlquery.Node) interface{}) (values []interface{}, errorlist []error) {
	if len(xp.results) == 0 {
		return
	}

	var dict map[uintptr]*htmlquery.Node = make(map[uintptr]*htmlquery.Node)
	for _, xpresult := range xp.results {

		result, err := xpresult.QueryAll(exp)
		var inodes []*htmlquery.Node
		for _, qnode := range result {
			inodes = append(inodes, qnode)
		}

		if err != nil {
			if xp.errorFlags == ErrorSkip {
				errorlist = append(errorlist, err)
			} else {
				break
			}
		}

		for _, n := range inodes {

			dict[reflect.ValueOf(n).Pointer()] = n
		}

	}

	for _, in := range dict {
		if want := do(in); want != nil {
			values = append(values, want)
		}
	}

	return
}

// ForEach new XPath( every result xpath get results ). note: not duplicate
func (xp *XPath) ForEach(exp string) (newxpath *XPath, errorlist []error) {
	if len(xp.results) == 0 {
		return
	}

	var results []*htmlquery.Node
	for _, xpresult := range xp.results {
		result, err := xpresult.QueryAll(exp)
		if err != nil {
			if xp.errorFlags == ErrorSkip {
				errorlist = append(errorlist, err)
			} else {
				break
			}
		}
		results = append(results, result...)
	}
	newxpath = newXPath(results...)
	return
}
