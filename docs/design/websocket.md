# Websocket
## 概述
前端通过websocket长连接实时获取某一资源的变化情况，用于页面的动态展示
## 目标和动机
k8s中大多数的资源操作都是异步的，需要一个通用的接口（websocket）异步通知前端资源的变化
gorest中通过websocket实现资源的watch接口
## 详细设计
### 接口定义
* handler
```go
type WatchHandler func(*Context) (<-chan interface{}, *goresterr.APIError)

type Handler interface {
	...
	GetWatchHandler() WatchHandler
}
```
* ResourceKind
在ResourceKind中增加GetWatchObj的接口，用于gorestdoc通过反射获取某一资源的watch的对象信息，生成文档使用
```go
type ResourceKind interface {
	...
	GetWatchObj() interface{}
}
```
> watch的对象需要能够json序列化
### 资源回收
websocket实际为一个tcp长连接，所以后端函数通常通过一个goroutine循环实现，必须考虑后端goroutine资源回收
在gorest中通过Context进行控制
* 实现
```go
type Context struct {
	...
	stopCh   chan struct{}
}

func (ctx *Context) GetStopCh() <-chan struct{} {
	return ctx.stopCh
}

func (ctx *Context) CloseStopCh() {
	close(ctx.stopCh)
}
```
* 原理
在Context的构造函数中初始化一个stopCh，并提供两个Context的函数GetStopCh和CloseStopCh，后端handler实现中通过GetStopCh获取websocket的连接情况，若连接已断开，则执行goroutine的回收操作
## Example
```go
type CLusterEvent struct {
	EventType string `json:"eventType"`
	Cluster   string `json:"cluster"`
}

func (c Cluster) GetWatchObj() interface{} {
	return &CLusterEvent{}
}

func (h *clusterHandler) Watch(ctx *resource.Context) (<-chan interface{}, *goresterr.APIError) {
	result := make(chan interface{}, 0)
	go func() {
		for {
			select {
			case <-ctx.GetStopCh():
				fmt.Println("ws connection has been closed, will return")
				return
			default:
				result <- &CLusterEvent{
					EventType: "update",
					Cluster:   "wyw",
				}
				<-time.After(time.Second * 2)
			}
		}
	}()
	return result, nil
}
```
## Todo
* 支持websocket连接非正常断开异常处理，可以自定义message发送至前端

