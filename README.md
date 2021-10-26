# Practical Byzantine Fault Tolerance 

> Practical Byzantine Fault Tolerance 实用拜占庭容错，是传统BFT的一种可用性实现，算法的时间复杂度为O(n2)

### 总览![](https://pic.imgdb.cn/item/6177c5d62ab3f51d91e7a640.jpg)

The algorithm works roughly as follows:

1. A client sends a request to invoke a service operation to the primary  客户端向主服务器发送调用服务操作的请求
2. The primary multicasts the request to the backups 主服务器将请求广播给备份服务器
3. Replicas execute the request and send a reply to the client  副本执行请求并向客户机发送应答
4. The client waits for1 replies from different replicas with the same result; this is the result of the operation. 客户端等待来自f+1个不同副本的相同结果的回复;这是操作的结果。

本节的其余部分将描述该算法的简化版本。我们忽略了节点如何从缺乏空间的故障中恢复的讨论。我们还省略了与消息重传相关的细节。此外，我们假设消息认证是通过数字签名实现的。

### client

client c send to primary  request<o,t,c> (操作、时间戳、客户端)

副本发送给客户端的每个消息都包含当前视图号，允许客户端跟踪视图，从而跟踪当前主副本。客户端使用点对点消息向它认为的当前主服务器发送请求。主节点使用下一节中描述的协议对所有备份进行原子多播。

replica i send reply<v,t,c,i,r> (视图、时间戳、客户端、副本标识、执行结果)

客户端等待来自f+1个不同副本的有效签名回复以确保结果是有效的，因为在大多数副本可能是错误的。

如果客户端没有及时收到回复，它就向所有副本广播请求。如果请求已经被处理，副本只需重新发送应答;副本会记住他们发送给每个客户端的最后一个回复消息。否则，如果副本不是主服务器，它会将请求转发给主服务器。如果主服务器没有将请求多播到组中，它最终会被足够多的副本怀疑是错误的，从而导致视图更改。

在本文中，我们假设客户端在发送下一个请求之前等待一个请求完成。但是我们可以允许客户端发出异步请求，同时保留对它们的顺序约束。

### 一般情况下的操作

每个副本的状态包括服务状态、消息日志(包含副本已接受的消息)和副本当前视图的整数。当主服务器接收到客户端请求时，它启动一个三阶段协议，以原子方式将请求多播到副本。

三阶段：pre-prepare、prepare、commit。

> * pre-prepare和prepare阶段用于对在同一个视图中发送的请求进行完全排序，即使主视图(提出请求的排序)出错。
>
> * prepare和commit阶段用于确保提交的请求完全跨视图排序。

* pre-prepare阶段：主节点为request分配一个序列号n，并准备一个pre-prepare消息将其广播到其他副本，并将message加入日志中。

  pre-prepare消息<<v,n,d>,m>（视图、序列号、消息的摘要、request的消息）

* 副本检查pre-prepare消息：

  * request和pre-prepare信息中的签名是正确的，并且是请求的摘要;

  * v是当前的视图
  * 没有接受视图v和包含不同摘要的序列号的pre-prepare消息
  * pre-prepare消息中的序列号介于低水位标志h和高水位标志H之间（防止有故障的主节点耗尽序列号空间）

* 如果副本接受pre-prepare消息：副本准备一个prepare消息广播到其他副本，并将两个消息都加入日志中

  prepare消息<v,n,d,i>（视图、序列号、消息的摘要、当前副本）

* 副本检查prepare消息：

  * 消息都加入日志中
  * 消息正确签名
  * v是当前的视图
  * 消息中的序列号介于低水位标志h和高水位标志H之间
  * 收到2f个来自不同副本的prepare消息
  * 副本通过检查它们是否具有相同的视图、序列号和摘要来验证准备是否与预准备匹配。
  * ......更多细节参见论文

* 如果副本接受prepare消息：副本准备一个commit消息广播到其他副本

  commit消息<v,n,D(m),i>（视图、序列号、当前节点签名消息、当前副本）

* 副本检查commit消息：

  * 消息正确签名
  * 消息中的序列号介于低水位标志h和高水位标志H之间
  * 收到2f个来自不同副本的commit消息
  * ......更多细节参见论文

* 副本接受commit消息：副本准备一个reply消息发送到提出消息的client，告知其结果

* client验证结果：检查是否收到f+1个正确的reply结果

### 垃圾收集

### 视图切换

### 安全性证明



## 代码实现

> 大体参照 https://github.com/corgi-kx/blockchain_consensus_algorithm
>
> 对网络部分使用RPC进行替换
>
> 计划补充：
>
> * Leader的选举和轮换（视图切换）
> * 消息重传

  数据从客户端输入，到接收到节点们的回复共分为5步

  1. 客户端向主节点发送请求信息
  2. 主节点N0接收到客户端请求后将请求数据里的主要信息提出，并向其余节点进行preprepare发送
  3. 从节点们接收到来自主节点的preprepare，首先利用主节点的公钥进行签名认证，其次将消息进行散列（消息摘要，以便缩小信息在网络中的传输大小）后，向其他节点广播prepare
  4. 节点接收到2f个prepare信息（包含自己）,并全部签名验证通过，则可以进行到commit步骤，向全网其他节点广播commit
  5. 节点接收到2f+1个commit信息（包含自己），并全部签名验证通过，则可以把消息存入到本地，并向客户端返回reply消息



参考：

https://github.com/corgi-kx/blockchain_consensus_algorithm

[Practical Byzantine Fault Tolerance 论文](https://dblp.org/rec/conf/osdi/CastroL99)

https://www.cnblogs.com/xiaohuiduan/p/12210891.html

