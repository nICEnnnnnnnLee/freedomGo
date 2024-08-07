<h1 align="center">  
    <strong>
        Freedom
    </strong>
</h1>
<p align="center">
    端到端数据流量伪装加密研究
  <br/>
    <strong>仅供学习研究使用，请勿用于非法用途</strong>
</p>


## :star:相关Repo
| 项目名称  | 简介 | 
| ------------- | ------------- |   
| [freedomGo](https://github.com/nICEnnnnnnnLee/freedomGo)  |  Go实现，包含local端、remote端  | 
| [freedom4py](https://github.com/nICEnnnnnnnLee/freedom4py)  |  python3实现，包含local端、remote端  | 
| [freedomRust](https://github.com/nICEnnnnnnnLee/freedomRust)  |  Rust实现，包含local端、remote端  | 
| [freedom4j](https://github.com/nICEnnnnnnnLee/freedom4j)  |  java实现，包含local端、remote端  | 
| [freedom4NG](https://github.com/nICEnnnnnnnLee/freedom4NG)  | Android java实现，仅包含local端；单独使用可作为DNS、Host修改器 | 
 



## :star:一句话说明  
将本地代理数据伪装成指向远程端的HTTP(S) WebSocket流量 或者 gRPC流量。

## :star:简介  
+ 在配置正确的情况下，Go、python3、java、Android版本的local端和remote端可以配合使用。  
+ local端实现了HTTP(S)、SOCKS5代理，仅需一个端口，即可自动识别各种代理类型。  
+ local端HTTP(S)代理支持按域名分流，可将流量分为直连和走remote端两种。  
+ local端到remote端可以套上一层HTTP(S)，表现行为与Websocket/gRPC无异，经测试**可过CDN与Nginx**。  
+ local端到remote端可以使用http2协议，但是中间层可能会缓存相关请求，未能实现转发。  
+ local端到remote端可以使用http3协议，但不支持CDN。  
+ local端到remote端支持简单的用户名密码验证。  

## :star:缺陷  
+ 仅支持TCP，不支持UDP

## :star:如何在Cloudflare Worker中使用该项目(remote端 + ws_real)
请参考[cf.js](/cf.js)

## :star:如何在Web框架中使用该项目(remote端)
实现了较多的基础Handler `func (w http.ResponseWriter, r *http.Request)`  
可以很方便地转化为其它框架的Handler，例如gin `func (c *gin.Context)`

<details>
<summary>基础http server示例</summary>


```go
// 相关代码在 extend文件夹下，该文件夹的内容可以独立出来

func StartRemoteHTTPWebSocketProxyServer() {
	// 在代码中初始化配置，或者设置环境变量也可。参考 extend/common.go
	UseSSL = true
	extend.Users = ...
	extend.Salt = ...
	extend.HttpPath = ...
	extend.DnsResolver = ...
	...
	// 设置路由
	http.HandleFunc(extend.HttpPath, extend.Handler)

	addr := fmt.Sprintf("%s:%d", extend.BindHost, extend.BindPort)
	var err error
	if UseSSL {
		err = http.ListenAndServeTLS(addr, extend.CertPath, extend.KeyPath, nil)
	} else {
		err = http.ListenAndServe(addr, nil)
	}
	panic(err)
}
```
</details>

## :star:如何配置  


<details>
<summary>local端配置</summary>



```yml
# socks5 http
ProxyType: http
# ws grpc http2 http3 ws_real
ProxyMode: grpc
BindHost: 127.0.0.1 
BindPort: 1081
# 在非Window系统下生效,可为空
#DNSServer: 114.114.114.114:53

# 按域名分流将下面注释去掉即可
# 全局代理将下面注释掉即可
# GeoDomain:
#   # 如果不匹配分流规则，那么就直连?
#   DirectIfNotInRules: true
#   # 同https://github.com/gfwlist/gfwlist
#   GfwPath: data/gfwlist.txt
#   # 将直连域名直接写入即可，每行一个
#   DirectPath: data/direct_domains.txt

# 该值可以是ip或者域名
RemoteHost: 127.0.0.1
RemotePort: 443
# 和远端的连接是否经过TLS加密
RemoteSSL: true
Salt: salt
Username: username
Password: pwd
# 是否允许不安全的HTTPS连接
AllowInsecure: true
## 模拟TLS特征 go custom random firefox ios chrome qq 360 safari android_okhttp edge
ClientHelloID: go
## 模拟该路径的ClientHello消息的特征，当 ClientHelloID = custom 时生效
##  1. 运行 freedomGo -t capture -c 127.0.0.1
##  2. 改host将 xxx.com 设为 127.0.0.1
##  3. 用Firefox浏览器访问 https://xxx.com/ 可获取Firefox浏览器特征数据(保存在当前工作目录下)。其它同理
ClientHelloRawPath:
## Alpn设置。 
##  当ProxyMode 为 ws/ws_real时，需要注意Websocket仅支持HTTP 1.1。 
##     此时，若服务器支持HTTP 2, 必须注释掉 `h2`; 不支持的时候，可以添加上，以模拟完整的TLS特征
##  当ProxyMode 为 http2时，必须添加 h2
ClientHelloNextProtos: 
    # - h2
    - http/1.1
# WebSocket 模拟的HTTP请求Path
# 该值gRPC无效, 默认为 /{GrpcServiceName}/Pipe
HttpPath: /
GrpcServiceName: freedomGo.grpc.Freedom
# 如果连接是经过加密的,该值还是Client Hello消息里面的SNI
HttpDomain: www.baidu.com
HttpUserAgent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0
```
</details>

<details>
<summary>remote端配置</summary>



```yml
# 可选模式为 grpc ws http2 http3 ws_real
ProxyMode: ws
BindHost: 127.0.0.1 
BindPort: 443 
Salt: salt
# 在非Window系统下生效,可为空
#DNSServer: 8.8.8.8:53
# ProxyMode为http3时生效，用来伪装成静态网站。为空时不做伪装。  
HTTP3WebDir: D:\WWW\
# ProxyMode为http3时生效，生效路径。 
HttpPath: /12345
UseSSL: true
# UseSSL为false时,下面三行可注释掉
SNI: www.baidu.com
CertPath: data/fullchain.pem
KeyPath: data/www.baidu.com.key
GrpcServiceName: freedomGo.grpc.Freedom
Users:
  user1: pwd1 
  username: pwd
```
</details>








## :star:如何运行  
```
$: .\freedomGo.exe -help
Usage of D:\Workspace\freedomGo\freedomGo.exe:
  -c string
        配置文件路径 (default "./conf.local.yaml")
  -t string
        模式local/remote (default "local")
```
+ 运行本地端  
```
freedomGo -t local -c "配置文件路径"
```

+ 运行远程端
```
freedomGo -t remote -c "配置文件路径"
```