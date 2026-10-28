package v1alpha1

// Ptr returns a pointer to b.
func Ptr(b bool) *bool {
	v := b
	return &v
}

// BoolOrDefault returns *p when set, otherwise def.
func BoolOrDefault(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// Int32Ptr returns a pointer to i.
func Int32Ptr(i int32) *int32 {
	v := i
	return &v
}

// Int32OrDefault returns *p when set, otherwise def.
func Int32OrDefault(p *int32, def int32) int32 {
	if p == nil {
		return def
	}
	return *p
}
