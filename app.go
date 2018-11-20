package gorad

import (
	"strings"
	"github.com/gin-gonic/gin"
	"strconv"
	"github.com/xiaojunli/gorad/config"
	"github.com/xiaojunli/gorad/plugins"
	"github.com/xiaojunli/gorad/logger"
	"github.com/xiaojunli/gorad/consul"
	"fmt"
	"reflect"

	"github.com/DeanThompson/ginpprof"
)

/**
 * 应用程序实例
 */
type App struct {
	Mode 			string 	//运行模式
	UseSession		bool 	//启用session
	UseHtml			bool	//启用页面

	//ThreadRoutes 	map[string] *Router //已启动的router
	ControllerMap 	map[string] interface{} //存储所有实例

	GroupHandles 	map[string] *GroupHandle //注册的路由信息
	parsedPkgPaths 	map[string] bool	//已解析路径

	StaticHandles 	[] *StaticHandle //静态文件处理
	Plugins 		[] *plugins.Plugin //插件

	HtmlGlob 		string //模板目录解析
}

/**
 * 注册静态文件
 */
func (self *App) RegStaticHandles(handles ...StaticHandle) {
	for _, handle := range handles {
		self.StaticHandles = append(self.StaticHandles, &handle)
	}
}


/**
 * 注册处理器
 */
func (self *App) RegHandle(ctl interface{}, methodType, groupPath, path, method, responseType string) {
	//TODO...
	logger.Debug("ooo - RegHandles controller...")

	//默认输出JSON
	if len(responseType) < 1 {
		responseType = "JSON"
	}

	if nil == self.GroupHandles {
		self.GroupHandles = make(map[string] *GroupHandle, 0)
	}

	//变化groupKey
	groupKey := groupPath
	if len(strings.TrimSpace(groupKey)) < 1 {
		groupKey = "_DEFAULT_"
	}

	//生成groupHandle
	groupHandle, ok := self.GroupHandles[groupKey]
	if !ok {
		groupHandle = & GroupHandle{
			Path: 			groupKey,
			ResponseType:	responseType,

			Methods: make([] *MethodHandle, 0),
		}
	}

	//注册句柄实例
	methodHandle := &MethodHandle{
		MethodType: methodType,
		Method: method,
		Path: path,
		ResponseType: responseType,
	}

	//合并到Methods
	groupHandle.Methods = append(groupHandle.Methods, methodHandle)

	//重设
	self.GroupHandles[groupKey] = groupHandle

	//全局实例
	ctlName := GetStructName(ctl)
	self.ControllerMap[ctlName] = ctl
}

/**
 * 通过结果注入
 */
func (self *App) RegControllerInfo(groupHandleMap map[string] *GroupHandle, ctlMap map[string] interface{}) {
	self.ControllerMap = ctlMap
	self.GroupHandles = groupHandleMap
}

/**
 * 注册controller, 调用注解解析器，仅源码运行模式（开发模式）
 */
func (self *App) RegController(controllers ...interface{}) {
	//TODO...
	logger.Debug("ooo - RegController and parse pkgPath...")

	if nil == self.ControllerMap {
		self.ControllerMap = make(map[string] interface{})
	}

	if nil == self.parsedPkgPaths {
		self.parsedPkgPaths = make(map[string]bool)
	}

	for _, ctl := range controllers {
		reflectType := reflect.TypeOf(ctl)
		if reflectType.Kind() == reflect.Ptr {
			reflectType = reflectType.Elem()
		}

		path := GetPathWithGoPath(reflectType.PkgPath())

		//实例存储
		ctlKey := GetModuleName(reflectType.PkgPath()) + "." + reflectType.Name()

		self.ControllerMap[ctlKey] = ctl

		//路径为合法有效的GO文件
		if ""  != path {
			if _, ok := self.parsedPkgPaths[path]; !ok {
				self.parsedPkgPaths[path] = true

				groupHandles, err := ParserPkg(ParseRelPath(path), path)
				if nil != err {
					logger.Warn(err)
					continue
				}

				self.GroupHandles = mergeGroupHandle(self.GroupHandles, groupHandles, path)
			}
		}//End of <<if ""  != path>>
	}
}

/**
 * 设置模板解析目录
 */
func (self *App) SetHtmlGlob(htmlGlob string) {
	self.HtmlGlob = htmlGlob
}

/**
 * 开启session
 */
func (self *App) SetUseSession(useSession bool) {
	self.UseSession = useSession
}

/**
 * 注册插件
 */
func (self *App) RegPlugin(plugin interface{}) {
	//TODO...
	logger.Debug("ooo - RegPlugin ...")
}

/**
 * 启动插件
 */
func (self *App) initPlugins(router *Router) {

	//TODO...
	fmt.Println("ooo - init and start plugins....")
}

/**
 * 初始化静态文件处理
 */
func (self *App) initStaticHandle(router *Router) {
	if self.UseHtml {
		if len(self.StaticHandles) < 1 {
			//默认静态文件处理器
			self.StaticHandles = DefaultStaticHandles()
		}

		for _, handle := range self.StaticHandles {
			router.Engine.StaticFile(handle.Path, handle.File)
		}
	}


}


/**
 * 注册方法
 */
func (self *App) intRouterHandle(router *Router) {
	//health健康检查
	router.Engine.GET(config.MustString("consul.check.url", "/health/check"),
		func(context *gin.Context) {
			context.JSON(200, gin.H{
				"code": 200,
				"message": "ok",
			})
		})

	//处理手工注册与注解注册
	for groupPath, groupHandle := range self.GroupHandles {
		if "_DEFAULT_" == groupPath {
			groupPath = "/"
		}

		groupRouter := router.Engine.Group(groupPath)

		for _, handle := range groupHandle.Methods {
			router.handleRouter(groupRouter, handle, self.UseSession, self.ControllerMap)
		}
	}
}


/**
 * 启动应用
 */
func (self *App) Run(params... string) {
	var host, port string

	//模式
	if len(self.Mode) < 0 {
		self.Mode = config.MustString("server.mode", "devel")
	}

	if strings.Contains(self.Mode, "prod") {
		gin.SetMode(gin.ReleaseMode)
	}

	if len(params) > 0 && params[0] != "" {
		strs := strings.Split(params[0], ":")

		if len(strs) == 2 {
			host = strs[0]
			port = strs[1]
		} else if len(strs) == 1 {
			port = strs[0]
		}

		/*if len(strs) > 0 && strs[0] != "" {
			host = strs[0]
		}
		if len(strs) > 1 && strs[1] != "" {
			port = strs[1]
		}*/
	}

	//服务器IP
	if len(host) < 1 {
		host = config.MustString("server.host", getIP().String())
	}

	if "*" == host {
		host = "127.0.0.1"
	}

	if len(port) < 1 {
		port = config.MustString("server.port", "8080")
	}

	startPort, endPort := ParsePort(port)

	//循环启动所有服务
	for cport := startPort; cport <= endPort; cport++ {
		go self.startServer(host, strconv.Itoa(cport))
	}

	select {}
}

/**
 启动服务
 */
func (self *App) startServer(host, port string) {
	//启动实例
	router := NewRouter()

	//是否为开发模式
	if strings.ToUpper(self.Mode) != "PROD" {
		ginpprof.Wrap(router.Engine)
	}

	//启动插件
	self.initPlugins(router)

	//启动静态目录
	self.initStaticHandle(router)

	//注册相关路径
	self.intRouterHandle(router)

	//设置模板解析目录
	if "" != self.HtmlGlob {
		router.Engine.LoadHTMLGlob(self.HtmlGlob)
	}

	//启动服务
	go router.Run(host + ":" + port)

	//注册服务
	if config.MustString("consul.address", "") != "" {
		consul.RegService(host, port)
	}

	logger.Info("******************************************************")
	logger.Info("ooo - [", host, ":", port, "], start server success....")
	logger.Info("******************************************************")
}
