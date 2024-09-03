package configuration

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
)

type Configuration struct {
	*config.Config
	watchedFiles map[string]bool
}

func loadFromJsonFile(filePaths ...string) (*Configuration, error) {

	conf := config.New("ConfigJson")

	conf.WithOptions(config.ParseEnv)

	err := conf.LoadFiles(filePaths...)

	if err != nil {
		return nil, err
	}

	return &Configuration{conf, make(map[string]bool)}, nil
}

func loadFromYamlFile(filePaths ...string) (*Configuration, error) {

	conf := config.New("ConfigYaml")

	conf.AddDriver(yamlv3.Driver)

	conf.WithOptions(config.ParseEnv)

	err := conf.LoadFiles(filePaths...)

	if err != nil {
		return nil, err
	}

	return &Configuration{conf, make(map[string]bool)}, nil
}

func LoadFromJsonFile(activeWatcher bool, filePaths ...string) error {

	conf, err := loadFromJsonFile(filePaths...)

	if err != nil {
		return err
	}

	appConfig = &AppConfig{conf, true}

	if activeWatcher {

		err = appConfig.Config.watchFiles()

		if err != nil {
			return err
		}
	}

	return nil
}

func LoadFromYamlFile(activeWatcher bool, filePaths ...string) error {

	conf, err := loadFromYamlFile(filePaths...)

	if err != nil {
		return err
	}

	appConfig = &AppConfig{conf, true}

	if activeWatcher {

		err = appConfig.Config.watchFiles()

		if err != nil {
			return err
		}
	}

	return nil

}

func getTypeValue[T any](key string, cfg *Configuration) (*T, error) {

	var value T

	err := cfg.BindStruct(key, &value)

	return &value, err
}

func (c *Configuration) watchFiles() error {

	files := c.LoadedFiles()

	if len(files) == 0 {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	go func(w *fsnotify.Watcher) {

		defer w.Close()

		for {
			select {

			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				if event.Has(fsnotify.Write) {
					c.ReloadFiles()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				if err != nil {
					panic(err)
				}

			}
		}
	}(watcher)

	for _, path := range files {

		if err := watcher.Add(path); err != nil {

			return err

		}
	}

	return nil

}

type AppConfig struct {
	Config       *Configuration
	IsBeenLoaded bool
}

var appConfig *AppConfig

func Get() *Configuration {

	if appConfig == nil ||
		!appConfig.IsBeenLoaded {
		panic("appConfig is not initialized")
	}
	return appConfig.Config
}

func GetSection[T any](section string) (*T, error) {
	return getTypeValue[T](section, Get())
}

func GetValue(key string) (any, bool) {
	return Get().GetValue(key)
}
