package function

import (
	"text/template"
)

func FuncMap() template.FuncMap {
	return template.FuncMap(GenericFuncMap())
}

func FuncMapMerge(fma, fmb template.FuncMap) template.FuncMap {
	for k, v := range fmb {
		fma[k] = v
	}
	return fma
}

// GenericFuncMap returns a copy of the basic function map as a map[string]interface{}.
func GenericFuncMap() map[string]interface{} {
	gfm := make(map[string]interface{}, len(genericMap))
	for k, v := range genericMap {
		gfm[k] = v
	}
	return gfm
}

var genericMap = map[string]interface{}{
	"YFilter":  YFilter,
	"YPipe":    YPipe,
	"YValue":   YValue,
	"KFilter":  KFilter,
	"KPipe":    KPipe,
	"KYFilter": NewKYFilter,
}
