package runtime

import "aspire/are/main/lib"

var globalSymbolReg = lib.NewMap[*String, *Symbol]()

var InspectKey = NewString("customInspect")
var InspectSymbol = NewSymbol(InspectKey.string)
var ToStringTagKey = NewString("toStringTag")
var ToStringTagSymbol = NewSymbol(ToStringTagKey.string)
var ToPrimitiveTagKey = NewString("toPrimitive")
var ToPrimitiveTagSymbol = NewSymbol(ToPrimitiveTagKey.string)

func init() {
	globalSymbolReg.Set(InspectKey, InspectSymbol)
	globalSymbolReg.Set(ToStringTagKey, ToStringTagSymbol)
	globalSymbolReg.Set(ToPrimitiveTagKey, ToPrimitiveTagSymbol)
}
