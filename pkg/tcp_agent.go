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
	"fmt"
	"io"
	"net"
)

type TcpAgent struct {
	Listen string `json:"listen"`
	Expose string `json:"expose"`
	Pass   string `json:"pass"`
}

func (_self *TcpAgent) listenServer() {
	if len(_self.Listen) <= 0 {
		panic("TCP listen server endpoint must not be empty.")
	}
	lis, err := net.Listen("tcp", _self.Listen)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("建立连接错误:%v\n", err)
			continue
		}
		fmt.Println(conn.RemoteAddr(), conn.LocalAddr())
		go _self.handleClientRequest(conn)
	}
}

func (_self *TcpAgent) handleClientRequest(sconn net.Conn) {
	if len(_self.Pass) <= 0 {
		panic("TCP pass backend endpoint must not be empty.")
	}
	defer sconn.Close()
	dconn, err := net.Dial("tcp", _self.Pass)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", _self.Pass, err)
		return
	}
	ExitChan := make(chan bool, 1)

	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(dconn, sconn)
		fmt.Printf("往%s发送数据失败:%s\n", _self.Pass, err)
		ExitChan <- true
	}(sconn, dconn, ExitChan)

	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(sconn, dconn)
		fmt.Printf("从%s接收数据失败:%s\n", _self.Pass, err)
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	<-ExitChan

	dconn.Close()
}
