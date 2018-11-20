package gorad

import (
	"github.com/xiaojunli/gorad/config"
	"strings"
)

var app *App

var groupHandleMap map[string] *GroupHandle
var ctlMap map[string] interface{}


func init() {
	app = &App{}

	groupHandleMap = make(map[string] *GroupHandle)
	ctlMap = make(map[string] interface{})
}

/**
 * 注册Router
 */
func RegRouter(groupHandlesMap map[string] *GroupHandle, ctlsMap map[string] interface{}) {
	groupHandleMap = groupHandlesMap
	ctlMap = ctlsMap
}

/**
 * 手动注册
 */
func RegController(controllers... interface{}) {
	app.RegController(controllers...)
}

func SetUseHtml() {
	app.UseHtml = true
}

/**
 * 运行程序
 */
func Run(args... string) {
	app.RegControllerInfo(groupHandleMap, ctlMap)

	app.Run(args...)
}

/**
 * 生成Router文件，!!!本次生成下次有效!!!
 */
func MakeRouterFile(modulePath string) {
	mode := config.MustString("server.mode", "devel")

	//非产品模式时生成下次启动编译的目录
	if strings.Contains(strings.ToUpper(mode), "DEVEL") {
		GenRouterFile(modulePath)
	}
}



