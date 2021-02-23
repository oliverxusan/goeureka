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
	Request(path string, param map[string]interface{}) interface{}
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

func (c *ClientService) Request(path string, param map[string]interface{}) interface{} {
	nodeList := c.getRegisterCenterData()
	if len(nodeList) == 0 {
		panic("Get Service Node List is null")
	}
	base := c.getServiceNode(nodeList) + "/" + path
	body, err := json.Marshal(param)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := goeureka.Req(base, goeureka.BytesToStr(body))
	if err != nil {
		log.Fatalln(err)
	}
	return resp
}

func (c *ClientService) getRegisterCenterData() []Node {
	instances, err := goeureka.GetAllServiceInstances(c.AppName)
	if err != nil {
		panic("Get Register Center Data" + err.Error())
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
