package policy

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func IsJwtInvalidByJTI(jti string, rdb *redis.Client) bool {
	ctx := context.Background()
	ok, _ := rdb.SIsMember(ctx, "tokens", jti).Result()
	if !ok {
		return false
	}
	return true
}

func InvalidateJwtByJTI(jti string, rdb *redis.Client) bool {
	ctx := context.Background()
	num, err := rdb.SAdd(ctx, "tokens", jti).Result()
	if num != 1 || err != nil {
		return false
	}
	return true
}

func IsJwtValidByVersion(version string) bool {
	return true
}

func InvalidateJwtByVersion(version string) bool {
	return true
}
