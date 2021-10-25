package pbft

/**
* @Description: Constant
* @author : kid1999
* @date Date : 2021/10/25 5:59 PM
* @version V1.0
 */

const (
	// 节点总数
	NodeCount = 4
	// leader
	LeaderID = "3"
)

// 节点table
var NodeTable = map[string]string{
	"0": "127.0.0.1:8000",
	"1": "127.0.0.1:8001",
	"2": "127.0.0.1:8002",
	"3": "127.0.0.1:8003",
}
