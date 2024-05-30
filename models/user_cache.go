package models

import (
	"context"
	"time"
)

/**
設置在線用戶到redis緩存
**/
func SetUserOnlineInfo(key string, val []byte, timeTTL time.Duration) {
	ctx := context.Background()
	Red.Set(ctx, key, val, timeTTL)
}
