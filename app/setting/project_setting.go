package setting

type ProjectSettings struct {
	Name string
	Path string
	Prefix string
	APIs   []APISettings
}

type APISettings struct {
	API      string
	ResponseValue interface{}
	ResponseRef string
}
