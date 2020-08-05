package extractor

import (
	"fmt"
	"os"
	"testing"
)

func TestRegexp(t *testing.T) {
	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	etor := ExtractXmlReader(f)
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
	etor := ExtractXmlReader(f)
	xp, err := etor.XPaths("//li")
	if err != nil {
		t.Error(err)
	}
	as, errs := xp.ForEach(".//a")
	if len(errs) > 0 {
		t.Error(errs)
	}

	for _, a := range as.GetNodeNames() {
		if a != "a" {
			t.Error(a)
		}
	}

	for _, a := range as.GetAttributes("role") {
		if a.NodeValue() != "button" {
			t.Error(a)
		}
	}
}

func TestHtml(t *testing.T) {

	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	etor := ExtractXmlReader(f)
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

	if len(xp.GetNodeStrings()) != 5 {
		t.Error(xp.GetNodeStrings())
	}

	if fmt.Sprint(xp.GetNodeNames()) != "[dt dt dt dt dt]" {
		t.Error(xp.GetNodeNames())
	}

	if fmt.Sprint(xp.GetNodeValues()) != "[Trends History 関連ゲームタイトル 関連チャンネル 関連キーワード]" {
		t.Error(xp.GetNodeValues())
	}

	xp, err = etor.XPaths("//*[contains(@class, 'l-headerMain__search__pcContent')]")
	if err != nil {
		t.Error(err)
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

	if txts, errs := xp.ForEachName(".//a[@role='button']"); len(errs) > 0 {
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
	Use []string `exp:".//use" method:"GetAttribute,width Value"`
}

func TestTag(t *testing.T) {
	// obj := &toject{}
	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}

	etor := ExtractXmlReader(f)
	xp, err := etor.XPaths("//body")
	results := xp.ForEachTag(toject{})

	for _, r := range results {
		if len(r.(*toject).Use) != 46 {
			t.Error("len != 46")
		}
	}

	// objvalue := reflect.ValueOf(obj).Elem()
	// objtype := reflect.TypeOf(obj).Elem()

	// for i := 0; i < objtype.NumField(); i++ {
	// 	f := objtype.Field(i)
	// 	v := objvalue.Field(i)

	// 	if exp, ok := f.Tag.Lookup("exp"); ok {
	// 		if method, ok := f.Tag.Lookup("method"); ok {
	// 			t.Error(exp, method)
	// 		}
	// 		if !v.CanSet() {
	// 			t.Error(f.Name, " the field is not can set. must use uppercase")
	// 		} else {
	// 			objvalue.Field(i).Set(reflect.ValueOf(exp))
	// 		}

	// 	}
	// }

	// t.Error(obj)
}
