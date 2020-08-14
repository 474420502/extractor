package extractor

import (
	"fmt"
	"testing"
)

type otag struct {
	Num  float64   `exp:"//div[@num]/@num" mth:"r:ParseNumber"`
	Nums []float64 `exp:"//div[@num]/@num" mth:"r:ParseNumber"`

	IntN  int32   `exp:"//div[@num]/@num" mth:"r:ParseNumber"`
	IntNs []int32 `exp:"//div[@num]/@num" mth:"r:ParseNumber"`

	Int64N  int64   `exp:"(//div[@num])[2]/@num" mth:"r:ParseNumber"`
	Int64Ns []int64 `exp:"//div[@num]/@num" mth:"r:ParseNumber"`

	Uint64N  uint64   `exp:"(//div[@num])[2]/@num" mth:"r:ParseNumber"`
	Uknt64Ns []uint64 `exp:"//div[@num]/@num" mth:"r:ParseNumber"`
}

func TestFunc(t *testing.T) {
	etor := ExtractHtmlString(`<html>
	<head></head>
	<body>
		<div class="red" num="123,123k">
			<a href="https://www.baidu.com"></a>
		</div>
		<div class="blue" num="3,000">
			<a href="https://www.google.com"></a>
		</div>

		<div class="black" num="456"> 
			<span>
				good你好
			</span>
		</div>
	</body>
</html>`)

	o := etor.GetObjectByTag(otag{}).(*otag)
	if o.Num != 123123000.0 {
		t.Error(o)
	}

	if len(o.Nums) != 3 {
		t.Error(o.Nums)
	}

	if fmt.Sprintf("%#v", o.Nums) != "[]float64{1.23123e+08, 3000, 456}" {
		t.Error(fmt.Sprintf("%#v", o.Nums))
	}

	if o.IntN != 123123000 {
		t.Error(o)
	}

	if fmt.Sprintf("%#v", o.IntNs) != "[]int32{123123000, 3000, 456}" {
		t.Error(fmt.Sprintf("%#v", o.IntNs))
	}

	if o.Int64N != 3000 {
		t.Error(o)
	}

	if fmt.Sprintf("%#v", o.Int64Ns) != "[]int64{123123000, 3000, 456}" {
		t.Error(fmt.Sprintf("%#v", o.Int64Ns))
	}
}

type enumbertag struct {
	Num  int   `exp:"//div[@num]/@num" mth:"r:ExtractNumber" index:"1" mindex:"1"`
	Nums []int `exp:"//div[@num]/@num" mth:"r:ExtractNumber"`
}

func TestExtractNumber(t *testing.T) {
	etor := ExtractHtmlString(`<html>
	<head></head>
	<body>
		<div class="red" num="3.3k">
			<a href="https://www.baidu.com"></a>
		</div>
		<div class="blue" num="hello 3,003k 20k">
			<a href="https://www.google.com"></a>
		</div>

		<div class="black" num="sd 456"> 
			<span>
				good你好
			</span>
		</div>
	</body>
</html>`)

	o := etor.GetObjectByTag(enumbertag{}).(*enumbertag)
	if o.Num != 20000 {
		t.Error(o.Num)
	}

	if fmt.Sprintf("%#v", o.Nums) != "[]int{3300, 3003000, 456}" {
		t.Error(fmt.Sprintf("%#v", o.Nums))
	}
}
