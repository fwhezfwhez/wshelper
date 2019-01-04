package wshelper

import (
	"github.com/fsnotify/fsnotify"
	"github.com/fwhezfwhez/errorx"
	"github.com/spf13/viper"
	"log"
	"os"
	"sync"
)

var (
	logger *log.Logger
	v *viper.Viper
	m sync.RWMutex
)

func init() {
	v = viper.New()
	logFilePath := ""
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("../config/")
	v.AddConfigPath("config/")
	v.AddConfigPath(".")

	ReadConfig(v)

	logFilePath = v.GetString("logFilePath")
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		ReadConfig(v)
		logFilePath = v.GetString("logFilePath")
	})

	file, err := os.OpenFile("G:\\go_workspace\\GOPATH\\src\\eyas\\wshelper\\error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	//log.SetFlags(0)
	logger = log.New(file, "", log.LstdFlags|log.Llongfile)
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
