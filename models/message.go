package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0" // 查一下說明跟用法
	"gorm.io/gorm"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// 消息
type Message struct {
	gorm.Model
	UserId     int64  // 發送者
	TargetId   int64  // 接收者
	Type       int    // 發送類型  1私聊  2群聊  3心跳
	Media      int    // 消息類型  1文字 2表情包 3語音 4圖片 /表情包
	Content    string // 消息內容
	CreateTime uint64 // 創建時間
	ReadTime   uint64 // 讀取時間
	Pic        string
	Url        string
	Desc       string
	Amount     int // 其他數字統計
}

func (table *Message) TableName() string {
	return "message"
}

// 建立一個節點的結構體
type Node struct {
	Conn          *websocket.Conn // 連接
	Addr          string          // 客戶端地址
	FirstTime     uint64          // 首次連接時間
	HeartbeatTime uint64          // 心跳時間
	LoginTime     uint64          // 登錄時間
	DataQueue     chan []byte     // 消息
	GroupSets     set.Interface   // 好友 / 群
}

// 建立 map 的映射關係
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 建立讀寫鎖
var rwLocker sync.RWMutex

// 需要 : 發送者 ID，接收者 ID，消息類型，發送的內容，發送類型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.  獲取參數 並 檢驗 token 等合法性
	//token := query.Get("token")
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//	context := query.Get("context")
	isvalida := true //checkToke()  待.........
	conn, err := (&websocket.Upgrader{
		//token 校驗
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//2.獲取conn
	currentTime := uint64(time.Now().Unix())
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(), // 客戶端地址
		HeartbeatTime: currentTime,                // 心跳時間
		LoginTime:     currentTime,                // 登錄時間
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}
	//3. 用戶關係
	//4. userid 跟 node 綁定 並加鎖
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	//5.完成發送邏輯
	go sendProc(node)
	//6.完成接收邏輯
	go recvProc(node)
	//7.加入在線用戶到緩存 (此部份關係到 sendProc(node), 因取不到緩存，所以就沒有目標用戶進行發送)
	SetUserOnlineInfo("online_"+Id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)

	//sendMsg(userId, []byte("Welcome to live MSN"))
}

// 5.完成發送邏輯
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue: // 將通道的資料讀取進 data
			fmt.Println("[ws]sendProc >>>> msg :", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 6.完成接收邏輯
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println("recvProc ==>", err)
		}
		// 心跳檢測 msg.Media == -1 || msg.Type == 3
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
		} else {
			dispatch(data)
			broadMsg(data) //todo 將消息廣播到局域網
			fmt.Println("[ws] recvProc <<<<< ", string(data))
		}

	}
}

// 建立一個通(管)道
var udpsendChan chan []byte = make(chan []byte, 1024)

// 將消息廣播到局域網
func broadMsg(data []byte) {
	udpsendChan <- data // 將資料送入通道中
}

// 初始化收、發協程
func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine ")
}

// 完成 udp 數據發送協程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255), // 走路由的網段地址 ?
		Port: viper.GetInt("port.udp"),
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpSendProc  data :", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

// 完成 udp 數據接收協程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: viper.GetInt("port.udp"),
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc  data :", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 後端調度邏輯處理
func dispatch(data []byte) {
	msg := Message{}
	msg.CreateTime = uint64(time.Now().Unix())
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: // 私信
		fmt.Println("dispatch  data :", string(data))
		sendMsg(msg.TargetId, data)
	case 2: // 群發
		//sendGroupMsg(msg.TargetId, data) // 發送的群ID ，消息內容
		// case 4: // 心跳
		// 	node.Heartbeat()
		//case 4:
		//
	}
}

// 群發消息
//func sendGroupMsg(targetId int64, msg []byte) {
//	fmt.Println("開始群發消息")
//	userIds := SearchUserByGroupId(uint(targetId))
//	for i := 0; i < len(userIds); i++ {
//		// 排除給自已的
//		if targetId != int64(userIds[i]) {
//			sendMsg(int64(userIds[i]), msg)
//		}
//
//	}
//}

// 私信發送
func sendMsg(userId int64, msg []byte) {

	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println("redis error:::", err)
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userID: ", userId, "  msg:", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	ress, e := Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() // 備用 後續擴展 記錄完整msg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println("sendMsg -> Red.ZAdd : ", ress)
}

// 需更改此方法才能完整的 msg 轉 byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// 取得湲存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Red.ZRevRange(ctx, key, 0, 10).Result()  //根据score倒叙

	var rels []string
	var err error
	if isRev {
		// 取得 init.go 中的 Red 實例
		rels, err = Red.ZRange(ctx, key, start, end).Result()
	} else {
		// 取得 init.go 中的 Red 實例
		rels, err = Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) // 沒有找到
	}
	// 發送推送消息
	/**
	// 後台通過 websoket 推送消息
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}

// 更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}
