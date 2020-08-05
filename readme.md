# Contain XPath Regexp Tag(XPath), is easy to Extractor the data format

## Example

1. eg:

```golang
type toject struct {
	Li  string   `exp:".//li" method:"String"` // xpath -> []Node.String()
	Use []string `exp:".//use" method:"GetAttribute,width Value"` // xpath -> []Node.GetAttribute("width").NodeValue()
}

func TestTag(t *testing.T) {
	f, err := os.Open("./testfile/test1.html")
	if err != nil {
		t.Error(err)
	}
	
	etor := ExtractXmlReader(f)
	xp, err := etor.XPaths("//body")
    results := xp.ForEachTag(toject{})
    t.Log(results)
}
```


2. eg:
```golang
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

    xp.GetAttributes("class") // get all xpath result Attribute class
    xp.GetNodeNames() // get all NodeName
    xp.ForEachText(".//dt") // all xpath result execute XPath(.//dt) Get all Text

    etor.XPath("{xpath}") // one result like libxml2 you can see https://github.com/lestrrat-go/libxml2
}
```