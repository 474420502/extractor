package extractor

import (
	"os"
	"testing"
)

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
	t.Error(xp.ForEachAttr("//*[contains(@class, 'c-header__modal__content__login')]"))
	// t.Error(xp.GetXPathResults())
}
