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
