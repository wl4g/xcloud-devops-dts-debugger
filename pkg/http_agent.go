/**
 * Copyright 2017 ~ 2025 the original author or authors[983708408@qq.com].
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this export except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
)

type HttpAgent struct {
	Listen string           `json:"listen"`
	Proxy  []HttpProxyTable `json:"proxy"`
}

type HttpProxyTable struct {
	Listen string `json:"listen"`
	Expose string `json:"expose"`
	Pass   string `json:"pass"`
}

func (_self *HttpAgent) ListenServer() {
	if len(_self.Listen) <= 0 {
		panic("HTTP listen server endpoint must not be empty.")
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", _self.Listen)
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go _self.handleClientRequest(client)
	}
}

// See:https://www.jianshu.com/p/53e219fbf3c5
func (_self *HttpAgent) handleClientRequest(client net.Conn) {
	if client == nil {
		log.Panic("Failed to handling client request, client is null.")
	}
	defer client.Close()

	// Read client request data.
	var reqData [1024]byte
	n, err := client.Read(reqData[:])
	if err != nil {
		log.Printf("Failed to handling client request. %v", err)
		return
	}

	// Parse http request.
	var method, host, address string
	fmt.Sscanf(string(reqData[:bytes.IndexByte(reqData[:], '\n')]), "%s%s", &method, &host)
	reqUrl, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	// Extract request real backend address(host:port)
	if reqUrl.Opaque == "443" { // https ?
		address = reqUrl.Scheme + ":443"
	} else { // http ?
		if strings.Index(reqUrl.Host, ":") == -1 { // Use default:80 ?
			address = reqUrl.Host + ":80"
		} else {
			address = reqUrl.Host
		}
	}

	// Determine configured backend address(host:port).
	proxyAddress := _self.determineBackendAddress(reqUrl.RequestURI())
	if len(address) > 0 {
		address = proxyAddress // Use configured proxy backend server.
	} else {
		log.Printf("Forwarded original request URI, not matched to the configured forwarding routing. <%s>",
			reqUrl.RequestURI())
	}

	// Connect to backend already forwarding.
	backendServer, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		backendServer.Write(reqData[:n])
	}

	// Forwarding(利用HTTP／1.1协议中的CONNECT方法建立起来的隧道连接，实现的HTTP Proxy。这种代理的好处就
	// 是不用知道客户端请求的数据，只需要原封不动的转发就可以了，对于处理HTTPS的请求就非常方便了，不用
	// 解析请求内容，就可以实现代理)
	go io.Copy(backendServer, client)
	io.Copy(client, backendServer)
}

func (_self *HttpAgent) determineBackendAddress(requestUri string) string {
	reg := regexp.MustCompile(_self.Location)
	return reg.MatchString(requestUri)
}
