package runtime

import "aspire/are/main/lib"

func jsonValueFromAS(v Value) (obj any, err error) {
	switch v := v.(type) {
	case Number:
		return v, nil
	case *String:
		return v.string, nil
	case *Array:
		arr := []any{}
		v.elements.ForEach(func(_ int, v Value) {
			if err == nil {
				str, e := jsonValueFromAS(v)
				err = e
				arr = append(arr, str)
			}
		})
		return arr, err
	case *Object:
		m := make(map[string]any)
		v.own.ForEach(func(key *String, pd PropertyDescriptor, _ bool) {
			if err == nil {
				m[key.string], err = jsonValueFromAS(pd.value)
			}
		})
		return m, err
	case *RAW:
		return v.value, nil
	}
	return "{}", err
}

func JSONParse(str string) Value {
	var v any
	err := lib.JSONDecode([]byte(str), &v)
	if err != nil {
		return NewString("[Invalid JSON: " + err.Error() + "]")
	}
	return ASValueFromJSON(v)
}

func ASValueFromJSON(v any) Value {
	if v == nil {
		return null
	}
	switch v := v.(type) {
	case bool:
		return Boolean(v)
	case string:
		return NewString(v)
	case float64:
		return Number(v)
	case []any:
		arr := NewArray()
		for _, el := range v {
			arr.push(ASValueFromJSON(el))
		}
		return arr
	case map[string]any:
		obj := NewObject()
		for k, p := range v {
			obj.own.Set(NewString(k), DefaultPropDesc(ASValueFromJSON(p)))
		}
		return obj
	}
	return NewString("[Unhandled JSON]")
}

func JSONStringify(v Value, indent bool) (Value, error) {
	switch v := v.(type) {
	case *Object:
		if v == nil {
			return NewString("null"), nil
		}
		m := make(map[string]any)
		var err error
		v.own.ForEach(func(key *String, pd PropertyDescriptor, _ bool) {
			if err == nil {
				d, e := jsonValueFromAS(pd.value)
				err = e
				m[key.string] = d
			}
		})
		if err != nil {
			return NewString("{}"), err
		}
		b, err := lib.JSONEncode(m, indent)
		return NewString(string(b)), err
	case *Array:
		arr, err := jsonValueFromAS(v)
		if err != nil {
			return undefined, err
		}
		b, err := lib.JSONEncode(arr, indent)
		return NewString(string(b)), err
	case Number, *String, Boolean:
		b, err := lib.JSONEncode(v, indent)
		return NewString(string(b)), err
	case *RAW:
		b, err := lib.JSONEncode(v.value, indent)
		if err != nil {
			return undefined, err
		}
		return NewString(string(b)), err
	}
	return undefined, nil
}
