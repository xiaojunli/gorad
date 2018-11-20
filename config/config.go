package config

import (
	"github.com/Unknwon/goconfig"
	"fmt"
	)

var config *goconfig.ConfigFile

func init () {
	config = loadConfigFile("config/conf.ini")
}

/**
 *  获取带默认值的配置值
 */
func MustInt(path string, defaultValue int) (int) {
	return config.MustInt(goconfig.DEFAULT_SECTION, path)
}

/**
 *  获取带默认值的配置值
 */
func MustString(path string, defaultValue string) (string) {
	return config.MustValue(goconfig.DEFAULT_SECTION, path, defaultValue)
}

/**
 *  获取带配置值：整型
 */
func GetInt(path string) (int) {
	result, err := config.Int(goconfig.DEFAULT_SECTION, path)
	if nil != err {
		fmt.Println("ooo - not found ", path, " setting section...")
	}

	return result
}

/**
 *  获取的配置值字符串
 */
func GetString(path string) (string) {
	result, err := config.GetValue(goconfig.DEFAULT_SECTION, path)
	if nil != err {
		fmt.Println("ooo - not found ", path, " setting section...")
	}

	return result
}

/**
 * 载入配置文件
 */
func loadConfigFile(path string) (*goconfig.ConfigFile) {
	//配置文件
	config, err := goconfig.LoadConfigFile(path)

	if nil != err {
		fmt.Println("ooo - LoadConfig fail, please check config path...")
	}

	return config
}