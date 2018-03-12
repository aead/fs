package sf

import (
	mrand "math/rand"
	"time"
)

// Random is non-cryptographic pseudo-random-number-generator.
type Random struct{ mrand.Rand }

// NewRandom creates a new RNG from the provided seed.
func NewRandom(seed int64) Random {
	r := mrand.New(mrand.NewSource(seed))
	return Random{*r}
}

// AlphaString returns a random string of the requested length
// containing only characters in [0-9a-zA-Z].
func (r Random) AlphaString(length int) string {
	const aplhaNum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str := make([]byte, length)
	for i := range str {
		str[i] = aplhaNum[r.Int()%len(aplhaNum)]
	}
	return string(str)
}

// Date returns a random date in the future.
func (r Random) Date() time.Time {
	return time.Now().Add(time.Duration(r.Int63()))
}

// DateIn returns a random date between after (inclusive) and before (exclusive).
// If before is in the past or present of after DateIn panics.
func (r Random) DateIn(after time.Time, before time.Time) time.Time {
	duration := before.Sub(after)
	if duration <= 0 {
		panic("sf: 'before' cannot be in the past or present of 'after'")
	}
	shift := time.Duration(r.Int63()) % duration
	return after.Add(shift)
}
