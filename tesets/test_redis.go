package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Models struct {
	RedisClient *redis.Client
	PublishKey  string
}

func (m *Models) Subscribe(c context.Context, channel string) (string, error) {
	pubsub := m.RedisClient.Subscribe(c, channel)
	defer pubsub.Close()

	_, err := pubsub.Receive(c)
	if err != nil {
		return "", fmt.Errorf("subscribe receive failed: %v", err)
	}

	ch := pubsub.Channel()
	for msg := range ch {
		return msg.Payload, nil
	}
	return "", fmt.Errorf("no messages received")
}

func main() {
	models := &Models{
		RedisClient: redis.NewClient(&redis.Options{
			Addr:     "localhost:6378",
			Password: "123456",
			DB:       0,
		}),
		PublishKey: "your_publish_key",
	}

	ws := &WebSocketMock{}

	go func() {
		for {
			msg, err := models.Subscribe(ctx, models.PublishKey)
			if err != nil {
				fmt.Println("MsgHandler 发送失败:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			tm := time.Now().Format("2006-01-02 15:04:05")
			m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
			err = ws.WriteMessage(1, []byte(m))
			if err != nil {
				log.Fatalln("WebSocket WriteMessage error:", err)
			}
			fmt.Println("|service|user_service|讯息 ===> ", msg)
		}
	}()

	// 发布测试消息
	go func() {
		for {
			time.Sleep(5 * time.Second)
			err := models.RedisClient.Publish(ctx, models.PublishKey, "Hello, World!").Err()
			if err != nil {
				fmt.Println("Publish 失败:", err)
			} else {
				fmt.Println("消息已发布")
			}
		}
	}()

	select {} // 阻止 main 函数退出
}

// 模拟的 WebSocket 类型，用于演示
type WebSocketMock struct{}

func (w *WebSocketMock) WriteMessage(messageType int, data []byte) error {
	// 模拟写消息的逻辑
	fmt.Printf("Writing message: %s\n", data)
	return nil
}
