# Contain XPath Regexp Tag(XPath), is easy to Extractor the data format

## Example

1. eg:

```golang
type tagObject1 struct {
	Color string `exp:".//div[2]" method:"AttributeValue,class"` // exp is xpath expression and method is call xpath result(htmlquery.Node) method.   class is arg. mean GetAttribute().Value()
	Herf  string `exp:".//div[2]/a" method:"AttributeValue,href"`
}

type tagObject2 struct {
	Color string `exp:"self::div" mth:"AttrValue,class"`
	Herf  string `exp:".//a" mth:"AttributeValue,href"`
}

type tagObject3 struct {
	Color string `exp:"self::div/@class"`
	Herf  string `exp:".//a"`
}

func TestTag1(t *testing.T) {
	etor := extractor.ExtractHtmlString(`<html>
		<head></head>
		<body>
			<div class="red">
				<a href="https://www.baidu.com"></a>
			</div>
			<div class="blue">
				<a href="https://www.google.com"></a>
			</div>

			<div class="black"> 
				<span>
					good你好
				</span>
			</div>
		</body>
	</html>`)

	to := etor.GetObjectByTag(tagObject1{}).(*tagObject1)
	if to.Color != "blue" {
		t.Error(to)
	}

	if to.Herf != "https://www.google.com" {
		t.Error(to)
	}

	xp, err := etor.XPaths("//div/a/..")
	if err != nil {
		t.Error(err)
	} else {
		tobj2 := xp.ForEachTag(tagObject2{})
		if o := tobj2[0].(*tagObject2); o.Herf != "https://www.baidu.com" || o.Color != "red" {
			t.Error(o)
		}

		if o := tobj2[1].(*tagObject2); o.Herf != "https://www.google.com" || o.Color != "blue" {
			t.Error(o)
		}

	}

	xp, err = etor.XPaths("//div/a/..")
	if err != nil {
		t.Error(err)
	} else {
		tag3 := xp.ForEachTag(tagObject3{})
		sr := spew.Sprint(tag3)
		if sr != `[<*>{<class>red</class> <a href="https://www.baidu.com"></a>} <*>{<class>blue</class> <a href="https://www.google.com"></a>}]` {
			t.Error(sr)
		}
	}

}

type tagObject4 struct {
	Num    int     `exp:"//div[@num]" mth:"AttrValue,num" `
	Num321 float64 `exp:"//div[@num]" mth:"AttrValue,num" index:"1"` // get all div[index]
	Numstr string  `exp:"//div[@num]" mth:"AttrValue,num"`
	Nums   []int32 `exp:"//div[@num]" mth:"AttrValue,num"`
}

func TestType(t *testing.T) {
	etor := extractor.ExtractHtmlString(`<html>
	<head></head>
	<body>
		<div class="red" num="123">
			<a href="https://www.baidu.com"></a>
		</div>
		<div class="blue" num="321">
			<a href="https://www.google.com"></a>
		</div>

		<div class="black" num="456"> 
			<span>
				good你好
			</span>
		</div>
	</body>
</html>`)

	obj4 := etor.GetObjectByTag(tagObject4{}).(*tagObject4)

	if obj4.Num == 0 {
		t.Error("tag parse errror", obj4.Num)
	}

	if len(obj4.Nums) != 3 {
		t.Error("tag parse errror", obj4.Nums)
	}

	if obj4.Numstr != "123" {
		t.Error("tag parse errror", obj4.Numstr)
	}

	if obj4.Num321 != 321 {
		t.Error("tag parse errror", obj4.Num321)
	}
}

```


2. eg:
```golang
func TestHtml(t *testing.T) {

	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	etor := extractor.ExtractHtmlReader(f)
	xp, err := etor.XPaths("//*[contains(@class, 'c-header__modal__content__login')]")
	if err != nil {
		t.Error(err)
    }

    xp.GetAttributes("class") // get all xpath result Attribute class
    xp.ForEachText(".//dt") // all xpath result execute XPath(.//dt) Get all Text

    etor.XPath("{xpath}") // one result like html.Node. i change some api. https://github.com/474420502/htmlquery/blob/feature/esonapi/xnode.go  forked from antchfx/htmlquery
}
```