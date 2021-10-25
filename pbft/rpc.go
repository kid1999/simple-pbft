package pbft

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
)

/**
* @Description: PBFT RPC server
* @author : kid1999
* @date Date : 2021/10/25 5:43 PM
* @version V1.0
 */

func RpcRegister(pbft *PBFT) {
	//注册一个服务器
	err := rpc.Register(pbft)
	if err != nil {
		log.Panic(err)
	}
	//把服务绑定到http协议上
	rpc.HandleHTTP()
	//监听端口
	err = http.ListenAndServe(pbft.node.addr, nil)
	if err != nil {
		fmt.Println("注册rpc服务失败", err)
	}
}

func (p *PBFT) broadcast(method string, args interface{}, fun func(ok bool)) {
	//设置不要自己给自己广播
	for nodeID, addr := range NodeTable {
		if nodeID == p.node.nodeID {
			continue
		}
		//连接远程rpc
		rp, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			fun(false)
			continue
		}

		var bo = false
		err = rp.Call(method, args, &bo)
		if err != nil {
			fun(false)
			continue
		}
		fun(bo)
	}
}

func (p *PBFT) reply(r Reply, addr string, b *bool) {
	rp, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		fmt.Println("reply error.")
		return
	}
	rp.Call("PBFT.CheckReply", r, nil)
}
