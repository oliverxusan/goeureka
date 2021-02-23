# goEureka

[![Production Ready](https://img.shields.io/badge/production-ready-blue.svg)](github.com/oliverxusan/goeureka)
[![License](https://img.shields.io/github/license/gogf/gf.svg?style=flat)](https://github.com/oliverxusan/goeureka)

# Installation
```
go get -u -v github.com/oliverxusan/goeureka
```

### 服务化调用列子
````
serviceName := "xxxx"
param := make(map[string]interface{})
param["para1"] = "ahhaha"
service.NEW(serviceName).Request("api/test2",param)
参数可以不传例如
service.NEW(serviceName).Request("api/test2")
````

### 服务注册

````
opt := make(map[string]string)
opt["username"] = "admin"
opt["password"] = "123456"
goeureka.RegisterClient("http://127.0.0.1:8716/", "", "serviceName", "8300", strconv.Itoa(43), opt)

````
