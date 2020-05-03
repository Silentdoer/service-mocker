package setting

type ProjectSettings struct {
	Name string `json:"name"`  // 序列化时Name以JSON name属性名输出，反序列化则本身就是支持JSON name属性转换为Name字段
	Path string `json:"path"`
	Prefix string `json:"prefix"`
	APIs   []APISettings
}

type APISettings struct {
	API      string
	ResponseValue interface{}
	ResponseRef string  // `json:"-"`  // 忽略字段，最好不要用，因为它是既不序列化，也不反序列化
}
