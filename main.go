package main

import (
	"distributed_system/pbft"
	"fmt"
	"os"
	"time"
)

/**
* @Description: root function
* @author : kid1999
* @date Date : 2021/10/25 8:32 PM
* @version V1.0
 */

func main() {
	//为所有节点生成公私钥
	pbft.GenRsaKeys()
	id := os.Args[1]
	Run(id, pbft.NodeTable[id])
}

func Run(id string, addr string) {
	p := pbft.NewPBFT(id, addr)
	// 注册 pbft rpc server
	fmt.Println("node ", id, " started.")
	go pbft.RpcRegister(p)
	// 开启一个监听节点
	if id == "1" {
		pbft.HttpListen(p)
	}
	for true {
		time.Sleep(time.Second)
	}
}
