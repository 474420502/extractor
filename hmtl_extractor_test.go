package extractor

import (
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestRegexp(t *testing.T) {
	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	etor := ExtractHtmlReader(f)
	for _, matches := range etor.RegexpString(`use xlink:href[^>]+width=\"(\d+)\"[^>]+height=\"(\d+)\"`) {
		if len(matches) != 3 {
			t.Error("may error match")
		}
	}
}

func TestXPathMethod(t *testing.T) {
	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	etor := ExtractHtmlReader(f)
	xp, err := etor.XPaths("//li")
	if err != nil {
		t.Error(err)
	}
	as, errs := xp.ForEach(".//a")
	if len(errs) > 0 {
		t.Error(errs)
	}

	for _, a := range as.GetTagNames() {
		if a != "a" {
			t.Error(a)
		}
	}

	for _, a := range as.GetAttributes("role") {
		if a.GetValue() != "button" {
			t.Error(a)
		}
	}
}

func TestHtml(t *testing.T) {

	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	etor := ExtractHtmlReader(f)
	xp, err := etor.XPaths("//*[contains(@class, 'c-header__modal__content__login')]")
	if err != nil {
		t.Error(err)
	}

	if len(xp.GetAttributes("class")) <= 10 {
		t.Error("class count is error")
	}

	xp, err = etor.XPaths("//dt")
	if err != nil {
		t.Error(err)
	}

	if len(xp.GetTexts()) != 5 {
		t.Error(xp.GetTexts())
	}

	if len(xp.GetStrings()) != 5 {
		t.Error(xp.GetStrings())
	}

	if fmt.Sprint(xp.GetTagNames()) != "[dt dt dt dt dt]" {
		t.Error(xp.GetTagNames())
	}

	if fmt.Sprint(xp.GetTexts()) != "[Trends History 関連ゲームタイトル 関連チャンネル 関連キーワード]" {
		t.Error(xp.GetTexts())
	}

	xp, err = etor.XPaths("//*[contains(@class, 'l-headerMain__search__pcContent')]")
	if err != nil {
		t.Error(err)
	}

	if len(xp.GetAttrKeysByValue("js-afterLogin l-headerMain__search__pcContent  ")) != 1 {
		t.Error("len should be 1")
		t.Error(xp.GetStrings())
	}

	if len(xp.GetAttrValuesByKey("class")) != 7 {
		t.Error("len should be 7", len(xp.GetAttrValuesByKey("class")))
		t.Error(xp.GetStrings())
	}

	if txts, errs := xp.ForEachText(".//dt"); len(errs) > 0 {
		t.Error(errs)
	} else {
		if len(txts) != 5 {
			t.Error(len(txts), txts)
		}
	}

	xp, err = etor.XPaths("//li")
	if err != nil {
		t.Error(err)
	}

	if txts, errs := xp.ForEachAttrKeys(".//a[@role='button']"); len(errs) > 0 {
		t.Error(errs)
	} else {
		if len(txts) <= 32 {
			t.Error(len(txts), txts)
		}
	}

	if txts, errs := xp.ForEachAttr(".//a[@role='button']"); len(errs) > 0 {
		t.Error(errs)
	} else {
		if len(txts) <= 32 {
			t.Error(len(txts), txts)
		}
	}

	if txts, errs := xp.ForEachAttrValue(".//a[@role='button']", "role"); len(errs) > 0 {
		t.Error(errs)
	} else {
		if len(txts) != 34 {
			t.Error(len(txts), txts)
		}
	}

	if txts, errs := xp.ForEachString(".//a[@role='button']"); len(errs) > 0 {
		t.Error(errs)
	} else {
		if len(txts) != 34 {
			t.Error(len(txts), txts)
		}
	}

	if txts, errs := xp.ForEachTagName(".//a[@role='button']"); len(errs) > 0 {
		t.Error(errs)
	} else {
		if len(txts) != 34 {
			t.Error(len(txts), txts)
		}

		for _, a := range txts {
			if a != "a" {
				t.Error(a)
			}
		}
	}

	// role

	// t.Error(xp.ForEachText(".//dt"))
}

// 测试的object
type toject struct {
	Li  string   `exp:".//li" method:"String"`
	Use []string `exp:".//use" method:"AttributeValue,width"`
}

func TestTag(t *testing.T) {
	// obj := &toject{}
	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}

	etor := ExtractHtmlReader(f)
	xp, err := etor.XPaths("//body")
	results := xp.ForEachObjectByTag(toject{})

	for _, r := range results {
		if len(r.(*toject).Use) != 54 {
			t.Error("len != 54, len is", len(r.(*toject).Use))
			l := r.(*toject).Use
			t.Errorf("%v", l)
		}
	}
}

type tagObject1 struct {
	Color string `exp:".//div[2]" method:"AttributeValue,class"`
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
	etor := ExtractHtmlString(`<html>
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
		tobj2 := xp.ForEachObjectByTag(tagObject2{})
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
		tag3 := xp.ForEachObjectByTag(tagObject3{})
		sr := spew.Sprint(tag3)
		if sr != `[<*>{red } <*>{blue }]` {
			t.Error(sr)
		}
	}

}

type tagObject4 struct {
	Num    int     `exp:"//div[@num]" mth:"AttrValue,num" `
	Num321 float64 `exp:"//div[@num]" mth:"AttrValue,num" index:"1"`
	Numstr string  `exp:"//div[@num]" mth:"AttrValue,num"`
	Nums   []int32 `exp:"//div[@num]" mth:"AttrValue,num"`
}

func TestType(t *testing.T) {
	etor := ExtractHtmlString(`<html>
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
