package service

import (
	"encoding/json"
	"github.com/oliverxusan/goeureka"
	"log"
	"strconv"
	"strings"
)

type ClientInterface interface {
	//获得服务化名称
	GetServiceName() string
	//调用服务化
	Request(path string, param ...interface{}) interface{}
	//获取注册中心节点
	getRegisterCenterData() []Node
	//负载均衡
	LoadBalanceStrategy
}
type ClientService struct {
	Schema   string
	AppName  string
	NodeList []Node
	Strategy LoadBalanceStrategy
}

//节点结构体
type Node struct {
	Ip   string
	Port string
}
type Error struct {
	ErrorNo  int
	ErrorMsg string
}

func ErrorNew(err string) *Error {
	return &Error{ErrorNo: -1, ErrorMsg: err}
}
func (e *Error) Error() string {
	return e.ErrorMsg
}

func NEW(appName string) *ClientService {
	c := &ClientService{
		Schema:   "http://",
		AppName:  strings.ToUpper(appName),
		Strategy: newRoundRobin(),
	}
	return c
}

func (c *ClientService) GetServiceName() string {
	return c.AppName
}

func (c *ClientService) getServiceNode(nodeList []Node) string {
	return c.Schema + c.Strategy.getServiceNode(nodeList)
}

func (c *ClientService) Request(path string, param ...interface{}) (response *goeureka.Response, err error) {
	nodeList := c.getRegisterCenterData()
	if len(nodeList) == 0 {
		log.Println("ERROR Get Service Node List is null")
		return nil, ErrorNew("ERROR Get Service Node List is null")
	}
	base := c.getServiceNode(nodeList) + "/" + path

	if len(param) > 0 && param[0] != nil {
		body, err := json.Marshal(param[0])
		if err != nil {
			panic(err)
		}
		method := "POST"
		if len(param) >= 2 {
			if strings.ToUpper(param[1].(string)) != "POST" || strings.ToUpper(param[1].(string)) != "GET" {
				log.Println("ERROR Request Method is mistake!please use post or get method.")
				return nil, ErrorNew("Request Method is mistake!please use post or get method.")
			}
			method = param[1].(string)
		}
		bytes, err := goeureka.Req(base, goeureka.BytesToStr(body), method)
		if err != nil {
			log.Printf("ERROR goeureka.Req %s", err.Error())
			return nil, err
		}
		err = json.Unmarshal(bytes, &response)
		if err != nil {
			log.Printf("Parse JSON Error(%v) from Eureka Server Response", err.Error())
			return nil, err
		}
		return response, nil
	} else {
		body := []byte("")
		bytes, err := goeureka.Req(base, goeureka.BytesToStr(body), "POST")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bytes, &response)
		if err != nil {
			log.Printf("Parse JSON Error(%v) from Eureka Server Response", err.Error())
			return nil, err
		}
		return response, nil
	}
}

func (c *ClientService) getRegisterCenterData() []Node {
	instances, err := goeureka.GetAllServiceInstances(c.AppName)
	if err != nil {
		log.Println("ERROR Get Register Center Data" + err.Error())
		return nil
	}
	nodeList := make([]Node, len(instances))
	if len(instances) > 0 {
		for k, ins := range instances {
			nodeList[k] = Node{
				Ip:   ins.IpAddr,
				Port: strconv.Itoa(ins.Port.Port),
			}
		}
	}
	return nodeList
}

type LoadBalanceStrategy interface {
	getServiceNode(nodeList []Node) string
}

//随机轮询
type RoundRobinStrategy struct {
}

func (s *RoundRobinStrategy) getServiceNode(nodeList []Node) string {
	rand := goeureka.NewRand()
	index := rand.RandRobin2(len(nodeList))
	node := nodeList[index]
	return node.Ip + ":" + node.Port
}

func newRoundRobin() LoadBalanceStrategy {
	balance := &RoundRobinStrategy{}
	return balance
}
