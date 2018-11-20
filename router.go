package gorad

import (
	"github.com/gin-gonic/gin"
	"github.com/codegangsta/inject"
	"github.com/xiaojunli/gorad/logger"
	"github.com/gin-contrib/sessions"
	"reflect"
	"strings"
	"errors"
	)

//参数类型
type ParamType int

const(
	FUNC_ARGS ParamType = iota 			//实际参数

	PARAM_REQUEST_PARAM			//@RequetParam
	PARAM_PATH_PARAM			//@PathParam
	PARAM_PATH_VARIABLE			//@PathVariable
)

/**
 * 枚举转字符值
 */
func (pt ParamType) String() string {
	switch pt {
	case FUNC_ARGS:
		return "FUNC_ARGS"
	case PARAM_REQUEST_PARAM:
		return "PARAM_REQUEST_PARAM"
	case PARAM_PATH_PARAM:
		return "PARAM_PATH_PARAM"
	case PARAM_PATH_VARIABLE:
		return "PARAM_PATH_VARIABLE"
	}

	return "N/A"
}

/////////////////////////////////////////
type Session sessions.Session

/////////////////////////////////////////

//返回值为json
type JsonResult struct {
	Code int
	Message string
	Result interface{}
}

//返回值为XML
type XMLResult struct {
	Code int
	Result []interface{}
}


//返回值为html
type HtmlResult struct {
	Code int
	Name string
	Result interface{}
}

//返回值为Text
type TextResult struct {
	Code int
	Format string
	Result []interface{}
}

/////////////////////////////////////

/**
 * 静态文件处理器
 */
type StaticHandle struct {
	Path string
	File string
}

/**
 * 组控制器
 */
type GroupHandle struct {
	Path 			string
	ResponseType 	string	//输出类型
	Name			string  //组名称

	Methods [] *MethodHandle
}

/**
 * 处理方法级注解信息
 */
type MethodHandle struct {
	Path string
	MethodType string
	Method    string
	ResponseType string

	Produces   string
	Consumes   string
	Headers    string

	Params     [] *ParamHandle

	ControllerName	string
	PackagePath		string
	ModuleName		string
}

/**
 * 注解中的参数
 */
type ParamHandle struct {
	Name     string			//注解名称
	DataType string			//注解数据类型
	Default string			//注解默认值
	Required bool			//是否为必须参数

	Type ParamType				//注解类型

	MethodParamName string //实际参数名称
	MethodParamType string //实际参数类型
	MethodParamIndex int //实际参数索引号

	MethodParamIsBindForm bool	//是否为进行表单绑定对象
	//MethodParamFormArgType reflect.Type //要绑定为表单对象实例类型
}

/////////////////////////////////////

//路由信息
type Router struct {
	Engine *gin.Engine
	Group *gin.RouterGroup

	inj inject.Injector
}

/**
 * 实例化
 */
func NewRouter() *Router {
	router := &Router{
		Engine: gin.Default(),
		inj: inject.New(),
	}

	//TODO.....................
	//logger middleware
	router.Engine.Use(gin.Logger())

	//recovery middleware
	router.Engine.Use(gin.Recovery())
	///////////////////////////////

	return router
}

/**
创建组
 */
func (self *Router) NewGroup(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	group := self.Engine.Group(path, handlers...)

	self.Group = group

	return self.Group
}

/**
 * 获取Query
 */
func getQuery(context *gin.Context) (result *Query) {
	request := context.Request

	//解析表单
	request.ParseForm()
	request.ParseMultipartForm(32 << 20) // 32 MB

	result = & Query{
		Params: 	&context.Params,
		Form : 		&request.Form,
		PostForm : 	&request.PostForm,
	}

	return result
}

/**
 * 请求处理
 */
func (self *Router) handleRouter(routerGroup *gin.RouterGroup,
	methodHandle *MethodHandle,
	useSession bool,
	controllerMap map[string] interface{} ) {

	ctlInstance, ok := controllerMap[methodHandle.ModuleName + "." + methodHandle.ControllerName]
	if !ok {
		return
	}

	ctlValue := reflect.ValueOf(ctlInstance)
	method := ctlValue.MethodByName(methodHandle.Method)

	ctlType := reflect.TypeOf(ctlInstance)
	methodTypeValue, err := ctlType.MethodByName(methodHandle.Method)

	if !err {
		logger.Error(errors.New("method:" + methodHandle.Method + " not found!"))
		return
	}

	methodType := methodTypeValue.Type

	if !method.IsValid() {
		logger.Warn("ooo - [WARN] method not exists, router info: ", ", path: ",
			methodHandle.Path, ", methodName: ", methodHandle.Method, ", controllerName: ",
			methodHandle.ControllerName, "]")
		return
	}

	handler := func(context *gin.Context) {
		//Query值由Param, Form, PostForm中获取
		var query = getQuery(context)

		args := make ([] reflect.Value, 0)
		for _, param := range methodHandle.Params {
			//自定义对象需要进行表单与对象绑定
			if param.MethodParamIsBindForm {
				var instance interface{}

				//实际参数第1个为结构体类型
				inBindFormArg := methodType.In(param.MethodParamIndex + 1)
				if nil != inBindFormArg {
					instance = reflect.New(inBindFormArg.Elem()).Interface()

					err := context.ShouldBind(instance)

					if nil != err {
						logger.Error("ooo - bind form error, arg: ", param.Name, " , type: ", param.MethodParamType)
					}
				}

				args = append(args, reflect.ValueOf(instance))
			} else if strings.Contains(param.MethodParamType, ".") {
				//按名称注入
				isFound := true
				switch strings.ToUpper(param.Name) {
				case "FORM":
					args = append(args, reflect.ValueOf(query.Form))
					break
				case "POSTFORM":
					args = append(args, reflect.ValueOf(query.PostForm))
					break
				case "PARAMS":
					args = append(args, reflect.ValueOf(query.Params))
					break
				case "QUERY":
					args = append(args, reflect.ValueOf(query))
					break
				case "REQUEST":
					args = append(args, reflect.ValueOf(context.Request))
					break
				case "RESPONSE":
					args = append(args, reflect.ValueOf(context.Writer))
					break
				case "CONTEXT":
					args = append(args, reflect.ValueOf(context))
					break
				default:
					isFound = false
					break
				}

				//按类型注入
				if !isFound {//按名称无法注入时使用类型进行注入
					switch strings.ToUpper(param.MethodParamType) {
					case "GIN.CONTEXT":
						args = append(args, reflect.ValueOf(context))
						break
					case "URL.VALUES":
						args = append(args, reflect.ValueOf(query.Form))
						break
					case "GIN.PARAMS":
						args = append(args, reflect.ValueOf(query.Params))
						break
					case "HTTP.REQUEST":
						args = append(args, reflect.ValueOf(context.Request))
						break
					case "GIN.RESPONSEWRITER":
						args = append(args, reflect.ValueOf(context.Writer))
						break
					default:
						args = append(args, reflect.ValueOf(nil))
						break
					}
				}
			} else {
				//原生类型
				var value interface{}

				switch strings.ToUpper(param.MethodParamType) {
				case "INT":
					value = query.MustInt(param.MethodParamName, 0)
					break
				case "INT64":
					value = query.MustInt(param.MethodParamName, 0)
					break
				case "INT8":
					value = query.MustInt8(param.MethodParamName, 0)
					break
				case "INT16":
					value = query.MustInt16(param.MethodParamName, 0)
					break
				case "INT32":
					value = query.MustInt32(param.MethodParamName, 0)
					break

				case "UINT":
					value = query.MustUint(param.MethodParamName, 0)
					break
				case "UINT64":
					value = query.MustUint(param.MethodParamName, 0)
					break
				case "UINT8":
					value = query.MustUint8(param.MethodParamName, 0)
					break
				case "UINT16":
					value = query.MustUint16(param.MethodParamName, 0)
					break
				case "UINT32":
					value = query.MustUint32(param.MethodParamName, 0)
					break

				case "FLOAT":
					value = query.MustFloat64(param.MethodParamName, 0.0)
					break
				case "FLOAT64":
					value = query.MustFloat64(param.MethodParamName, 0.0)
					break

				case "BOOL":
					value = query.MustGetBool(param.MethodParamName, false)
					break

				case "STRING":
					value = query.MustString(param.MethodParamName, "")
					break
				}

				args = append(args, reflect.ValueOf(value))
			}
		}

		results := method.Call(args)

		var result interface{}

		if len(results) > 1 {
			resultArray := make([] interface{}, len(results))

			for idx, res := range results {
				value := res.Interface()

				resultArray[idx] = value
			}

			result = resultArray
		} else if len(results) == 1 {
			result = results[0].Interface()
		}


		switch strings.ToUpper(methodHandle.ResponseType) {
			case "JSON":
				context.Header("Content-Type", "application/json; charset=utf-8")

				context.JSON(200, gin.H{
					"Code": 200,
					"Message": "success",
					"Result": result,
				})
				break
			case "STRING":
				context.Header("Content-Type", "text/plain; charset=utf-8")

				if nil == result {
					result = "success"
				}

				context.String(200, "%s", result)
				break
			case "HTML":
				context.Header("Content-Type", "text/html; charset=utf-8")

				context.HTML(200, "template.tmpl", result)
				break
			case "XML":
				context.Header("Content-Type", "text/xml; charset=utf-8")

				context.XML(200, result)
				break
			default:
				context.Header("Content-Type", "application/json; charset=utf-8")

				context.JSON(200, result)
			break
		}
	}

	routerGroup.Handle(methodHandle.MethodType, methodHandle.Path, handler)

	logger.Info("ooo - [GROUP_ROUTER] gPath: '", routerGroup.BasePath(), "', method: '", methodHandle.MethodType,
				"' , path: '",  methodHandle.Path, "', ctl: '", methodHandle.ControllerName, "', method: '", methodHandle.Method, "' success...")
}

////////////////////////////////////////////////////

/**
 * 运行服务
 */
func (self *Router) Run(addrs ... string) *Router {
	err := self.Engine.Run(addrs ...)

	if nil != err {
		panic(err)
	}

	return self
}
