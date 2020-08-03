package extractor

import "github.com/tidwall/gjson"

type JsonExtractor struct {
	content []byte
	result  gjson.Result
}
