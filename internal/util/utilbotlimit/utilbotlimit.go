package utilbotlimit

import (
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
	"time"
)

// // hmacMD5ToUint64 262756 ns/op
// func hmacMD5ToUint64(data []byte, secret []byte) uint64 {
// 	h := hmac.New(md5.New, []byte(secret))
// 	_, _ = h.Write(data)   // Write the data to the hasher
// 	_, _ = h.Write(secret) // Write the data to the hasher
// 	hash := h.Sum(nil)
// 	// Use the first 8 bytes to create a uint64
// 	return binary.BigEndian.Uint64(hash[:8])
// }

// // hmacSHA256ToUint64 1144548 ns/op
// func hmacSHA256ToUint64(data []byte, secret []byte) uint64 {
// 	h := hmac.New(sha256.New, []byte(secret))
// 	_, _ = h.Write(data)   // Write the data to the hasher
// 	_, _ = h.Write(secret) // Write the data to the hasher
// 	hash := h.Sum(nil)
// 	// Use the first 8 bytes to create a uint64
// 	return binary.BigEndian.Uint64(hash[:8])
// }

// // sha256ToUint64 991.3 ns/op
//
//	func sha256ToUint64(data []byte, secret []byte) uint64 {
//		h := sha256.New()
//		_, _ = h.Write(data)   // Write the data to the hasher
//		_, _ = h.Write(secret) // Write the data to the hasher
//		hash := h.Sum(nil)
//		// Use the first 8 bytes to create a uint64
//		return binary.BigEndian.Uint64(hash[:8])
//	}
//
// // md5ToUint64 575.2 ns/op
// func md5ToUint64(data []byte, secret []byte) uint64 {
// 	h := md5.New()
// 	_, _ = h.Write(data)   // Write the data to the hasher
// 	_, _ = h.Write(secret) // Write the data to the hasher
// 	hash := h.Sum(nil)
// 	// Use the first 8 bytes to create a uint64
// 	return binary.BigEndian.Uint64(hash[:8])
// }

// fnvToUint64 371.4 ns/op
func fnvToUint64(data []byte, secret []byte) uint64 {
	h := fnv.New64a()      // Create a new FNV-1a 64-bit hash [Non-cryptographic hash]
	_, _ = h.Write(data)   // Write the data to the hasher
	_, _ = h.Write(secret) // Write the data to the hasher
	return h.Sum64()       // Return the hash as uint64
}

// HashBytesToUint64 computes a uint64 hash from a byte array
func hashBytesToUint64(data []byte, secret []byte) uint64 {

	return fnvToUint64(data, secret)
}

func mustGenerateRandomBytes(length int) []byte {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// BotLimitManager hash collision allowed
type BotLimitManager struct {
	mu         sync.Mutex
	createdAt  time.Time
	data       []byte
	secret     []byte // make index less predictable
	lifetime   time.Duration
	memorySize int
	limit      byte
	disabled   bool
}

func NewBotLimitManager(
	memorySize int,
	lifetime time.Duration,
	limit int,
) *BotLimitManager {

	if lifetime <= 0 {
		lifetime = 5 * time.Minute // 5 minutes
	}

	if memorySize <= 0 {
		memorySize = 1_000_000 // 1 mb
	}
	if limit <= 0 {
		limit = math.MaxUint8
	}
	return &BotLimitManager{

		memorySize: memorySize,
		lifetime:   lifetime,
		limit:      byte(min(limit, math.MaxUint8)), // 255
	}
}

var NoLimitManager = BotLimitManager{disabled: true}

// reset reset after lifetime
func (x *BotLimitManager) resetDataIfOld() {
	isDataOld := time.Now().UTC().Sub(x.createdAt) > x.lifetime
	if isDataOld {
		x.resetData()
	}
}

// resetForce reset after lifetime
func (x *BotLimitManager) resetData() {

	x.data = make([]byte, x.memorySize)
	x.createdAt = time.Now().UTC()
	x.secret = mustGenerateRandomBytes(10)
}

// limitActivity not a robot
func (x *BotLimitManager) limitActivity(key string) bool {

	if x.disabled {
		return false
	}

	x.mu.Lock()
	defer x.mu.Unlock()

	x.resetDataIfOld()

	hash := hashBytesToUint64([]byte(key), x.secret)
	indx := hash % uint64(len(x.data))

	hits := x.data[indx]

	if hits == x.limit /*255*/ {
		return true
	}

	x.data[indx] = hits + 1

	return false
}

// LimitIPActivity not a robot
func (x *BotLimitManager) LimitIPActivity(ipAddress string) bool {
	return x.limitActivity(fmt.Sprintf("LimitIPActivity %s", ipAddress))
}

// LimitSignupActivity signup with email
func (x *BotLimitManager) LimitSignupActivity(inboxAddress string) bool {
	return x.limitActivity(fmt.Sprintf("LimitSignupActivity %s", inboxAddress))
}

// LimitSignupMessage signup, sending secret via sms
func (x *BotLimitManager) LimitSignupMessage(inboxAddress string) bool {
	return x.limitActivity(fmt.Sprintf("LimitSignupMessage %s", inboxAddress))
}

// LimitAccountAccess any access to account
func (x *BotLimitManager) LimitAccountAccess(userID string) bool {
	return x.limitActivity(fmt.Sprintf("LimitAccountAccess %s", userID))
}

// LimitUserMessage forgot pw
func (x *BotLimitManager) LimitUserMessage(inboxAddress string, userID string) bool {
	return x.limitActivity(fmt.Sprintf("LimitUserMessage %s %s", inboxAddress, userID))
}