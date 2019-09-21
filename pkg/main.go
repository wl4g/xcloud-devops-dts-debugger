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
	"encoding/json"
	"flag"
	"fmt"
	hosts2 "github.com/wl4g/super-devops-tool-debug-agent/pkg/hosts"
	"io/ioutil"
	"log"
	"time"
)

const (
	BANNER = `
__  _ ___                  _   
\ \/ | . | ___  ___ ._ _ _| |_ 
 \ \ |   |/ . |/ ._>| ' | | |  
_/\_\|_|_|\_. |\___.|_|_| |_|  
          <___'                
`
)

type Config struct {
	Tcp  []TcpAgent  `json:"tcp"`
	Http []HttpAgent `json:"http"`
}

var (
	config Config
)

func init() {
	confPath := ""
	// Command config path
	flag.StringVar(&confPath, "c", "resources/xagent.json", "XAgent config path.")
	flag.Parse()
	//flag.Usage()
	log.Printf("Initialize config path for - '%s'\n", confPath)
	confData, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Printf("Read config '%s' error! %s", confPath, err)
		panic(err)
		return
	}
	// Parse configuration.
	if err := json.Unmarshal(confData, &config); err != nil || &config == nil {
		log.Panicf("Failed started XAgent, parse configuration error. config:%v, %v", config, err)
	}
}

func main() {
	fmt.Printf(BANNER)
	fmt.Printf("\nwiki: http://wiki.wl4g.com/xagent/docs/index.html")
	fmt.Printf("\nversion: v1.0.0")
	fmt.Printf("\nauthors: <wanglsir@gmail.com, 983708408@qq.com>")
	fmt.Printf("\ntime: %s", time.Now().Format(time.RFC3339))
	fmt.Printf("\n")

	// Starting TCP channel port forwarding.
	if config.Tcp != nil && len(config.Tcp) > 0 {
		for _, t := range config.Tcp {
			addLocalhostDomain(t.Expose) // e.g. Add 127.0.0.1 => my.domain.com
			t.listenServer()
		}
	}

	// Starting HTTP channel route forwarding.
	if config.Http != nil && len(config.Http) > 0 {
		for _, h := range config.Http {
			h.ListenServer()
			for _, p := range h.Proxy {
				addLocalhostDomain(p.Expose) // e.g. Add 127.0.0.1 => my.domain.com
			}
		}
	}

}

func addLocalhostDomain(domain string) {
	if len(domain) <= 0 {
		log.Panicf("Failed to started, add host ip/domain is empty.")
	}
	hosts := make(map[string][]string)
	hosts["127.0.0.1"] = []string{domain}
	acc := &hosts2.HostAccessor{HostLines: hosts}
	acc.Run()
}
