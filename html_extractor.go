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

	"github.com/474420502/htmlquery"
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
func (etor *HmtlExtractor) GetObjectByTag(obj interface{}) interface{} {
	if nobj, ok := getInfoByTag(etor.doc, getFieldTags(obj)); ok {
		return nobj.Addr().Interface()
	}
	return nil
}

// XPaths multi xpath extractor
func (etor *HmtlExtractor) XPaths(exp string) (*XPath, error) {
	result, err := etor.doc.QueryAll(exp)
	return newXPath(result...), err
}

// XPath libxml2 xpathresult
func (etor *HmtlExtractor) XPath(exp string) (result *htmlquery.Node, err error) {
	n, err := etor.doc.Query(exp)
	return (*htmlquery.Node)(n), err
}

type ErrorFlags int

const (
	ERROR_BREAK ErrorFlags = 0
	ERROR_SKIP  ErrorFlags = 1
)

// XPath for easy extractor data
type XPath struct {
	results    []*htmlquery.Node
	errorFlags ErrorFlags
}

func newXPath(result ...*htmlquery.Node) *XPath {
	xp := &XPath{results: result, errorFlags: ERROR_SKIP}
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
	IsRegister bool
	Method     string
	Args       []reflect.Value
}

type fieldtag struct {
	Type   reflect.Type
	Kind   reflect.Kind
	VType  string
	VIndex int // exp results selected index
	MIndex int // method results selected index
	Index  int
	Exp    string
	// Method string
	// Args   []reflect.Value
	Methods []methodtag
}

// DefaultMethod 默认函数 如果tag没写mth(method) 的标识. 默认就是call Text()
var DefaultMethod = "Text"

var methodDict map[string]string

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

func getFieldTags(obj interface{}) []*fieldtag {
	otype := reflect.TypeOf(obj)
	var fieldtags []*fieldtag
	for i := 0; i < otype.NumField(); i++ {

		f := otype.Field(i)
		if exp, ok := f.Tag.Lookup("exp"); ok {
			ft := &fieldtag{}
			ft.Index = i
			ft.Exp = exp
			ft.Kind = f.Type.Kind()
			ft.Type = otype

			var smethod string
			var ok bool
			for _, mth := range []string{"method", "mth"} {
				if smethod, ok = f.Tag.Lookup(mth); ok {
					for _, method := range strings.Split(smethod, " ") {
						methodAndArgs := strings.Split(method, ",")
						mt := methodtag{}
						mt.Method = methodAndArgs[0]

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

			if index, ok := f.Tag.Lookup("index"); ok {
				i, err := strconv.Atoi(index)
				if err != nil {
					log.Panic(err)
				}
				ft.VIndex = i
			} else {
				ft.VIndex = -1
			}

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
			return autoValueType(ft.VType, fvalue.Index(sel).Interface())
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

func callMehtod(becall reflect.Value, method *methodtag) []reflect.Value {
	var callresult []reflect.Value
	if method.IsRegister { // call register function
		if becall.Kind() != reflect.String {
			becall = becall.MethodByName(DefaultMethod).Call(nil)[0] // call String()
		}
		callresult = []reflect.Value{becall}
		callresult = append(callresult, method.Args...)
		// var retcallresult []reflect.Value
		return register[method.Method].Call(callresult)
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

func getInfoByTag(node *htmlquery.Node, fieldtags []*fieldtag) (createobj reflect.Value, isCreateObj bool) {

	for _, ft := range fieldtags {
		result, err := node.QueryAll(ft.Exp)
		if err == nil {
			if ft.Kind == reflect.Slice {
				var callresults [][]reflect.Value
				for _, n := range result {
					becall := reflect.ValueOf(n)
					var isVaild = true
					var callresult []reflect.Value
					for _, method := range ft.Methods {
						if !becall.IsNil() {
							callresult = callMehtod(becall, &method)
							becall = callresult[0]
						} else {
							isVaild = false
							break
						}
					}

					if isVaild {
						callresults = append(callresults, callresult)
						if !isCreateObj {
							isCreateObj = true
							createobj = reflect.New(ft.Type).Elem()
						}
					}
				}

				if isCreateObj {
					fvalue := createobj.Field(ft.Index)
					for _, callresult := range callresults {
						fvalue = reflect.Append(fvalue, autoStrToValueByType(ft, callresult[0]))
					}
					createobj.Field(ft.Index).Set(fvalue)
				}

			} else {

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
						callresult = callMehtod(becall, &method)
						becall = callresult[0]
					} else {
						isVaild = false
						break
					}

					if isVaild {
						if !isCreateObj {
							isCreateObj = true
							createobj = reflect.New(ft.Type).Elem()
						}
						fvalue := callresult[0]
						createobj.Field(ft.Index).Set(autoStrToValueByType(ft, fvalue))
					}
				}
			}
		}
	}

	return
}

// ForEachObjectByTag after executing xpath, get the String of all result
func (xp *XPath) ForEachObjectByTag(obj interface{}) []interface{} {
	var results []interface{}
	fieldtags := getFieldTags(obj)
	for _, xpresult := range xp.results {
		if nobj, isCreateObj := getInfoByTag(xpresult, fieldtags); isCreateObj {
			results = append(results, nobj.Addr().Interface())
		}

	}

	return results
}

// ForEachTagName after executing xpath, get the String of all result
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

// ForEachString after executing xpath, get the String of all result
func (xp *XPath) ForEachString(exp string) (sstr []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		return node.OutputHTML(true)
	})

	for _, i := range inames {
		sstr = append(sstr, i.(string))
	}

	return sstr, errlist
}

// ForEachText after executing xpath, get the TextContent of all result
func (xp *XPath) ForEachText(exp string) (texts []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node *htmlquery.Node) interface{} {
		return node.Text()
	})

	for _, i := range inames {
		texts = append(texts, i.(string))
	}

	return texts, errlist
}

// ForEachAttr after executing xpath, get the Attributes of all result
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

// ForEachAttrKeys after executing xpath, get the Attribute Key of all result
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
		for _, s := range i.([]string) {
			keyslist = append(keyslist, s)
		}
	}

	return keyslist, errlist
}

// ForEachAttrValue after executing xpath, get the Attribute Value of all result
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

// ForEachEx foreach after executing xpath do funciton. note: duplicate
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
			if xp.errorFlags == ERROR_SKIP {
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
			if xp.errorFlags == ERROR_SKIP {
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
