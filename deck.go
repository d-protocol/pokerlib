package pokerlib

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

type CardSuit int32

const (
	CardSuitSpade CardSuit = iota
	CardSuitHeart
	CardSuitDiamond
	CardSuitClub
)

var CardSuits = []string{
	"S",
	"H",
	"D",
	"C",
}

var CardPoints = []string{
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"8",
	"9",
	"T",
	"J",
	"Q",
	"K",
	"A",
}

func NewStandardDeckCards() []string {

	cards := make([]string, 0, 52)

	for _, suit := range CardSuits {
		for i := 0; i < 13; i++ {
			cards = append(cards, fmt.Sprintf("%s%s", suit, CardPoints[i]))
		}
	}

	return cards
}

func NewShortDeckCards() []string {

	cards := make([]string, 0, 36)

	for _, suit := range CardSuits {

		// Take off 2, 3, 4 and 5
		for i := 4; i < 13; i++ {
			cards = append(cards, fmt.Sprintf("%s%s", suit, CardPoints[i]))
		}
	}

	return cards
}

func ShuffleCards(cards []string) []string {
	// Create a copy of the original cards to avoid modifying the input slice
	result := make([]string, len(cards))
	copy(result, cards)

	// Use Fisher-Yates algorithm with crypto/rand for true randomness
	for i := len(result) - 1; i > 0; i-- {
		// Generate cryptographically secure random number
		// We use crypto/rand instead of math/rand for better randomness
		max := big.NewInt(int64(i + 1))
		j64, err := rand.Int(rand.Reader, max)
		if err != nil {
			// Fallback to time-seeded entropy if crypto/rand fails
			// This creates a different seed each time by combining time elements
			source := binary.BigEndian.Uint64(timeBasedSeed())
			j := uint64(source) % uint64(i+1)
			result[i], result[j] = result[j], result[i]
			continue
		}

		j := int(j64.Int64())
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// timeBasedSeed creates a seed using multiple time sources to increase entropy
func timeBasedSeed() []byte {
	now := time.Now()
	seed := make([]byte, 8)
	// Mix nano time with unix time and monotonic clock for better unpredictability
	value := now.UnixNano() + now.Unix() + now.UnixMicro()
	binary.BigEndian.PutUint64(seed, uint64(value))
	return seed
}
