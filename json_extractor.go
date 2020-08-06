package extractor

import "github.com/tidwall/gjson"

type JsonExtractor struct {
	result gjson.Result
}

// EtractorJson 提取json
func EtractorJson(content string) *JsonExtractor {
	etor := &JsonExtractor{}
	etor.result = gjson.Parse(content)
	return etor
}

// EtractorJsonBytes 提取json
func EtractorJsonBytes(content []byte) *JsonExtractor {
	return EtractorJson(string(content))
}

func (etor *JsonExtractor) GetObjectByTags(obj interface{}) interface{} {
	return nil
}
