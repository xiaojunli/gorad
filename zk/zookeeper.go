package zk

var nodePath = "/gorad/user-api-service"

func RegServer(host string, port int)  {
	/*var hosts = []string{"localhost:8000"}

	flags := zk.FlagEphemeral
	acls := zk.WorldACL(zk.PermAll)

	dataMap := make(map[string] interface{})
	dataMap["host"] = host
	dataMap["port"] = port
	jsonStr, err := json.Marshal(dataMap)

	if err != nil {
		fmt.Println(err)
	}

	nodeData := []byte(jsonStr)

	conn, _, err := zk.Connect(hosts, time.Second*5, option)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	_, _, _, err = conn.ExistsW(nodePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	create(conn, nodePath, nodeData)

*/
}
