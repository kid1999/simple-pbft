package pbft

import (
	"encoding/hex"
	"fmt"
	"sync"
)

/**
* @Description: PBFT
* @author : kid1999
* @date Date : 2021/10/25 5:45 PM
* @version V1.0
 */

//本地消息池（模拟持久化层），只有确认提交成功后才会存入此池
var localMessagePool = []Message{}

// 节点信息
type node struct {
	//节点ID
	nodeID string
	//节点监听地址
	addr string
	//RSA私钥
	rsaPrivKey []byte
	//RSA公钥
	rsaPubKey []byte
}

type PBFT struct {
	//节点信息
	node node
	//每笔请求自增序号
	sequenceID int
	//锁
	lock sync.Mutex
	//临时消息池，消息摘要对应消息本体
	messagePool map[string]Request
	//存放收到的prepare数量(至少需要收到并确认2f个)，根据摘要来对应
	prePareConfirmCount map[string]map[string]bool
	//存放收到的commit数量（至少需要收到并确认2f+1个），根据摘要来对应
	commitConfirmCount map[string]map[string]bool
	//该笔消息是否已进行Commit广播
	isCommitBordcast map[string]bool
	//该笔消息是否已对客户端进行Reply
	isReply map[string]bool
	//当前节点的领导
	currentLeader string
	//心跳信号
	heartBeat chan bool
}

func NewPBFT(nodeID, addr string) *PBFT {
	p := new(PBFT)
	p.node.nodeID = nodeID
	p.node.addr = addr
	p.node.rsaPrivKey = p.getPivKey(nodeID) //从生成的私钥文件处读取
	p.node.rsaPubKey = p.getPubKey(nodeID)  //从生成的私钥文件处读取
	p.sequenceID = 0
	p.messagePool = make(map[string]Request)
	p.prePareConfirmCount = make(map[string]map[string]bool)
	p.commitConfirmCount = make(map[string]map[string]bool)
	p.isCommitBordcast = make(map[string]bool)
	p.isReply = make(map[string]bool)
	p.currentLeader = LeaderID
	p.heartBeat = make(chan bool)
	return p
}

// 处理客户端发来的请求
func (p *PBFT) LeaderReceiveMessage(request Request, b *bool) error {
	fmt.Println("主节点已接收到客户端发来的request ...", request.ID)
	//添加信息序号
	p.sequenceIDAdd()
	//获取消息摘要
	digest := getDigest(request)
	fmt.Println("已将request存入临时消息池")
	//存入临时消息池
	p.messagePool[digest] = request
	//主节点对消息摘要进行签名
	digestByte, _ := hex.DecodeString(digest)
	signInfo := RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
	//拼接成PrePrepare，准备发往follower节点
	pp := PrePrepare{request, digest, p.sequenceID, signInfo}
	fmt.Println("正在向其他节点进行进行PrePrepare广播 ...")
	//进行PrePrepare广播
	go p.broadcast("PBFT.HandlePrePrepare", pp, func(ok bool) {})
	fmt.Println("PrePrepare广播完成")
	*b = true
	return nil
}

//处理预准备消息
func (p *PBFT) HandlePrePrepare(pp PrePrepare, b *bool) error {
	fmt.Println(p.node.nodeID, "节点已接收到主节点发来的PrePrepare ...")
	//获取主节点的公钥，用于数字签名验证
	primaryNodePubKey := p.getPubKey(p.currentLeader)
	digestByte, _ := hex.DecodeString(pp.Digest)
	if digest := getDigest(pp.RequestMessage); digest != pp.Digest {
		fmt.Println("信息摘要对不上，拒绝进行prepare广播")
	} else if p.sequenceID+1 != pp.SequenceID {
		fmt.Println("消息序号对不上，拒绝进行prepare广播")
	} else if !RsaVerySignWithSha256(digestByte, pp.Sign, primaryNodePubKey) {
		fmt.Println("主节点签名验证失败！,拒绝进行prepare广播")
	} else {
		//序号赋值
		p.sequenceID = pp.SequenceID
		//将信息存入临时消息池
		fmt.Println("已将消息存入临时节点池")
		p.messagePool[pp.Digest] = pp.RequestMessage
		//节点使用私钥对其签名
		sign := RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
		//拼接成Prepare
		pre := Prepare{pp.Digest, pp.SequenceID, p.node.nodeID, sign}
		//进行准备阶段的广播
		fmt.Println("正在进行Prepare广播 ...")
		go p.broadcast("PBFT.HandlePrepare", pre, func(ok bool) {})
		fmt.Println("Prepare广播完成")
	}
	return nil
}

//处理准备消息
func (p *PBFT) HandlePrepare(pre Prepare, b *bool) error {
	fmt.Printf("本节点已接收到%s节点发来的Prepare ... \n", pre.NodeID)
	//获取消息源节点的公钥，用于数字签名验证
	MessageNodePubKey := p.getPubKey(pre.NodeID)
	digestByte, _ := hex.DecodeString(pre.Digest)
	if _, ok := p.messagePool[pre.Digest]; !ok {
		fmt.Println("当前临时消息池无此摘要，拒绝执行commit广播")
	} else if p.sequenceID != pre.SequenceID {
		fmt.Println("消息序号对不上，拒绝执行commit广播")
	} else if !RsaVerySignWithSha256(digestByte, pre.Sign, MessageNodePubKey) {
		fmt.Println("节点签名验证失败！,拒绝执行commit广播")
	} else {
		p.setPrePareConfirmMap(pre.Digest, pre.NodeID, true)
		count := 0
		for range p.prePareConfirmCount[pre.Digest] {
			count++
		}
		//因为主节点不会发送Prepare，所以不包含自己
		specifiedCount := 0
		if p.node.nodeID == LeaderID {
			specifiedCount = NodeCount / 3 * 2
		} else {
			specifiedCount = (NodeCount / 3 * 2) - 1
		}
		//如果节点至少收到了2f个prepare的消息（包括自己）,并且没有进行过commit广播，则进行commit广播
		p.lock.Lock()
		//获取消息源节点的公钥，用于数字签名验证
		if count >= specifiedCount && !p.isCommitBordcast[pre.Digest] {
			fmt.Println("本节点已收到至少2f个节点(包括本地节点)发来的Prepare信息 ...")
			//节点使用私钥对其签名
			sign := RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
			c := Commit{pre.Digest, pre.SequenceID, p.node.nodeID, sign}
			//进行提交信息的广播
			fmt.Println("正在进行commit广播")
			go p.broadcast("PBTF.HandleCommit", c, func(ok bool) {})
			p.isCommitBordcast[pre.Digest] = true
			fmt.Println("commit广播完成")
		}
		p.lock.Unlock()
	}
	return nil
}

//处理提交确认消息
func (p *PBFT) HandleCommit(c Commit, b *bool) error {
	fmt.Printf("本节点已接收到%s节点发来的Commit ... \n", c.NodeID)
	//获取消息源节点的公钥，用于数字签名验证
	MessageNodePubKey := p.getPubKey(c.NodeID)
	digestByte, _ := hex.DecodeString(c.Digest)
	if _, ok := p.prePareConfirmCount[c.Digest]; !ok {
		fmt.Println("当前prepare池无此摘要，拒绝将信息持久化到本地消息池")
	} else if p.sequenceID != c.SequenceID {
		fmt.Println("消息序号对不上，拒绝将信息持久化到本地消息池")
	} else if !RsaVerySignWithSha256(digestByte, c.Sign, MessageNodePubKey) {
		fmt.Println("节点签名验证失败！,拒绝将信息持久化到本地消息池")
	} else {
		p.setCommitConfirmMap(c.Digest, c.NodeID, true)
		count := 0
		for range p.commitConfirmCount[c.Digest] {
			count++
		}
		//如果节点至少收到了2f+1个commit消息（包括自己）,并且节点没有回复过,并且已进行过commit广播，则提交信息至本地消息池，并reply成功标志至客户端！
		p.lock.Lock()
		if count >= NodeCount/3*2 && !p.isReply[c.Digest] && p.isCommitBordcast[c.Digest] {
			fmt.Println("本节点已收到至少2f + 1 个节点(包括本地节点)发来的Commit信息 ...")
			//将消息信息，提交到本地消息池中！
			localMessagePool = append(localMessagePool, p.messagePool[c.Digest].Message)
			fmt.Println("正在reply客户端 ...")
			r := Reply{NodeID: p.node.nodeID, MessageID: p.messagePool[c.Digest].ID, Result: true}
			//将消息返回客户端
			go p.reply(r, p.messagePool[c.Digest].ClientAddr, b)
			p.isReply[c.Digest] = true
			fmt.Println("reply完毕")
		}
		p.lock.Unlock()
	}
	return nil
}

func (p *PBFT) CheckReply(r Reply, b *bool) error {
	fmt.Println(r)
	return nil
}

//为多重映射开辟赋值
func (p *PBFT) setPrePareConfirmMap(val, val2 string, b bool) {
	if _, ok := p.prePareConfirmCount[val]; !ok {
		p.prePareConfirmCount[val] = make(map[string]bool)
	}
	p.prePareConfirmCount[val][val2] = b
}

//为多重映射开辟赋值
func (p *PBFT) setCommitConfirmMap(val, val2 string, b bool) {
	if _, ok := p.commitConfirmCount[val]; !ok {
		p.commitConfirmCount[val] = make(map[string]bool)
	}
	p.commitConfirmCount[val][val2] = b
}

//序号累加
func (p *PBFT) sequenceIDAdd() {
	p.lock.Lock()
	p.sequenceID++
	p.lock.Unlock()
}
