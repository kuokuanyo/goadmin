package captcha

type Captcha interface {
	Validate(token string) bool
}

var List = make(map[string]Captcha)

// 將參數key、captcha加入List(make(map[string]Captcha))
func Add(key string, captcha Captcha) {
	if _, exist := List[key]; exist {
		panic("captcha exist")
	}
	List[key] = captcha
}

// 判斷List(make(map[string]Captcha))裡是否有參數key的值並回傳Captcha(interface)
func Get(key string) (Captcha, bool) {
	captcha, ok := List[key]
	return captcha, ok
}
