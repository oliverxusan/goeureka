package goeureka

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	VPort              string
	VLocalIp           string
	username           string                    // login username
	password           string                    // login password
	eurekaPath         = "/eureka/apps/"         // define eureka path
	discoveryServerUrl = "http://127.0.0.1:8761" // local eureka url
)

// RegisterClient register this app at the Eureka server
// params: eurekaUrl, eureka server url
// params: appName define your app name what you want
// params: port app instance port
// params: securePort
func RegisterClient(eurekaUrl string, localIp string, appName string, port string, securePort string, opt map[string]string) {
	eurekaUrl = strings.Trim(eurekaUrl, "/")
	user, _ := opt["username"]
	passwd, _ := opt["password"]
	if len(user) > 1 && len(passwd) > 1 {
		username = user
		password = passwd
		discoveryServerUrl = eurekaUrl
	} else if len(user) == 0 && len(passwd) == 0 {
		discoveryServerUrl = eurekaUrl
	} else {
		panic("username or password is valid!")
	}
	Register(appName, localIp, port, securePort)
}

// Register :register your app at the local Eureka server
// params: port app instance port
// params: securePort
// Register new application instance
// POST /eureka/v2/apps/appID
// Input: JSON/XML payload HTTP Code: 204 on success
func Register(appName string, localIp string, port string, securePort string) {
	appName = strings.ToUpper(appName)
	VPort = port
	if localIp == "" {
		VLocalIp = getLocalIP()
	} else {
		VLocalIp = localIp
	}
	cfg := newConfig(appName, VLocalIp, port, securePort)

	// define Register request
	registerAction := RequestAction{
		Url:         discoveryServerUrl + eurekaPath + appName,
		Method:      "POST",
		ContentType: "application/json;charset=UTF-8",
		Body:        cfg,
	}
	var result bool
	// loop send heart beat every 5s
	for {
		result = isDoHttpRequest(registerAction)
		if result {
			log.Println("Registration OK")
			handleSigtermProcess(appName)
			go startHeartbeat(appName, localIp)
			break
		} else {
			log.Println("Registration attempt of " + appName + " failed...")
			time.Sleep(time.Second * 5)
		}
	}

}

// GetAllServiceInstances is a function query all instances by appName
// params: appName
// Query for all appID instances
// GET /eureka/v2/apps/appID
// HTTP Code: 200 on success Output: JSON
func GetAllServiceInstances(appName string) ([]Instance, error) {
	var m ServiceResponse
	appName = strings.ToUpper(appName)
	// define get instance request
	requestAction := RequestAction{
		Url:         discoveryServerUrl + eurekaPath + appName,
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	log.Println("Query Eureka server using URL: " + requestAction.Url)
	bytes, err := exeQuery(requestAction)
	if len(bytes) == 0 {
		log.Printf("Query Eureka Response is None")
		return nil, err
	}
	if err != nil {
		return nil, err
	} else {
		//log.Println("Response from Eureka:\n" + string(bytes))
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			log.Printf("Parse JSON Error(%v) from Eureka Server Response", err.Error())
			return nil, err
		}
		return m.Application.Instance, nil
	}
}

// GetServiceInstanceIdWithappName : in this function, we can get InstanceId by appName
// Notes:
//		1. use SendHeartBeat
// 		2. deregister
// return instanceId, lastDirtyTimestamp
func GetInfoWithAppName(appName string) (string, string, error) {
	appName = strings.ToUpper(appName)
	instances, err := GetAllServiceInstances(appName)
	if err != nil {
		return "", "", err
	}
	var instanceId = VLocalIp + ":" + appName + ":" + VPort
	for _, ins := range instances {
		if ins.App == appName && ins.InstanceId == instanceId {
			return ins.InstanceId, ins.LastDirtyTimestamp, nil
		}
	}
	return "", "", err
}

// GetServices :get all services for eureka
// Notes: /gotest/TestGetServiceInstances has a test case
// Query for all instances
// GET /eureka/v2/apps
// HTTP Code: 200 on success Output: JSON
func GetServices() ([]Application, error) {
	var m ApplicationsRootResponse
	requestAction := RequestAction{
		Url:         discoveryServerUrl + eurekaPath,
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	log.Println("Query all services URL:" + requestAction.Url)
	bytes, err := exeQuery(requestAction)
	if err != nil {
		return nil, err
	} else {
		//log.Println("query all services response from Eureka:\n" + string(bytes))
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			log.Printf("Parse JSON Error(%v) from Eureka Server Response", err.Error())
			return nil, err
		}
		return m.Resp.Applications, nil
	}
}

// startHeartbeat function will start as goroutine, will loop indefinitely until application exits.
// params: appName
func startHeartbeat(appName string, localIp string) {
	for {
		time.Sleep(time.Second * 30)
		SendHeartBeat(appName, localIp)
	}
}

// heartbeat Send application instance heartbeat
// PUT /eureka/v2/apps/appID/instanceID
//HTTP Code:
//* 200 on success
//* 404 if instanceID doesn’t exist
func heartbeat(appName string, localIp string) {
	appName = strings.ToUpper(appName)
	instanceId, lastDirtyTimestamp, err := GetInfoWithAppName(appName)
	if instanceId == "" {
		log.Printf("instanceId is None , Please check at (%v) \n", discoveryServerUrl)
		return
	}
	if err != nil {
		log.Printf("Can't get instanceId from Eureka server by appName \n")
		return
	} else {
		if localIp != "" {
			// "58.49.122.210:GOLANG-SERVER:8889"
			instanceId = localIp + ":" + appName + ":" + VPort
		}
		heartbeatAction := RequestAction{
			//http://127.0.0.1:8761/eureka/apps/TORNADO-SERVER/127.0.0.1:tornado-server:3333/status?value=UP&lastDirtyTimestamp=1607321668458
			Url:         discoveryServerUrl + eurekaPath + appName + "/" + instanceId + "/status?value=UP&lastDirtyTimestamp=" + lastDirtyTimestamp,
			Method:      "PUT",
			ContentType: "application/json;charset=UTF-8",
		}
		log.Println("Sending heartbeat to " + heartbeatAction.Url)
		isDoHttpRequest(heartbeatAction)
	}
}

// SendHeartBeat is a test case for heartbeat
// you can test this function: send a heart beat to eureka server
func SendHeartBeat(appName string, localIp string) {
	heartbeat(appName, localIp)
}

// deregister De-register application instance
// DELETE /eureka/v2/apps/appID/instanceID
// HTTP Code: 200 on success
func deregister(appName string) {
	appName = strings.ToUpper(appName)
	log.Println("Trying to deregister application " + appName)
	instanceId, _, _ := GetInfoWithAppName(appName)
	// cancel registerion
	deregisterAction := RequestAction{
		//http://127.0.0.1:8761/eureka/apps/TORNADO-SERVER/127.0.0.1:tornado-server:3333/status?value=UP&lastDirtyTimestamp=1607321668458
		Url:         discoveryServerUrl + eurekaPath + appName + "/" + instanceId, //+ "/status?value=UP&lastDirtyTimestamp=" + lastDirtyTimestamp,
		ContentType: "application/json;charset=UTF-8",
		Method:      "DELETE",
	}
	isDoHttpRequest(deregisterAction)
	log.Println("Cancel App: " + appName + " InstanceId:" + instanceId)
}

// handleSigterm when has signal os Interrupt eureka would exit
func handleSigtermProcess(appName string) {
	c := make(chan os.Signal, 1)
	// Ctr+C shut down
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		deregister(appName)
		os.Exit(1)
	}()
}

func Req(url string, body string) (m map[interface{}]interface{}, err error) {
	requestAction := RequestAction{
		Url:         url,
		Method:      "POST",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
		Body:        body,
	}
	log.Println("Client URL:" + requestAction.Url)
	bytes, err := exeQuery(requestAction)
	if err != nil {
		return nil, err
	} else {
		//log.Println("query all services response from Eureka:\n" + string(bytes))
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			log.Printf("Parse JSON Error(%v) from Server Response", err.Error())
			return nil, err
		}
		return m, nil
	}
}
