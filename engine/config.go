package engine

import (
	"time"

	"github.com/pelletier/go-toml"
	"github.com/satori/go.uuid"
)

type Configuration struct {
	*toml.Tree
}

func NewConfiguration(filename string) Configuration {
	configuration := Configuration{}
	var err error
	configuration.Tree, err = toml.LoadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	return configuration
}

func (self *Configuration) String(key string, def string) string {
	return self.GetDefault(key, def).(string)
}

func (self *Configuration) Int(key string, def int) int {
	return self.GetDefault(key, def).(int)
}

func (self *Configuration) Int32(key string, def int32) int32 {
	return self.GetDefault(key, def).(int32)
}

func (self *Configuration) Int64(key string, def int64) int64 {
	return self.GetDefault(key, def).(int64)
}

func (self *Configuration) Duration(key string, def time.Duration) time.Duration {
	tmp := self.GetDefault(key, def).(int64)
	return time.Duration(tmp)
}

func (self *Configuration) Float32(key string, def float32) float32 {
	return self.GetDefault(key, def).(float32)
}

func (self *Configuration) Float64(key string, def float64) float64 {
	return self.GetDefault(key, def).(float64)
}

func (self *Configuration) Bool(key string, def bool) bool {
	return self.GetDefault(key, def).(bool)
}

func (self *Configuration) StringArray(key string, def []string) []string {
	tmp := self.GetDefault(key, def).([]interface{})
	array := make([]string, len(tmp))
	for i, value := range tmp {
		array[i] = value.(string)
	}
	return array
}

func (self *Configuration) Int64Array(key string, def []int64) []int64 {
	tmp := self.GetDefault(key, def).([]interface{})
	array := make([]int64, len(tmp))
	for i, value := range tmp {
		array[i] = value.(int64)
	}
	return array
}

func (self *Configuration) IntArray(key string, def []int) []int {
	tmp := self.GetDefault(key, def).([]interface{})
	array := make([]int, len(tmp))
	for i, value := range tmp {
		array[i] = value.(int)
	}
	return array
}

func (self *Configuration) GetRandomUUID() string {
	return uuid.NewV4().String()
}
