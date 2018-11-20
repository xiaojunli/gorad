package gorad

import (
	"strings"
	"regexp"
	"go/token"
	"go/parser"
	"os"
	"go/ast"
	"fmt"
	"time"
	"errors"
	"strconv"
	"path/filepath"
	"io/ioutil"
	"github.com/xiaojunli/gorad/logger"
	)

//生成信息
var genInfoList  map[string] *GroupHandle

var groupRegexStr = `(\s+)?\((.*)\)`
var methodRegexStr = `(\s+)?\((.*)\)`
var paramRegexStr = `(\s+)?\((.*)\)`

var routerTemplate = `package router

import (
	"github.com/xiaojunli/gorad"
{{__IMPORT__}}
)

/**
 *  -------------------------------------------------------
 *  - Application Router Settings
 * -------------------------------------------------------
 *  - @Version: {{__VERSION__}}
 *  - @Time：{{__TIME__}}
 * -------------------------------------------------------
 *  - ！！！重要说明！！！
 *  -     本程序由系统自动生成开发模式下每次启动时将会按代
 *  - 码解析结果重新生成此文件，不要试图手动修改此文件
 * -------------------------------------------------------
 */
func init() {
	gorad.RegRouter(groupHandleMap, ctlInstanceMap)
}

{{__CTL_INSTANCES__}}

{{__GROUP_HANDLES__}}
`

var routeRegex = regexp.MustCompile(`@RequestMapping` + groupRegexStr)

//所有处理的定义在controller的struct中的注释
var groupRegexs = map[string] *regexp.Regexp {
	"RequestMapping" : regexp.MustCompile(`@RequestMapping` + groupRegexStr),
	"RestController" : regexp.MustCompile(`@RestController` + groupRegexStr),
}

//所有需要处理的路由记录
var methodRegexs = map[string] *regexp.Regexp {
	"RequestMapping" : regexp.MustCompile(`@RequestMapping` + methodRegexStr),

	"GetMapping" : 	regexp.MustCompile(`@GetMapping` + methodRegexStr),
	"PostMapping" : regexp.MustCompile(`@PostMapping` + methodRegexStr),
	"PutMapping" : 	regexp.MustCompile(`@PutMapping` + methodRegexStr),
	"DeleteMapping" : 	regexp.MustCompile(`@DeleteMapping` + methodRegexStr),
	"OptionsMapping" : 	regexp.MustCompile(`@OptionsMapping` + methodRegexStr),
	"PatchMapping" : 	regexp.MustCompile(`@PatchMapping` + methodRegexStr),
	"HeadMapping" : 	regexp.MustCompile(`@HeadMapping` + methodRegexStr),

	"AnyMapping" :		regexp.MustCompile(`@AnyMapping` + methodRegexStr),
}

/** 参数匹配 */
var paramRegexs = map[string] *regexp.Regexp {
	"RequestParam" : 	regexp.MustCompile(`@RequestParam` + paramRegexStr),
	"PathParam" : 		regexp.MustCompile(`@PathParam` + paramRegexStr),
	"PathVariable" : 	regexp.MustCompile(`@PathVariable` + paramRegexStr),
}


/**
 * 解析为真实路径
 */
func ParseRelPath(path string) (string){

	pathRegexp, _ := regexp.Compile("_/(.*)_")
	result := pathRegexp.FindAllStringSubmatch(path, -1)

	if len(result) > 0 {
		if len(result[0]) > 1 {
			path = strings.Replace(path, result[0][0], result[0][1] + ":", 1)
		}
	}

	return path
}

/**
 * 解析pkg路径
 *
 * @ modify beego's parser.go
 */
func ParserPkg(pkgRealpath string, pkgpath string) (map[string] *GroupHandle, error) {
	genInfoList = make(map[string] *GroupHandle)

	fileSet := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fileSet, pkgRealpath, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
	}, parser.ParseComments)

	if err != nil {
		return nil, err
	}
	for _, pkg := range astPkgs {
		for _, fl := range pkg.Files {
			//每个文件可能有一个组信息
			var groupHandle *GroupHandle

			for _, d := range fl.Decls {
				switch specDecl := d.(type) {
					//方法注解
					case *ast.FuncDecl:
						if specDecl.Recv != nil {
							exp, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr) // Check that the type is correct first beforing throwing to parser
							if ok {
								methodHandles, err := parseComments(specDecl, fmt.Sprint(exp.X), pkgpath, groupHandle)

								if nil == err && nil != methodHandles{
									currentMethods := groupHandle.Methods
									if nil == currentMethods {
										currentMethods = [] *MethodHandle{}
									}

									for _, method := range methodHandles {
										currentMethods = append(currentMethods, method)
									}

									groupHandle.Methods = currentMethods
								}
							}
						}
					break

					//类上注解：先处理类上注解后初始化groupComment值
					case *ast.GenDecl:
						if nil != specDecl.Doc && nil != specDecl.Doc.List && len(specDecl.Doc.List) > 0 {
							//注解数组
							groupArray := make([] string, len(specDecl.Doc.List))
							for idx, comment := range specDecl.Doc.List {
								groupArray[idx] = comment.Text
							}

							//获取类名称
							exp, ok := specDecl.Specs[0].(*ast.TypeSpec)
							if ok {
								groupHandle = parserGroupComments(groupArray, exp.Name.Name, pkgpath)
							}
						}
					break
				}
			}

			//合并groupComment, ResponseBody仅为下级方法返回格式
			group, ok := genInfoList[groupHandle.Path]
			if ok {
				methods := group.Methods

				//已存在的值合并到新值中
				for _, method := range groupHandle.Methods {
					methods = append(methods, method)
				}

				group.Methods = methods
			}

			//生成信息
			genInfoList[groupHandle.Path] = groupHandle
		}
	}

	return genInfoList, nil
}

/**
 * 将行处理为map值
 */
func parseLine(line string) (map[string]string, error) {
	resultMap := make(map[string]string)
	parseRegexp, _ := regexp.Compile("\"")

	value := parseRegexp.ReplaceAllString(line, "")
	if !strings.Contains(line, "=") {
		resultMap["path"] = strings.TrimSpace(value)
	}

	splitArray := strings.Split(value, ",")
	for _, kv := range splitArray {
		kvArray := strings.Split(kv, "=")

		if len(kvArray) > 1 {
			key := strings.TrimSpace(kvArray[0])
			if "value" == key {
				key = "path"
			}

			resultMap[key] = strings.TrimSpace(kvArray[1])
		}
	}

	if len(resultMap) == 0 {
		return resultMap, errors.New("can't parse comment line, current comment line: \"" + line + "\"")
	}

	return resultMap, nil
}

/**
 * 解析参数注解
 */
func parseParamLine(line string) (map[string]string, error) {
	resultMap := make(map[string]string)

	//处理\"字符
	parseRegexp, _ := regexp.Compile("\"")
	value := parseRegexp.ReplaceAllString(line, "")

	if !strings.Contains(line, "=") {
		resultMap["name"] = strings.TrimSpace(value)
	} else {
		splitArray := strings.Split(value, ",")
		for _, kv := range splitArray {
			kvArray := strings.Split(kv, "=")

			if len(kvArray) > 1 {
				key := strings.TrimSpace(kvArray[0])
				if "value" == key {
					key = "name"
				}

				resultMap[key] = strings.TrimSpace(kvArray[1])
			}
		}
	}

	if len(resultMap) == 0 {
		return resultMap, errors.New("can't parse comment line, current comment line: \"" + line + "\"")
	}

	return resultMap, nil
}

/**
 * 解析类上定义
 */
func parserGroupComments (comments []string, controllerName, pkgpath string) (*GroupHandle){
	var groupComment = &GroupHandle{ Path: "_DEFAULT_", Methods: make([] *MethodHandle, 0)}

	for _, line := range comments {
		//单行注释及有@符的注解有效
		if strings.Contains(line, "//") && strings.Contains(line, "@") {
			for name, groupRegex := range groupRegexs {
				if !strings.Contains(strings.ToUpper(name), "RESTCONTROLLER") {
					groupArray := groupRegex.FindAllStringSubmatch(line, -1)

					if len(groupArray) > 0 {
						//[0][2]表示匹配到的内容,[0][0]为带表达条件内容
						pMap, _ := parseLine(groupArray[0][2])

						if pPath, ok := pMap["path"]; ok {
							groupComment.Path = pPath
						}

						if pProduces, ok := pMap["producer"]; ok {
							groupComment.ResponseType = pProduces

						}
					} else {//有@RestController注解时返回内容为JSON
						groupComment.ResponseType = "JSON"
					}
				} //End of "if !strings.Contains(strings.ToUpper(name), "RESTCONTROLLER") {"
			}//End of for
		}//End of "if strings.Contains(line, "//") && strings.Contains(line, "@") {"
	}

	return groupComment
}

/**
 * 解析method comment区块
 */
func parseComments(f *ast.FuncDecl, controllerName, pkgPath string, groupComment *GroupHandle) (methodCommentArray []*MethodHandle, err error) {
	if f.Doc == nil {
		return nil,errors.New("f.Doc is nil, can't parse comments...")
	}

	//windows风格转为uninx风格
	pkgPath = strings.Replace(pkgPath, "\\", "/", -1)

	moduleName := GetModuleName(pkgPath)

	methodCommentArray = []*MethodHandle{}
	lines := f.Doc.List

	for _, c := range lines {
		var methodComment = &MethodHandle{ ControllerName: controllerName, PackagePath: pkgPath, ModuleName: moduleName}

		line := strings.TrimSpace(c.Text)
		if strings.Contains(line, "//") && strings.Contains(line, "@") {
			//1. 处理方法注解:@RequestMapping
			for name, methodRegex := range methodRegexs {
				methodArray := methodRegex.FindAllStringSubmatch(line, -1)

				if nil == methodArray {
					continue
				}

				//有匹配值时才设置method值
				methodComment.Method = f.Name.Name
				if len(methodArray[0]) > 0 {
					//fmt.Println("ooo - ", name, ":", strings.Join(groupArray[0], " -> "), " ---> controller:", controllerName)

					//[0][2]表示匹配到的内容,[0][0]为带表达条件内容
					pMap, _ := parseLine(methodArray[0][2])

					if pPath, ok := pMap["path"]; ok {
						methodComment.Path = pPath
					}

					if consumes, ok := pMap["consumes"]; ok {
						methodComment.Consumes = consumes
					}

					if headers, ok := pMap["headers"]; ok {
						methodComment.Headers = headers
					}

					if produces, ok := pMap["produces"]; ok {
						methodComment.Produces = produces
						methodComment.ResponseType = produces
					}

					if methodType, ok := pMap["method"]; ok {
						methodComment.MethodType = strings.ToUpper(methodType)
					}
				}

				switch name {
					case "GETMAPPING":
						methodComment.MethodType = "GET"
						break
					case "POSTMAPPING":
						methodComment.MethodType = "POST"
						break
					case "PUTMAPPING":
						methodComment.MethodType = "PUT"
						break
					case "DELETEMAPPING":
						methodComment.MethodType = "DELETE"
						break
					case "OPTIONSMAPPING":
						methodComment.MethodType = "OPTIONS"
						break
					case "HEADERMAPPING":
						methodComment.MethodType = "HEADER"
						break
					case "ANYMAPPING":
						methodComment.MethodType = "ANY"
						break
					default: //@RequestMapping, 无定义method时使用默认值"GET"
						if len(methodComment.MethodType) < 1 {
							methodComment.MethodType = "GET"
						}

						break
				}

				//默认返回值类型
				if len(methodComment.Produces) < 1 {
					methodComment.Produces = groupComment.ResponseType
					methodComment.ResponseType = groupComment.ResponseType
				}
			}
		}

		//非有效的注解时
		if len(methodComment.Path) < 1 && len(methodComment.Method) < 1 {
			continue
		}

		//2. 重新处理@Param中包含的参数的信息
		paramCommentArray := make([] *ParamHandle, 0)
		for _, p := range lines {
			for name, paramRegex := range paramRegexs {
				paramArray := paramRegex.FindAllStringSubmatch(p.Text, -1)

				if nil == paramArray {
					continue
				}

				paramComment := &ParamHandle{ }

				//有匹配值时才设置method值
				if len(paramArray[0]) > 0 {
					//[0][2]表示匹配到的内容,[0][0]为带表达条件内容
					pMap, _ := parseParamLine(paramArray[0][2])

					if pValue, ok := pMap["name"]; ok {
						paramComment.Name = pValue
					}

					if pDataType, ok := pMap["dataType"]; ok {
						paramComment.DataType = pDataType
					}

					if pDefault, ok := pMap["default"]; ok {
						paramComment.Default = pDefault
					}

					if pRequired, ok := pMap["required"]; ok {
						bRequired, err := strconv.ParseBool( pRequired )

						if nil != err {
							paramComment.Required = bRequired
						}
					}
				}

				switch strings.ToUpper(name) {
					case "REQUESTPARAM":
						paramComment.Type = PARAM_REQUEST_PARAM
						break
					case "PATHPARAM":
						paramComment.Type = PARAM_PATH_PARAM
						break
					case "PATHVARIABLE":
						paramComment.Type = PARAM_PATH_VARIABLE
						break
					default:
						break
				}

				//合并params
				if len(paramComment.Name) > 0 && len(methodComment.Path) > 0 {
					paramCommentArray = append(paramCommentArray, paramComment)
				}
			} //End of "for name, paramRegex := range paramRegexs"
		}// End of "for _, p := range lines"


		//3. 取实际参数列表
		paramArray := make([] *ParamHandle, 0)

		//取所有定义的参数列表
		for idx, param := range f.Type.Params.List {
			for i, p := range param.Names {
				paramHandle := &ParamHandle{}

				//原生类型
				exp, ok := param.Type.(*ast.Ident) // Check that the type is correct first beforing throwing to parser
				if ok {
					paramHandle.MethodParamType = exp.Name
				}

				//封装类型
				expStarExpr, okStarExpr := param.Type.(*ast.StarExpr) // Check that the type is correct first beforing throwing to parser
				if okStarExpr {
					refType := fmt.Sprint(expStarExpr.X)

					if strings.Contains(refType, "&") {
						selectorExpr := expStarExpr.X.(*ast.SelectorExpr)
						paramType := fmt.Sprint(selectorExpr.X) + "." + selectorExpr.Sel.Name

						//检测包名是否为系统包名
						if !checkSystemPackage(paramType) {
							paramHandle.MethodParamIsBindForm = true
						}

						paramHandle.MethodParamType = paramType
					} else{//无包名时（与项目同级）
						selectorExpr := expStarExpr.X.(*ast.Ident)

						paramHandle.MethodParamType = selectorExpr.Name
						paramHandle.MethodParamIsBindForm = true
					}
				}

				index := 0
				for j := 0; j < idx; j++ {
					index += len(f.Type.Params.List[j].Names)
				}
				index += i

				//类型为实际参数
				paramHandle.Type = FUNC_ARGS

				paramHandle.MethodParamName = p.Name
				paramHandle.MethodParamIndex = index

				paramHandle.Name = paramHandle.MethodParamName
				paramHandle.DataType = paramHandle.MethodParamType

				//将所有参数与注解中定义的参数绑定
				for _, paramCommentHandle := range paramCommentArray {
					if paramHandle.MethodParamName == paramCommentHandle.Name {
						//参数合并绑定
						paramHandle.Name = paramCommentHandle.Name
						paramHandle.Default = paramCommentHandle.Default
						paramHandle.Required = paramCommentHandle.Required
						paramHandle.Type = paramCommentHandle.Type

						//数据类型
						dataType := strings.TrimSpace(paramCommentHandle.DataType)
						if len(dataType) < 1 {
							dataType = "string"
						}
						paramHandle.DataType = dataType

						if paramHandle.DataType != paramHandle.MethodParamType {
							logger.Warn("ooo - Param: \"", paramHandle.Name, "\" dataType wrong, defined type: ",
								paramHandle.DataType, ", actual type: ", paramHandle.MethodParamType)

							paramHandle.DataType = paramHandle.MethodParamType
						}
					}
				}

				paramArray = append(paramArray, paramHandle)
			}
		}// End of "for idx, param := range f.Type.Params.List"

		//4. 将params合并到methodComment
		if len(paramArray) > 0 {
			methodComment.Params = paramArray
		}

		//5. 合并方向注解组
		methodCommentArray = append(methodCommentArray, methodComment)
	}

	return methodCommentArray, nil
}

/**
 * 检测系统包名
 */
func checkSystemPackage(packageName string) (result bool) {
	systemPackages := []string{
		"url.Values", "gin.Context", "http.Request", "gin.Params", "gin.ResponseWriter",
	}

	for _, systemPackage := range systemPackages {
		if packageName == systemPackage {
			return true
		}
	}

	return result
}

//////////////////////////////////////////////

/**
 * 保存组信息到文件
 */
func SaveRouterFile(groupHandles map[string]*GroupHandle, routerPath string) (string) {
	//import记录
	importMap := make(map[string]string, 0)

	//存储本groupHandle下所有的controller
	controllerMap := make(map[string]string, 0)

	groupHandlesStr := "//group handles settings\n"
	groupHandlesStr += "var groupHandleMap  = map[string] *gorad.GroupHandle {\n"
	for _, groupHandle := range groupHandles {
		groupHandlesStr += "	\"" + groupHandle.Path + "\": { \n"
		groupHandlesStr += "		Path:			\"" + groupHandle.Path + "\", \n"
		groupHandlesStr += "		ResponseType:	\"" + groupHandle.ResponseType + "\", \n"

		groupHandlesStr += "		Methods: 		[] *gorad.MethodHandle { \n"

		for _, method := range groupHandle.Methods {
			//导入信息
			if _, exists := importMap[method.ModuleName]; !exists {
				importMap[method.ModuleName] = method.PackagePath
			}

			//实例信息
			controllerNameKey := method.ModuleName + "." + method.ControllerName
			if _, exists := controllerMap[controllerNameKey]; !exists {
				controllerMap[controllerNameKey] = method.ControllerName
			}

			methodHandleStr := "			{\n"
			methodHandleStr += "				Path: 		\"" + method.Path + "\",\n"
			methodHandleStr += "				MethodType:	\"" + method.MethodType + "\",\n"
			methodHandleStr += "				Method:		\"" + method.Method + "\",\n"

			methodHandleStr += "				Produces:	\"" + method.Produces + "\",\n"
			methodHandleStr += "				Consumes:	\"" + method.Consumes + "\",\n"
			methodHandleStr += "				Headers:	\"" + method.Headers + "\",\n"
			methodHandleStr += "				ModuleName:	\"" + method.ModuleName + "\",\n"

			methodHandleStr += "				ResponseType: \"" + method.ResponseType + "\",\n"
			methodHandleStr += "				ControllerName: \"" + method.ControllerName + "\",\n"
			methodHandleStr += "				PackagePath:	\"" + method.PackagePath + "\",\n"


			methodHandleStr += "				Params: 	[] *gorad.ParamHandle { \n"

			for _, param := range method.Params {
				paramStr := "					{\n";
				paramStr += "						Name:		\"" + param.Name + "\",\n"
				paramStr += "						DataType:	\"" + param.DataType + "\",\n"
				paramStr += "						Default:	\"" + param.Default + "\",\n"
				paramStr += "						Type:		gorad." + fmt.Sprintf("%s", param.Type) + ",\n"
				paramStr += "						MethodParamName:	\"" + param.MethodParamName + "\",\n"
				paramStr += "						MethodParamType:	\"" + param.MethodParamType + "\",\n"
				paramStr += "						MethodParamIndex:	" + strconv.Itoa(param.MethodParamIndex) + ",\n"
				paramStr += "						MethodParamIsBindForm: " + strconv.FormatBool(param.MethodParamIsBindForm) + ",\n"
				paramStr += "					},\n";

				methodHandleStr += paramStr
			}

			methodHandleStr += "				}, \n"
			methodHandleStr += "			},\n"


			groupHandlesStr += methodHandleStr
		}

		groupHandlesStr += "		},\n"
		groupHandlesStr += "	}, \n\n"
	}
	groupHandlesStr += "} \n"

	//import中记录
	importStr := ""
	for name, path := range importMap {
		importStr += "\n	" + name + " \"" + path + "\""
	}

	//实例表
	instanceStr := "//controller instance map\n"
	instanceStr += "var ctlInstanceMap = map[string] interface{} {\n"
	for key, _ := range controllerMap {
		instanceStr += "	\"" + key + "\": &" + key + "{}, \n"
	}
	instanceStr += "}\n"

	routerTemplate = strings.Replace(routerTemplate, "{{__TIME__}}", time.Now().Format("2006-01-02 15:04:05"), -1)
	routerTemplate = strings.Replace(routerTemplate, "{{__VERSION__}}", "1.0.0", -1)


	routerTemplate = strings.Replace(routerTemplate, "{{__IMPORT__}}", importStr, -1)

	routerTemplate = strings.Replace(routerTemplate, "{{__CTL_INSTANCES__}}", instanceStr, -1)
	routerTemplate = strings.Replace(routerTemplate, "{{__GROUP_HANDLES__}}", groupHandlesStr, -1)


	fmt.Println("\n\n--------------------------------------")
	routerFile := filepath.Join(routerPath, "router.go")
	content := []byte(routerTemplate)

	err := ioutil.WriteFile(routerFile, content, 0644)

	if CheckError(err) {
		fmt.Println("- ooo Write RouterFile Fail: routerFile: \"", routerFile, "\"")
	} else {
		fmt.Println("- ooo Write RouterFile Success: \"", routerFile, "\", fileSize:", len(content) / 1024, "K.")
	}
	fmt.Println("--------------------------------------\n\n")


	return routerFile
}

///////////////////////////////////////////////////
func init() {

}