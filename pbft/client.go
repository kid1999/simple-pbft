package pbft

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/rpc"
	"time"
)

/**
* @Description: PBFT Client
* @author : kid1999
* @date Date : 2021/10/25 8:20 PM
* @version V1.0
 */

//等待节点访问
func (p *PBFT) getRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if len(request.Form["message"]) > 0 {
		msg := request.Form["message"][0]
		fmt.Println("request : ", msg)
		// 构建请求
		r := new(Request)
		r.Timestamp = time.Now().UnixNano()
		r.ClientAddr = p.node.addr
		r.Message.ID = getRandom()
		r.Message.Content = []byte(msg)
		fmt.Println("http监听到了消息，准备发送给领导者，消息id:", r.Message.ID)
		//http://localhost:8080/req?message=ohmygod
		rp, err := rpc.DialHTTP("tcp", NodeTable[p.currentLeader])
		if err != nil {
			log.Panic(err)
		}
		b := false
		err = rp.Call("PBFT.LeaderReceiveMessage", r, &b)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("消息是否已发送到领导者：", b)
		writer.Write([]byte("ok!!!"))
	}
}

func HttpListen(p *PBFT) {
	//创建getRequest()回调方法
	http.HandleFunc("/req", p.getRequest)
	fmt.Println("监听8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
		return
	}
}

func HttpListenTest(p *PBFT) {
	r := new(Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = p.node.addr
	r.Message.ID = getRandom()
	r.Message.Content = []byte("hello")
	fmt.Println("http监听到了消息，准备发送给领导者，消息id:", r.Message.ID)
	//http://localhost:8080/req?message=ohmygod
	rp, err := rpc.DialHTTP("tcp", NodeTable[p.currentLeader])
	if err != nil {
		log.Panic(err)
	}
	b := false
	err = rp.Call("PBFT.LeaderReceiveMessage", r, &b)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("消息是否已发送到领导者：", b)
}

//返回一个十位数的随机数，作为msgid
func getRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := rand.Int(rand.Reader, x)
		if err != nil {
			log.Panic(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}
