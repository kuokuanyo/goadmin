package form

import (
	"errors"
)

const (
	PostTypeKey           = "__go_admin_post_type"
	PostResultKey         = "__go_admin_post_result"
	PostIsSingleUpdateKey = "__go_admin_is_single_update"

	PreviousKey = "__go_admin_previous_"
	TokenKey    = "__go_admin_t_"
	MethodKey   = "__go_admin_method_"

	NoAnimationKey = "__go_admin_no_animation_"
)

// Values maps a string key to a list of values.
// It is typically used for query parameters and form values.
// Unlike in the http.Header map, the keys in a Values map
// are case-sensitive.
type Values map[string][]string

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
// 透過參數key判斷Values[key]長度是否大於0，如果大於零回傳Values[key][0]，反之回傳""
func (f Values) Get(key string) string {
	if len(f[key]) > 0 {
		return f[key][0]
	}
	return ""
}

// Add adds the value to key. It appends to any existing
// values associated with key.
// 將參數key、value加入Values(map[string][]string)中
func (f Values) Add(key string, value string) {
	f[key] = []string{value}
}

// IsEmpty check the key is empty or not.
func (f Values) IsEmpty(key ...string) bool {
	for _, k := range key {
		if f.Get(k) == "" {
			return true
		}
	}
	return false
}

// Has check the key exists or not.
func (f Values) Has(key ...string) bool {
	for _, k := range key {
		if f.Get(k) != "" {
			return true
		}
	}
	return false
}

// Delete deletes the values associated with key.
// 透過參數key刪除Values(map[string][]string)[key]
func (f Values) Delete(key string) {
	delete(f, key)
}

// ToMap turn the values to a map[string]string type.
func (f Values) ToMap() map[string]string {
	var m = make(map[string]string)
	for key, v := range f {
		if len(v) > 0 {
			m[key] = v[0]
		}
	}
	return m
}

// IsUpdatePost check the param if is from an update post request type or not.
func (f Values) IsUpdatePost() bool {
	return f.Get(PostTypeKey) == "0"
}

// IsInsertPost check the param if is from an insert post request type or not.
func (f Values) IsInsertPost() bool {
	return f.Get(PostTypeKey) == "1"
}

// PostError get the post result.
func (f Values) PostError() error {
	msg := f.Get(PostResultKey)
	if msg == "" {
		return nil
	}
	return errors.New(msg)
}

// IsSingleUpdatePost check the param if from an single update post request type or not.
func (f Values) IsSingleUpdatePost() bool {
	return f.Get(PostIsSingleUpdateKey) == "1"
}

// RemoveRemark removes the PostType and IsSingleUpdate flag parameters.
// 刪除__go_admin_post_type與__go_admin_is_single_update的鍵與值後回傳map[string][]string
func (f Values) RemoveRemark() Values {
	// PostTypeKey = __go_admin_post_type
	f.Delete(PostTypeKey)
	// PostIsSingleUpdateKey = __go_admin_is_single_update
	f.Delete(PostIsSingleUpdateKey)
	return f
}

// RemoveSysRemark removes all framework post flag parameters.
func (f Values) RemoveSysRemark() Values {
	f.Delete(PostTypeKey)
	f.Delete(PostIsSingleUpdateKey)
	f.Delete(PostResultKey)
	f.Delete(PreviousKey)
	f.Delete(TokenKey)
	f.Delete(MethodKey)
	f.Delete(NoAnimationKey)
	return f
}
