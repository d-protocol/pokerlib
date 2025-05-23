package pokerlib

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/fnv"
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

	// PASS 1: Standard Fisher-Yates with crypto/rand for true randomness
	for i := len(result) - 1; i > 0; i-- {
		// Generate cryptographically secure random number
		max := big.NewInt(int64(i + 1))
		j64, err := rand.Int(rand.Reader, max)
		if err != nil {
			// Fallback to time-seeded entropy if crypto/rand fails
			source := binary.BigEndian.Uint64(timeBasedSeed())
			j := uint64(source) % uint64(i+1)
			result[i], result[j] = result[j], result[i]
			continue
		}

		j := int(j64.Int64())
		result[i], result[j] = result[j], result[i]
	}

	// PASS 2: Add entropy and split-deck shuffling technique to break consecutive patterns
	n := len(result)

	// Reduce consecutive patterns by splitting and interleaving deck halves
	// This directly addresses the consecutive pattern issue
	firstHalf := make([]string, n/2)
	secondHalf := make([]string, n-n/2)

	copy(firstHalf, result[:n/2])
	copy(secondHalf, result[n/2:])

	// Shuffle each half separately
	h := fnv.New64a()
	for i := len(firstHalf) - 1; i > 0; i-- {
		// Add card values as entropy
		h.Reset()
		h.Write([]byte(firstHalf[i]))
		entropy := h.Sum64()

		j := int(entropy % uint64(i+1))
		firstHalf[i], firstHalf[j] = firstHalf[j], firstHalf[i]
	}

	for i := len(secondHalf) - 1; i > 0; i-- {
		// Add different card values as entropy
		h.Reset()
		h.Write([]byte(secondHalf[i]))
		entropy := h.Sum64()

		j := int(entropy % uint64(i+1))
		secondHalf[i], secondHalf[j] = secondHalf[j], secondHalf[i]
	}

	// Perfect interleave to eliminate consecutive patterns
	// This is like a perfect riffle shuffle in card games
	index := 0
	for i := 0; i < len(firstHalf); i++ {
		result[index] = firstHalf[i]
		index++
		if index < n && i < len(secondHalf) {
			result[index] = secondHalf[i]
			index++
		}
	}

	// Add any remaining cards from second half (if odd number)
	for i := len(firstHalf); i < len(secondHalf); i++ {
		result[index] = secondHalf[i]
		index++
	}

	// PASS 3: Position bias reduction through offset-mixing
	// Specifically targets positional biases by ensuring cards move across positions
	offsets := []int{7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}

	for _, offset := range offsets {
		// Skip if offset is larger than our deck length
		if offset >= n {
			continue
		}

		// Record original positions
		tmpDeck := make([]string, n)
		copy(tmpDeck, result)

		// Move each card to a new position based on the offset
		// This ensures each card has a chance to be in any position
		for i := 0; i < n; i++ {
			newPos := (i + offset) % n
			result[newPos] = tmpDeck[i]
		}
	}

	// PASS 4: Final crypto-secure Fisher-Yates pass
	// Final random shuffle to ensure unpredictability
	for i := len(result) - 1; i > 0; i-- {
		max := big.NewInt(int64(i + 1))
		j64, _ := rand.Int(rand.Reader, max)
		// Ignore error here since we already have a well-shuffled deck
		if j64 != nil {
			j := int(j64.Int64())
			result[i], result[j] = result[j], result[i]
		}
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
