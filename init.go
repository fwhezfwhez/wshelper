package wshelper

import (
	"github.com/fsnotify/fsnotify"
	"github.com/fwhezfwhez/errorx"
	"github.com/spf13/viper"
	"sync"
)

var (
	v *viper.Viper
	m sync.RWMutex
)

func init() {
	v = viper.New()

	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("../config/")
	v.AddConfigPath("config/")
	v.AddConfigPath(".")

	ReadConfig(v)
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		ReadConfig(v)
	})
}

func ReadConfig(v *viper.Viper) error {
	m.Lock()
	defer m.Unlock()
	err := v.ReadInConfig()
	if err != nil {
		return errorx.NewFromStringf("Error on parsing config file! '%s'", err.Error())
	}
	return nil
}
