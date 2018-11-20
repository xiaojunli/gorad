package consul

import (
	"fmt"
	"strconv"
	"github.com/hashicorp/consul/api"

	"github.com/xiaojunli/gorad/config"
	"github.com/xiaojunli/gorad/logger"

	)

var client *api.Client

func init() {
	var err error
	consulAddress := config.MustString("consul.address", "127.0.0.1:8500")

	cfg := &api.Config{ Address: consulAddress }
	client, err = api.NewClient(cfg)

	if nil != err {
		panic(err)
	}
}

/**
 注册服务
 */
func RegService(addr, port string) {
	serviceName := config.MustString("consul.service.name", "OSRAD")
	version := config.MustString("consul.service.version", "v1.0")
	author := config.MustString("consul.service.author", "v1.0")

	//创建一个新服务
	iPort, _ := strconv.Atoi(port)
	registration := new(api.AgentServiceRegistration)
	registration.ID = addr + ":" + port
	registration.Name = serviceName
	registration.Port = iPort
	registration.Tags = []string{serviceName, version}
	registration.Address = addr

	metaMap := make(map[string] string)
	metaMap["version"] = version
	metaMap["author"] = author

	registration.Meta = metaMap

	//增加check
	timeout := config.MustString("consul.check.timeout", "5s")
	interval := config.MustString("consul.check.interval", "5s")
	deregisterCriticalServiceAfter := config.MustString("consul.check.deregisterCriticalServiceAfter", "2m")

	check := new(api.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, "health/check")
	check.Timeout = timeout  //设置超时
	check.Interval = interval //设置间隔
	check.DeregisterCriticalServiceAfter = deregisterCriticalServiceAfter

	registration.Check = check //注册check

	err := client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Warn("register server error : ", err)
		return
	}

	logger.Debug("ooo - success reg to consul server: ", registration.Name, "->", registration.Address, "->", registration.Port)
}

/**
 获取服务列表
 */
func GetService(serviceName string) (*api.AgentService) {
	services, err := client.Agent().Services()

	if err != nil {
		logger.Fatal("register server error : ", err)
	}

	result, ok := services[serviceName]
	if !ok {
		logger.Debug("ooo - service not exist...")

		return nil
	}

	return result
}
