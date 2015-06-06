package engine

type Context struct {
	Configuration
	*Engine
}

func NewContext(configFile string) *Context {
	instance := &Context{}
	config := NewConfiguration(configFile)

	instance.Configuration = config
	instance.Engine = NewEngine()
	return instance
}
