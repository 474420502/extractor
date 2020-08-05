package extractor

import (
	"bytes"
	"io"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/474420502/libxml2"
	"github.com/474420502/libxml2/clib"
	"github.com/474420502/libxml2/parser"
	"github.com/474420502/libxml2/types"
	"github.com/pkg/errors"
)

// XmlExtractor 提取器
type XmlExtractor struct {
	content []byte
	doc     types.Document
}

// ExtractXmlString extractor xml(html)
func ExtractXmlString(content string, options ...parser.HTMLOption) *XmlExtractor {
	c := []byte(content)
	doc, err := libxml2.ParseHTML(c, options...)
	if err != nil {
		panic(err)
	}
	e := &XmlExtractor{}
	e.doc = doc
	e.content = c
	return e
}

// ExtractXml extractor xml(html)
func ExtractXml(content []byte, options ...parser.HTMLOption) *XmlExtractor {
	doc, err := libxml2.ParseHTML(content, options...)
	if err != nil {
		panic(err)
	}
	e := &XmlExtractor{}
	e.doc = doc
	e.content = content
	return e
}

// ExtractXmlReader extractor xml(html)
func ExtractXmlReader(in io.Reader, options ...parser.HTMLOption) *XmlExtractor {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(in); err != nil {
		panic(errors.Wrap(err, "failed to rea from io.Reader"))
	}
	return ExtractXml(buf.Bytes(), options...)
}

// RegexpBytes multi xpath extractor
func (etor *XmlExtractor) RegexpBytes(exp string) [][][]byte {
	return regexp.MustCompile(exp).FindAllSubmatch(etor.content, -1)
}

// RegexpString multi xpath extractor
func (etor *XmlExtractor) RegexpString(exp string) [][]string {
	return regexp.MustCompile(exp).FindAllStringSubmatch(string(etor.content), -1)
}

// XPaths multi xpath extractor
func (etor *XmlExtractor) XPaths(exp string) (*XPath, error) {
	result, err := etor.doc.Find(exp)
	return NewXPath([]types.XPathResult{result}), err
}

// XPath libxml2 xpathresult
func (etor *XmlExtractor) XPath(exp string) (result types.XPathResult, err error) {
	result, err = etor.doc.Find(exp)
	return
}

type ErrorFlags int

const (
	ERROR_BREAK ErrorFlags = 0
	ERROR_SKIP  ErrorFlags = 1
)

// XPath for easy extractor data
type XPath struct {
	results    []types.XPathResult
	errorFlags ErrorFlags
}

func NewXPath(result []types.XPathResult) *XPath {
	xp := &XPath{results: result, errorFlags: ERROR_SKIP}
	return xp
}

// GetXPathResults Get Current XPath Results
func (xp *XPath) GetXPathResults() []types.XPathResult {
	return xp.results
}

// GetAttributes Get the Attribute of the Current XPath Results
func (xp *XPath) GetAttributes(key string) []types.Attribute {
	if len(xp.results) == 0 {
		return nil
	}

	var attrs []types.Attribute
	for _, xpresult := range xp.results {
		iter := xpresult.NodeIter()
		for iter.Next() {
			ele := iter.Node().(types.Element)
			if ele != nil {
				if attr, err := ele.GetAttribute(key); err == nil {
					attrs = append(attrs, attr)
				}
			}
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
		iter := xpresult.NodeIter()
		for iter.Next() {
			ele := iter.Node().(types.Element)
			if ele != nil {

				txts = append(txts, ele.TextContent())

			}
		}
	}
	return txts
}

// GetNodeValues Get the NodeValue of the Current XPath Results
func (xp *XPath) GetNodeValues() []string {
	if len(xp.results) == 0 {
		return nil
	}

	var nvalues []string
	for _, xpresult := range xp.results {
		iter := xpresult.NodeIter()
		for iter.Next() {
			ele := iter.Node().(types.Element)
			if ele != nil {
				nvalues = append(nvalues, ele.NodeValue())
			}
		}
	}

	return nvalues
}

// GetNodeNames Get the NodeName of the Current XPath Results
func (xp *XPath) GetNodeNames() []string {
	if len(xp.results) == 0 {
		return nil
	}

	var nvalues []string
	for _, xpresult := range xp.results {
		iter := xpresult.NodeIter()
		for iter.Next() {
			ele := iter.Node().(types.Element)
			if ele != nil {
				nvalues = append(nvalues, ele.NodeName())
			}
		}
	}

	return nvalues
}

// GetNodeStrings Get the String of the Current XPath Results
func (xp *XPath) GetNodeStrings() []string {
	if len(xp.results) == 0 {
		return nil
	}

	var nvalues []string
	for _, xpresult := range xp.results {
		iter := xpresult.NodeIter()
		for iter.Next() {
			ele := iter.Node().(types.Element)
			if ele != nil {
				nvalues = append(nvalues, ele.String())
			}
		}
	}

	return nvalues
}

// GetTypes Get the Type of the Current XPath Results
func (xp *XPath) GetTypes() []clib.XMLNodeType {
	if len(xp.results) == 0 {
		return nil
	}

	var txts []clib.XMLNodeType
	for _, xpresult := range xp.results {
		iter := xpresult.NodeIter()
		for iter.Next() {
			txts = append(txts, iter.Node().NodeType())
		}
	}
	return txts
}

type methodtag struct {
	Method string
	Args   []reflect.Value
}

type fieldtag struct {
	Type  reflect.Type
	Kind  reflect.Kind
	Index int
	Exp   string
	// Method string
	// Args   []reflect.Value
	Methods []methodtag
}

var methodDict map[string]string

type nodeMethod string

const (
	// TextContent Node.TextContent()
	TextContent nodeMethod = "TextContent"
	// GetAttribute Node.GetAttribute()
	GetAttribute nodeMethod = "GetAttribute"
	// NodeName Node.NodeName()
	NodeName nodeMethod = "NodeName"
	// NodeValue Node.NodeValue()
	NodeValue nodeMethod = "NodeValue"
	// ParentNode Node.NodeValue()
	ParentNode nodeMethod = "ParentNode"
)

func init() {
	methodDict = make(map[string]string)
	methodDict["Text"] = string(TextContent)
	methodDict["Attribute"] = string(GetAttribute)
	methodDict["Name"] = string(NodeName)
	methodDict["Parent"] = string(ParentNode)
	methodDict["Value"] = string(NodeValue)
}

// ForEachTag after executing xpath, get the String of all result
func (xp *XPath) ForEachTag(obj interface{}) []interface{} {
	otype := reflect.TypeOf(obj)
	var results []interface{}
	var fieldtags []*fieldtag
	for i := 0; i < otype.NumField(); i++ {

		f := otype.Field(i)
		if exp, ok := f.Tag.Lookup("exp"); ok {
			ft := &fieldtag{}
			ft.Index = i
			ft.Exp = exp
			ft.Kind = f.Type.Kind()
			ft.Type = otype
			if smethod, ok := f.Tag.Lookup("method"); ok {
				for _, method := range strings.Split(smethod, " ") {
					methodAndArgs := strings.Split(method, ",")
					mt := methodtag{}
					mt.Method = methodAndArgs[0]
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

			} else {
				mt := methodtag{}
				mt.Method = string(NodeName)
				if v, ok := methodDict[mt.Method]; ok {
					mt.Method = v
				}
				mt.Args = nil
				ft.Methods = append(ft.Methods, mt)
			}

			fieldtags = append(fieldtags, ft)
		}

	}

	// var dict map[uintptr]types.Node = make(map[uintptr]types.Node)
	for _, xpresult := range xp.results {

		iter := xpresult.NodeIter()
		for iter.Next() {
			node := iter.Node()
			var nobj reflect.Value
			var isCreateObj bool = false
			for _, ft := range fieldtags {
				result, err := node.Find(ft.Exp)
				// var inodes []types.Node
				if err == nil {

					iter := result.NodeIter()
					if ft.Kind == reflect.Slice {

						var callresults [][]reflect.Value
						for iter.Next() {
							becall := reflect.ValueOf(iter.Node())
							var isVaild = true
							var callresult []reflect.Value
							for _, method := range ft.Methods {
								if !becall.IsNil() {
									callresult = becall.MethodByName(method.Method).Call(method.Args)
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
									nobj = reflect.New(ft.Type).Elem()
								}
							}
						}

						if isCreateObj {
							fvalue := nobj.Field(ft.Index)
							for _, callcallresult := range callresults {
								fvalue = reflect.Append(fvalue, callcallresult[0])
							}
							nobj.Field(ft.Index).Set(fvalue)
						}

						// nobj.Elem().Field(ft.Index).Set(callresults[0])
					} else {

						if iter.Next() {

							var isVaild = true
							becall := reflect.ValueOf(iter.Node())
							var callresult []reflect.Value
							for _, method := range ft.Methods {
								if !becall.IsNil() {
									callresult = becall.MethodByName(method.Method).Call(method.Args)
									becall = callresult[0]
								} else {
									isVaild = false
									break
								}
							}

							if isVaild {
								if !isCreateObj {
									isCreateObj = true
									nobj = reflect.New(ft.Type).Elem()
								}
								nobj.Field(ft.Index).Set(callresult[0])
							}
						}
					}
				}
			}

			if isCreateObj {
				// var nobj = reflect.New(otype)
				results = append(results, nobj.Interface())
			}
		}
	}

	return results
}

// ForEachString after executing xpath, get the String of all result
func (xp *XPath) ForEachString(exp string) (sstr []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		return node.String()
	})

	for _, i := range inames {
		sstr = append(sstr, i.(string))
	}

	return sstr, errlist
}

// ForEachText after executing xpath, get the TextContent of all result
func (xp *XPath) ForEachText(exp string) (texts []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		return node.TextContent()
	})

	for _, i := range inames {
		texts = append(texts, i.(string))
	}

	return texts, errlist
}

// ForEachType after executing xpath, get the XMLNodeType of all result
func (xp *XPath) ForEachType(exp string) (typelist []clib.XMLNodeType, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		return node.NodeType()
	})

	for _, i := range inames {
		typelist = append(typelist, i.(clib.XMLNodeType))
	}

	return typelist, errlist
}

// ForEachValue after executing xpath, get the NodeValue of all result
func (xp *XPath) ForEachValue(exp string) (values []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		return node.NodeValue()
	})

	for _, i := range inames {
		values = append(values, i.(string))
	}

	return values, errlist
}

// ForEachAttr after executing xpath, get the Attributes of all result
func (xp *XPath) ForEachAttr(exp string) (attributes []types.Attribute, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		ele := node.(types.Element)
		attrs, err := ele.Attributes()
		if err != nil {
			log.Println(err)
		} else {
			return attrs
		}
		return nil
	})

	for _, i := range inames {
		for _, attr := range i.([]types.Attribute) {
			attributes = append(attributes, attr)
		}
	}

	return attributes, errlist
}

// ForEachAttrKeys after executing xpath, get the Attribute Key of all result
func (xp *XPath) ForEachAttrKeys(exp string) (keyslist []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		var ir []string

		ele := node.(types.Element)
		attributes, err := ele.Attributes()
		if err != nil {
			log.Println(err)
		} else {
			for _, attr := range attributes {
				ir = append(ir, attr.NodeName())
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

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		var ir []string

		ele := node.(types.Element)
		for _, attr := range attributes {
			attribute, err := ele.GetAttribute(attr)
			if err != nil {
				log.Println(err)
			} else {
				ir = append(ir, attribute.Value())
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

// ForEachName after executing xpath, get the NodeName of all result
func (xp *XPath) ForEachName(exp string) (names []string, errorlist []error) {

	inames, errlist := xp.ForEachEx(exp, func(node types.Node) interface{} {
		return node.NodeName()
	})

	for _, i := range inames {
		names = append(names, i.(string))
	}

	return names, errlist
}

// ForEachEx foreach after executing xpath do funciton. note: duplicate
func (xp *XPath) ForEachEx(exp string, do func(types.Node) interface{}) (values []interface{}, errorlist []error) {
	if len(xp.results) == 0 {
		return
	}

	var dict map[uintptr]types.Node = make(map[uintptr]types.Node)
	for _, xpresult := range xp.results {

		iter := xpresult.NodeIter()
		for iter.Next() {
			node := iter.Node()
			result, err := node.Find(exp)
			var inodes []types.Node
			for iter := result.NodeIter(); iter.Next(); {
				inodes = append(inodes, iter.Node())
			}

			if err != nil {
				if xp.errorFlags == ERROR_SKIP {
					errorlist = append(errorlist, err)
				} else {
					break
				}
			}

			for _, n := range inodes {
				dict[n.Pointer()] = n
			}

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

	var results []types.XPathResult

	for _, xpresult := range xp.results {

		iter := xpresult.NodeIter()
		for iter.Next() {
			node := iter.Node()
			result, err := node.Find(exp)
			if err != nil {
				if xp.errorFlags == ERROR_SKIP {
					errorlist = append(errorlist, err)
				} else {
					break
				}
			}
			results = append(results, result)
		}
	}

	newxpath = NewXPath(results)
	return
}
