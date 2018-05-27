package walls

import (
	"fmt"
	"testing"
	"time"
)

import (
	"github.com/ihexxa/quickshare/server/libs/cfg"
	"github.com/ihexxa/quickshare/server/libs/encrypt"
	"github.com/ihexxa/quickshare/server/libs/limiter"
)

func newAccessWalls(limiterCap int64, limiterTtl int32, limiterCyc int32, bucketCap int16) *AccessWalls {
	config := cfg.NewConfig()
	config.Production = true
	config.LimiterCap = limiterCap
	config.LimiterTtl = limiterTtl
	config.LimiterCyc = limiterCyc
	config.BucketCap = bucketCap
	encrypterMaker := encrypt.JwtEncrypterMaker
	ipLimiter := limiter.NewRateLimiter(config.LimiterCap, config.LimiterTtl, config.LimiterCyc, config.BucketCap, map[int16]int16{})
	opLimiter := limiter.NewRateLimiter(config.LimiterCap, config.LimiterTtl, config.LimiterCyc, config.BucketCap, map[int16]int16{})

	return NewAccessWalls(config, ipLimiter, opLimiter, encrypterMaker).(*AccessWalls)
}
func TestIpLimit(t *testing.T) {
	ip := "0.0.0.0"
	limit := int16(10)
	ttl := int32(60)
	cyc := int32(5)
	walls := newAccessWalls(1000, ttl, cyc, limit)

	testIpLimit(t, walls, ip, limit)
	// wait for tokens are re-fullfilled
	time.Sleep(time.Duration(cyc) * time.Second)
	testIpLimit(t, walls, ip, limit)

	fmt.Println("ip limit: passed")
}

func testIpLimit(t *testing.T, walls Walls, ip string, limit int16) {
	for i := int16(0); i < limit; i++ {
		if !walls.PassIpLimit(ip) {
			t.Fatalf("ipLimiter: should be passed", time.Now().Unix())
		}
	}

	if walls.PassIpLimit(ip) {
		t.Fatalf("ipLimiter: should not be passed", time.Now().Unix())
	}
}

func TestOpLimit(t *testing.T) {
	resourceId := "id"
	op1 := int16(1)
	op2 := int16(2)
	limit := int16(10)
	ttl := int32(1)
	walls := newAccessWalls(1000, 5, ttl, limit)

	testOpLimit(t, walls, resourceId, op1, limit)
	testOpLimit(t, walls, resourceId, op2, limit)
	// wait for tokens are re-fullfilled
	time.Sleep(time.Duration(ttl) * time.Second)
	testOpLimit(t, walls, resourceId, op1, limit)
	testOpLimit(t, walls, resourceId, op2, limit)

	fmt.Println("op limit: passed")
}

func testOpLimit(t *testing.T, walls Walls, resourceId string, op int16, limit int16) {
	for i := int16(0); i < limit; i++ {
		if !walls.PassOpLimit(resourceId, op) {
			t.Fatalf("opLimiter: should be passed")
		}
	}

	if walls.PassOpLimit(resourceId, op) {
		t.Fatalf("opLimiter: should not be passed")
	}
}

func TestLoginCheck(t *testing.T) {
	walls := newAccessWalls(1000, 5, 1, 10)

	testValidToken(t, walls)
	testInvalidAdminIdToken(t, walls)
	testExpiredToken(t, walls)
}

func testValidToken(t *testing.T, walls *AccessWalls) {
	config := cfg.NewConfig()

	tokenMaker := encrypt.JwtEncrypterMaker(string(config.SecretKeyByte))
	tokenMaker.Add(config.KeyAdminId, config.AdminId)
	tokenMaker.Add(config.KeyExpires, fmt.Sprintf("%d", time.Now().Unix()+int64(10)))
	tokenStr, getTokenOk := tokenMaker.ToStr()
	if !getTokenOk {
		t.Fatalf("passLoginCheck: fail to generate token")
	}

	if !walls.passLoginCheck(tokenStr) {
		t.Fatalf("loginCheck: should be passed")
	}

	fmt.Println("loginCheck: valid token passed")
}

func testInvalidAdminIdToken(t *testing.T, walls *AccessWalls) {
	config := cfg.NewConfig()

	tokenMaker := encrypt.JwtEncrypterMaker(string(config.SecretKeyByte))
	tokenMaker.Add(config.KeyAdminId, "invalid admin id")
	tokenMaker.Add(config.KeyExpires, fmt.Sprintf("%d", time.Now().Unix()+int64(10)))
	tokenStr, getTokenOk := tokenMaker.ToStr()
	if !getTokenOk {
		t.Fatalf("passLoginCheck: fail to generate token")
	}

	if walls.passLoginCheck(tokenStr) {
		t.Fatalf("loginCheck: should not be passed")
	}

	fmt.Println("loginCheck: invalid admin id passed")
}

func testExpiredToken(t *testing.T, walls *AccessWalls) {
	config := cfg.NewConfig()

	tokenMaker := encrypt.JwtEncrypterMaker(string(config.SecretKeyByte))
	tokenMaker.Add(config.KeyAdminId, config.AdminId)
	tokenMaker.Add(config.KeyExpires, fmt.Sprintf("%d", time.Now().Unix()-int64(1)))
	tokenStr, getTokenOk := tokenMaker.ToStr()
	if !getTokenOk {
		t.Fatalf("passLoginCheck: fail to generate token")
	}

	if walls.passLoginCheck(tokenStr) {
		t.Fatalf("loginCheck: should not be passed")
	}

	fmt.Println("loginCheck: expired token passed")
}
