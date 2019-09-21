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
package hosts

import (
	"XAgent/pkg/hostspath"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	annotation     = "# --- HOST DOMAIN FOR TEMPORARY PARSING ---"
	backupHostPath = hostspath.HostsPath + "bak"
)

type HostAccessor struct {
	HostLines map[string][]string
}

func (_self *HostAccessor) addHostLines() error {
	if _self.HostLines == nil {
		log.Printf("Failed to add hosts, hosts lines is null.")
		return nil
	}
	log.Printf("\nAdd hosts line to <%s>", hostspath.HostsPath)

	f, err := os.OpenFile(hostspath.HostsPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Open host file error: %v\n", err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, _ = fmt.Fprintln(w)             // New empty line.
	_, _ = fmt.Fprintln(w, annotation) // Separation

	for k, v := range _self.HostLines {
		line := fmt.Sprintf("%s\t\t%s", k, strings.Join(v, "\t\t"))
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func (_self *HostAccessor) backupHostsIfNecessary() bool {
	// Check if you have backed up.
	if isBackupHosts() {
		log.Printf("\nAlready backup hosts file: %s", backupHostPath)
		return false
	}
	log.Printf("\nBackuping hosts from: %s => %s", hostspath.HostsPath, backupHostPath)

	sourceFile, err := os.Open(hostspath.HostsPath)
	//养成好习惯。操作文件时候记得添加 defer 关闭文件资源代码
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer sourceFile.Close()
	//只写模式打开文件 如果文件不存在进行创建 并赋予 644的权限。详情查看linux 权限解释
	destFile, err := os.OpenFile(backupHostPath, os.O_CREATE|os.O_WRONLY, 644)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//养成好习惯。操作文件时候记得添加 defer 关闭文件资源代码
	defer destFile.Close()
	//进行数据拷贝
	_, copyErr := io.Copy(destFile, sourceFile)
	if copyErr != nil {
		log.Println(copyErr.Error())
		return false
	} else {
		return true
	}
}

func (_self *HostAccessor) watchHandleExitingAndReset() {
	// Check if you have backed up.
	if isBackupHosts() {
		log.Printf("\nAlready started watch exiting handler.")
		return
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c // 阻塞等待
		fmt.Printf("\nReset origin hosts %s \t=> %s", backupHostPath, hostspath.HostsPath)
		err := os.Rename(backupHostPath, hostspath.HostsPath)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}()
}

func (_self *HostAccessor) Run() {
	_self.watchHandleExitingAndReset()
	_self.backupHostsIfNecessary()
	_self.addHostLines()
}

func isBackupHosts() bool {
	_, err := os.Stat(backupHostPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// For testing.
//func main() {
//	fmt.Printf(BANNER)
//
//	hosts := make(map[string][]string)
//	hosts["127.0.0.1"] = []string{
//		"anjiancloud.alpha",
//		"mp.anjiancloud.alpha",
//		"ems.anjiancloud.alpha",
//		"portal.anjiancloud.alpha",
//	}
//
//	hp := &HostAccessor{HostLines: hosts}
//	hp.Run()
//
//	// Blocking the main thread to make the listening exit signal valid.
//	select {}
//}
