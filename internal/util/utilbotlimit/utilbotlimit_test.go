package utilbotlimit

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkLimitManager1(b *testing.B) {
	rl := NewBotLimitManager(0, 0, 100) // defaults
	for i := 0; i < b.N; i++ {
		rl.createdAt = time.UnixMicro(0) // reset each time
		rl.LimitIPActivity("123")        // memory alloc
	}
}

func BenchmarkLimitManager2(b *testing.B) {
	rl := NewBotLimitManager(0, 0, 100) // defaults
	for i := 0; i < b.N; i++ {
		rl.resetData() // memory alloc
	}
}

func BenchmarkLimitIPActivity(b *testing.B) {
	rl := NewBotLimitManager(0, 0, 0) // defaults
	for i := 0; i < b.N; i++ {
		rl.LimitIPActivity("123") // memory alloc
	}
}
func TestLimitManager(t *testing.T) {
	t.Run("Single IP Rate Limiting", func(t *testing.T) {
		rl := NewBotLimitManager(1_000_000, time.Minute, 100) // defaults
		ip := "192.168.1.1"

		assert.Equal(t, 1_000_000, rl.memorySize)
		assert.Equal(t, time.Minute, rl.lifetime)
		assert.Equal(t, 100, int(rl.limit))

		// Should allow 0-99 requests
		for i := 0; i < 100; i++ {
			if rl.LimitIPActivity(ip) {
				t.Errorf("Expected request %d to be allowed, but was limited", i)
			}
		}

		// Should limit after exceeding threshold
		if !rl.LimitIPActivity(ip) {
			t.Error("Expected request to be limited after exceeding threshold")
		}

		if !rl.LimitIPActivity(ip) {
			t.Error("Expected subsequent request to also be limited")
		}

		rl.resetData() // test reset

		if rl.LimitIPActivity(ip) {
			t.Errorf("Expected request to be allowed, but was limited")
		}
	})

	t.Run("Multiple IPs Rate Limiting", func(t *testing.T) {
		rl := NewBotLimitManager(0, 0, 0)

		// Test with different IPs
		for i := 0; i < 1_000_000; i++ {
			ip := "192.168.1." + strconv.Itoa(i) // not real ip >= 255
			if rl.LimitIPActivity(ip) {
				t.Errorf("Expected request from IP %s to be allowed", ip)
			}
		}
	})
}
