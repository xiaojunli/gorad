package gorad

import (
	"strings"
	"strconv"
	"fmt"
	"time"
	"reflect"
	"io/ioutil"
	"path/filepath"

	"net"
	"os"
	"runtime"

	"github.com/xiaojunli/gorad/logger"
)

/**
 *
 */
func ParsePort(port string) (int, int){
	var err error
	var startPort, endPort = 8080, 8081

	if len(strings.TrimSpace(port)) > 0 {
		if  strings.Contains(port, "-") {
			portArray := strings.Split(port, "-")

			if nil != portArray && len(portArray) > 0 {
				startPort, err = strconv.Atoi(portArray[0])
				if nil != err {
					fmt.Println("ooo - start port fail, current value: " + portArray[0])
				}

				endPort, err = strconv.Atoi(portArray[1])
				if nil != err {
					fmt.Println("ooo - end port fail, current value: " + portArray[1])
				}
			}
		} else {
			startPort, err = strconv.Atoi(port)
			if nil == err {
				endPort = startPort
			}
		}
	}

	return startPort, endPort
}

func DefaultStaticHandles() [] *StaticHandle {
	return [] *StaticHandle {
		{Path: "/index.html", File: "./src/gorad.org/static/index.html"},
		{Path: "/favicon.ico", File: "./src/gorad.org/static/images/favicon.ico"},

		//{Path: "/static", File: "./src/gorad.org/static"},
		{Path: "/javascript", File: "./src/gorad.org/static/javascript"},
		{Path: "/images", File: "./src/gorad.org/static/images"},
	}
}

/**
 * 获取实例结构名称
 */
func GetStructName(instance interface{}) (string) {
	reflectType := reflect.TypeOf(instance).Elem()
	return reflectType.Name()
}


/**
 * 获取模块名称
 */
func GetModuleName(pkgPath string) (result string) {
	pathArray := strings.Split(strings.Replace(pkgPath, "\\", "/", -1), "/")

	if size := len(pathArray); size > 1 {
		return pathArray[size - 2]
	}

	return "_NONE_"
}

/**
 * 检测错误值
 */
func CheckError(err error) (result bool) {
	if nil != err {
		logger.Warn("ooo - exec file, message:", err)

		result = true
	}

	return result
}

/**
 * 生成路由文件
 */
func GenRouterFile(basePath string) (result string){
	startTime := time.Now()

	existsGroupHandles := make(map[string] *GroupHandle)

	projectPath := GetPathWithGoPath(basePath)

	routerPath := filepath.Join(projectPath, "router")
	path :=  filepath.Join(projectPath, "modules")

	//读取上级目录
	modulePaths, err := ioutil.ReadDir(path)
	if nil != err {
		fmt.Println(err)
	}

	//处理所有的controller子目录
	for _, modulePath := range modulePaths {
		subPath := filepath.Join(path, modulePath.Name(), "controller")
		ctlPackage := filepath.Join(basePath, "modules", modulePath.Name(), "controller")

		groupHandles, error := ParserPkg(subPath, ctlPackage)

		if nil != error {
			fmt.Println("error: ", error)
		}

		existsGroupHandles = mergeGroupHandle(existsGroupHandles, groupHandles, subPath)
	}

	result = SaveRouterFile(existsGroupHandles, routerPath)

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("ooo - gen router file and writer to routerFile, times:", duration, "ms. ")
	fmt.Println("------------------------------------------------------------------------")


	return result
}


/**
 * 合并运行值
 */
func mergeGroupHandle(existsGroupHandles map[string] *GroupHandle,
	groupHandles map[string] *GroupHandle, path string) (result map[string] *GroupHandle) {
	if nil == existsGroupHandles || len(existsGroupHandles) < 1 {
		existsGroupHandles = make( map[string] *GroupHandle)
	}

	//有值时进行合并
	if nil != groupHandles {
		//将解析出来的组注解合并到已经存在的groupComments
		for key, groupHandle := range groupHandles {
			methods := groupHandle.Methods

			_, ok := existsGroupHandles[key]
			if !ok {//不存在时直接赋值
				existsGroupHandles[key] = groupHandle
			} else {//存在时需要进行合并
				existMethods := existsGroupHandles[key].Methods

				if nil == existMethods {
					existMethods = make([] *MethodHandle, 0)
				}

				for _, method :=  range methods {
					existMethods = append(existMethods, method)
				}

				existsGroupHandles[key].Methods = existMethods
			}
		}
	} else {//无值时直接设置
		existsGroupHandles = groupHandles
	}

	return existsGroupHandles
}


//获取IP地址
func getIP() (*net.IP) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var result net.IP
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil {
				result = ipnet.IP.To4()
				break
			}
		}
	}

	return &result
}

/**
 截取字符串
 */
func SubString(source string, start int, end int) string {
	var r= []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	var substring= ""
	for i := start; i < length; i++ {
		substring += string(r[i])
	}

	return substring
}


// GetGOPATHs returns all paths in GOPATH variable.
func GetGOPATHs() []string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" && strings.Compare(runtime.Version(), "go1.8") >= 0 {
		gopath = defaultGOPATH()
	}
	return filepath.SplitList(gopath)
}

func defaultGOPATH() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	if home := os.Getenv(env); home != "" {
		return filepath.Join(home, "go")
	}
	return ""
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

/**
 * 获取相对GOPATH的路径
 */
func GetPathWithGoPath(passPath string) (string){
	goPaths := GetGOPATHs()
	if len(goPaths) == 0 {
		panic("you are in dev mode. So please set gopath")
	}

	path := ""
	for _, wg := range goPaths {
		wg, _ = filepath.EvalSymlinks(filepath.Join(wg, "src", passPath))
		if FileExists(wg) {
			path = wg
			break
		}
	}

	return path
}