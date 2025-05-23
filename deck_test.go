package pokerlib

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
)

// TestShuffleCardDistribution runs a series of simulations to verify
// that the shuffle algorithm produces sufficiently random distributions
func TestShuffleCardDistribution(t *testing.T) {
	// Number of simulations to run
	simCount := 1000

	// Track how many times each card appears in each position
	cardPositionCount := make(map[string]map[int]int)

	// Initialize counters for each card
	for _, suit := range CardSuits {
		for _, rank := range CardPoints {
			card := fmt.Sprintf("%s%s", suit, rank)
			cardPositionCount[card] = make(map[int]int)
		}
	}

	// Run multiple simulations
	for i := 0; i < simCount; i++ {
		// Create a standard deck
		deck := NewStandardDeckCards()

		// Shuffle the deck
		shuffled := ShuffleCards(deck)

		// Count each card's position in this shuffle
		for pos, card := range shuffled {
			cardPositionCount[card][pos]++
		}
	}

	// Check for uniform distribution (each card should appear in each position roughly simCount/52 times)
	expectedPerPosition := float64(simCount) / 52.0

	// Check for deviations
	deviations := 0
	for card, positions := range cardPositionCount {
		for pos, count := range positions {
			deviation := float64(count) / expectedPerPosition
			if deviation > 1.3 || deviation < 0.7 {
				deviations++
				t.Logf("Card %s at position %d: count=%d, expected=%.2f, deviation=%.2f",
					card, pos, count, expectedPerPosition, deviation)
			}
		}
	}

	// Log overall statistics
	t.Logf("Expected occurrences per position: %.2f", expectedPerPosition)
	t.Logf("Total deviations outside 30%% range: %d", deviations)
	t.Logf("Deviation percentage: %.2f%%", float64(deviations)*100.0/float64(52*52))

	// Fail the test if we have significant deviations (more than 5% of possible positions)
	maxAllowedDeviations := int(math.Floor(float64(52*52) * 0.05)) // 5% of all card-position combinations
	if deviations > maxAllowedDeviations {
		t.Errorf("Shuffle algorithm shows significant position bias: %d deviations", deviations)
	}
}

// TestShuffleHandFrequency creates multiple poker hands and checks for duplicate hand types
func TestShuffleHandFrequency(t *testing.T) {
	// Number of games to simulate
	gameCount := 500

	// Track hand type frequencies
	similarHandCount := 0
	totalHandCount := 0

	// Run simulations
	for i := 0; i < gameCount; i++ {
		// Create and shuffle a deck
		deck := NewStandardDeckCards()
		shuffled := ShuffleCards(deck)

		// Deal cards to players (7 players, 2 cards each)
		playerHands := make([][]string, 7)
		for p := 0; p < 7; p++ {
			playerHands[p] = []string{
				shuffled[p*2],
				shuffled[p*2+1],
			}
		}

		// Deal community cards
		communityCards := []string{
			shuffled[14], // Flop 1
			shuffled[15], // Flop 2
			shuffled[16], // Flop 3
			shuffled[17], // Turn
			shuffled[18], // River
		}

		// Evaluate each player's hand
		handTypes := make([]string, 7)
		for p := 0; p < 7; p++ {
			handTypes[p] = evaluateHand(playerHands[p], communityCards)
		}

		// Count duplicate hand types
		for i := 0; i < len(handTypes); i++ {
			for j := i + 1; j < len(handTypes); j++ {
				if handTypes[i] == handTypes[j] {
					similarHandCount++
				}
				totalHandCount++
			}
		}
	}

	// Calculate and report statistics
	similarHandPercentage := float64(similarHandCount) * 100.0 / float64(totalHandCount)
	t.Logf("Similar hand percentage: %.2f%% (%d out of %d comparisons)",
		similarHandPercentage, similarHandCount, totalHandCount)

	// In completely random distribution, some similarity is expected
	// but it should be below a threshold
	if similarHandPercentage > 25.0 {
		t.Errorf("Too many similar hands: %.2f%% (expected less than 25%%)", similarHandPercentage)
	}
}

// Simple hand evaluator that returns the hand type
func evaluateHand(holeCards, communityCards []string) string {
	// Combine hole cards and community cards
	allCards := append([]string{}, holeCards...)
	allCards = append(allCards, communityCards...)

	// Count suits for flush detection
	suitCount := make(map[string]int)
	for _, card := range allCards {
		if len(card) > 0 {
			suit := string(card[0])
			suitCount[suit]++
		}
	}

	// Check for flush
	hasFlush := false
	for _, count := range suitCount {
		if count >= 5 {
			hasFlush = true
			break
		}
	}

	// Count ranks for pairs, etc.
	rankCount := make(map[string]int)
	for _, card := range allCards {
		if len(card) >= 2 {
			rank := string(card[1])
			rankCount[rank]++
		}
	}

	// Determine hand type
	if hasFlush {
		return "Flush"
	}

	fourOfAKind := false
	threeOfAKind := false
	pairCount := 0

	for _, count := range rankCount {
		if count == 4 {
			fourOfAKind = true
		} else if count == 3 {
			threeOfAKind = true
		} else if count == 2 {
			pairCount++
		}
	}

	if fourOfAKind {
		return "Four of a Kind"
	} else if threeOfAKind && pairCount > 0 {
		return "Full House"
	} else if threeOfAKind {
		return "Three of a Kind"
	} else if pairCount >= 2 {
		return "Two Pair"
	} else if pairCount == 1 {
		return "Pair"
	}

	return "High Card"
}

// TestCompareShufflingMethods compares the original and improved shuffling methods
func TestCompareShufflingMethods(t *testing.T) {
	// Create a visual demonstration of shuffling quality
	t.Log("Running comparative distribution test between shuffling methods")

	// Function to implement original (simpler) Fisher-Yates shuffle
	originalShuffle := func(cards []string) []string {
		result := make([]string, len(cards))
		copy(result, cards)

		// Basic Fisher-Yates with a single pass
		for i := len(result) - 1; i > 0; i-- {
			// Use crypto/rand but with a single pass only
			max := big.NewInt(int64(i + 1))
			j64, err := rand.Int(rand.Reader, max)
			if err != nil {
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

	// Run a higher number of poker games with each method for statistical significance
	gameCount := 1000
	playerCount := 7 // Number of players per game

	// Track duplicate hand types for each method
	origSimilarHandCount := 0
	newSimilarHandCount := 0
	handComparisons := 0

	// Also track consecutive card patterns
	origConsecutivePatterns := 0
	newConsecutivePatterns := 0

	// Track different types of hands
	origHandTypeDistribution := make(map[string]int)
	newHandTypeDistribution := make(map[string]int)

	// Track winning hand types
	origWinningHandTypes := make(map[string]int)
	newWinningHandTypes := make(map[string]int)

	// For Kolmogorov-Smirnov test: track position distributions
	origPositionDistribution := make(map[string]map[int]int)
	newPositionDistribution := make(map[string]map[int]int)

	// Initialize position distribution maps
	for _, suit := range CardSuits {
		for _, rank := range CardPoints {
			card := fmt.Sprintf("%s%s", suit, rank)
			origPositionDistribution[card] = make(map[int]int)
			newPositionDistribution[card] = make(map[int]int)
		}
	}

	for i := 0; i < gameCount; i++ {
		// Create a standard deck
		deck := NewStandardDeckCards()

		// Shuffle with both methods
		origShuffled := originalShuffle(deck)
		newShuffled := ShuffleCards(deck)

		// Track card positions for distribution analysis
		for pos, card := range origShuffled {
			if _, exists := origPositionDistribution[card]; exists {
				origPositionDistribution[card][pos]++
			}
		}
		for pos, card := range newShuffled {
			if _, exists := newPositionDistribution[card]; exists {
				newPositionDistribution[card][pos]++
			}
		}

		// Deal cards for both methods (players, 2 cards each + 5 community)
		origPlayerHands := make([][]string, playerCount)
		newPlayerHands := make([][]string, playerCount)

		for p := 0; p < playerCount; p++ {
			origPlayerHands[p] = []string{
				origShuffled[p*2],
				origShuffled[p*2+1],
			}

			newPlayerHands[p] = []string{
				newShuffled[p*2],
				newShuffled[p*2+1],
			}
		}

		// Community cards
		origCommunity := []string{
			origShuffled[playerCount*2],   // Flop 1
			origShuffled[playerCount*2+1], // Flop 2
			origShuffled[playerCount*2+2], // Flop 3
			origShuffled[playerCount*2+3], // Turn
			origShuffled[playerCount*2+4], // River
		}

		newCommunity := []string{
			newShuffled[playerCount*2],   // Flop 1
			newShuffled[playerCount*2+1], // Flop 2
			newShuffled[playerCount*2+2], // Flop 3
			newShuffled[playerCount*2+3], // Turn
			newShuffled[playerCount*2+4], // River
		}

		// Check for consecutive card patterns (cards of same rank or suit in sequence)
		for i := 0; i < len(origShuffled)-1; i++ {
			// Check if consecutive cards have same suit or rank
			origCard1, origCard2 := origShuffled[i], origShuffled[i+1]
			newCard1, newCard2 := newShuffled[i], newShuffled[i+1]

			// Same suit or rank in original shuffle
			if origCard1[0] == origCard2[0] || (len(origCard1) > 1 && len(origCard2) > 1 && origCard1[1] == origCard2[1]) {
				origConsecutivePatterns++
			}

			// Same suit or rank in new shuffle
			if newCard1[0] == newCard2[0] || (len(newCard1) > 1 && len(newCard2) > 1 && newCard1[1] == newCard2[1]) {
				newConsecutivePatterns++
			}
		}

		// Evaluate hands for both methods
		origHandTypes := make([]string, playerCount)
		newHandTypes := make([]string, playerCount)

		// Evaluate hands with strengths for determining winners
		origHandStrengths := make([]float64, playerCount)
		newHandStrengths := make([]float64, playerCount)

		for p := 0; p < playerCount; p++ {
			// Evaluate original hands
			origHandTypes[p] = evaluateHand(origPlayerHands[p], origCommunity)
			origHandStrengths[p] = getHandStrength(origHandTypes[p])

			// Evaluate new hands
			newHandTypes[p] = evaluateHand(newPlayerHands[p], newCommunity)
			newHandStrengths[p] = getHandStrength(newHandTypes[p])

			// Track distribution of hand types
			origHandTypeDistribution[origHandTypes[p]]++
			newHandTypeDistribution[newHandTypes[p]]++
		}

		// Find winners for original shuffle
		origWinnerIndices := findWinners(origHandStrengths)
		for _, winnerIdx := range origWinnerIndices {
			winningType := origHandTypes[winnerIdx]
			origWinningHandTypes[winningType]++
		}

		// Find winners for improved shuffle
		newWinnerIndices := findWinners(newHandStrengths)
		for _, winnerIdx := range newWinnerIndices {
			winningType := newHandTypes[winnerIdx]
			newWinningHandTypes[winningType]++
		}

		// Count duplicate hand types
		for i := 0; i < len(origHandTypes); i++ {
			for j := i + 1; j < len(origHandTypes); j++ {
				if origHandTypes[i] == origHandTypes[j] {
					origSimilarHandCount++
				}

				if newHandTypes[i] == newHandTypes[j] {
					newSimilarHandCount++
				}

				handComparisons++
			}
		}

		// For the first few games, print the actual hands for visual inspection
		if i < 3 {
			t.Logf("Game %d Original Shuffle:", i+1)
			t.Logf("  Community: %v", origCommunity)

			for p := 0; p < playerCount; p++ {
				isWinner := false
				for _, winnerIdx := range origWinnerIndices {
					if p == winnerIdx {
						isWinner = true
						break
					}
				}

				if isWinner {
					t.Logf("  Player %d: %v - %s (WINNER)", p+1, origPlayerHands[p], origHandTypes[p])
				} else {
					t.Logf("  Player %d: %v - %s", p+1, origPlayerHands[p], origHandTypes[p])
				}
			}

			t.Logf("Game %d Improved Shuffle:", i+1)
			t.Logf("  Community: %v", newCommunity)

			for p := 0; p < playerCount; p++ {
				isWinner := false
				for _, winnerIdx := range newWinnerIndices {
					if p == winnerIdx {
						isWinner = true
						break
					}
				}

				if isWinner {
					t.Logf("  Player %d: %v - %s (WINNER)", p+1, newPlayerHands[p], newHandTypes[p])
				} else {
					t.Logf("  Player %d: %v - %s", p+1, newPlayerHands[p], newHandTypes[p])
				}
			}
			t.Logf("--------------------------")
		}
	}

	// Calculate statistics
	origSimilarHandPercentage := float64(origSimilarHandCount) * 100.0 / float64(handComparisons)
	newSimilarHandPercentage := float64(newSimilarHandCount) * 100.0 / float64(handComparisons)

	origPatternPerDeck := float64(origConsecutivePatterns) / float64(gameCount)
	newPatternPerDeck := float64(newConsecutivePatterns) / float64(gameCount)

	// Calculate positional bias
	origPositionalDeviations := 0
	newPositionalDeviations := 0
	expectedPerPosition := float64(gameCount) / 52.0

	for _, positions := range origPositionDistribution {
		for _, count := range positions {
			deviation := math.Abs(float64(count)-expectedPerPosition) / expectedPerPosition
			if deviation > 0.3 { // More than 30% deviation from expected
				origPositionalDeviations++
			}
		}
	}

	for _, positions := range newPositionDistribution {
		for _, count := range positions {
			deviation := math.Abs(float64(count)-expectedPerPosition) / expectedPerPosition
			if deviation > 0.3 { // More than 30% deviation from expected
				newPositionalDeviations++
			}
		}
	}

	// Report results
	t.Logf("STATISTICS BASED ON %d GAMES WITH %d PLAYERS:", gameCount, playerCount)
	t.Logf("")
	t.Logf("ORIGINAL SHUFFLE:")
	t.Logf("  Similar hand percentage: %.2f%% (%d out of %d comparisons)",
		origSimilarHandPercentage, origSimilarHandCount, handComparisons)
	t.Logf("  Consecutive card patterns per deck: %.2f", origPatternPerDeck)
	t.Logf("  Position bias deviations: %d (%.2f%%)",
		origPositionalDeviations, float64(origPositionalDeviations)*100.0/float64(52*52))

	t.Logf("")
	t.Logf("IMPROVED SHUFFLE:")
	t.Logf("  Similar hand percentage: %.2f%% (%d out of %d comparisons)",
		newSimilarHandPercentage, newSimilarHandCount, handComparisons)
	t.Logf("  Consecutive card patterns per deck: %.2f", newPatternPerDeck)
	t.Logf("  Position bias deviations: %d (%.2f%%)",
		newPositionalDeviations, float64(newPositionalDeviations)*100.0/float64(52*52))

	t.Logf("")
	t.Logf("IMPROVEMENT:")
	t.Logf("  Similar hand reduction: %.2f%%",
		100.0-(newSimilarHandPercentage*100.0/origSimilarHandPercentage))
	t.Logf("  Consecutive pattern reduction: %.2f%%",
		100.0-(newPatternPerDeck*100.0/origPatternPerDeck))
	t.Logf("  Position bias reduction: %.2f%%",
		100.0-(float64(newPositionalDeviations)*100.0/float64(origPositionalDeviations)))

	// Hand type distribution analysis
	t.Logf("")
	t.Logf("HAND TYPE DISTRIBUTION:")
	t.Logf("  Type         | Original |  Improved  | Difference")
	t.Logf("  -------------|----------|------------|------------")
	totalOrig := 0
	totalNew := 0
	for handType := range origHandTypeDistribution {
		totalOrig += origHandTypeDistribution[handType]
	}
	for handType := range newHandTypeDistribution {
		totalNew += newHandTypeDistribution[handType]
	}

	handTypes := []string{
		"High Card",
		"Pair",
		"Two Pair",
		"Three of a Kind",
		"Straight",
		"Flush",
		"Full House",
		"Four of a Kind",
	}

	for _, handType := range handTypes {
		origCount := 0
		newCount := 0

		// Count all matching types (prefix matching)
		for ht, count := range origHandTypeDistribution {
			if strings.HasPrefix(ht, handType) {
				origCount += count
			}
		}
		for ht, count := range newHandTypeDistribution {
			if strings.HasPrefix(ht, handType) {
				newCount += count
			}
		}

		origPct := float64(origCount) * 100.0 / float64(totalOrig)
		newPct := float64(newCount) * 100.0 / float64(totalNew)
		diff := newPct - origPct

		t.Logf("  %-12s| %6.2f%% | %6.2f%%   | %+6.2f%%",
			handType, origPct, newPct, diff)
	}

	// Winning hand type distribution analysis
	t.Logf("")
	t.Logf("WINNING HAND TYPE DISTRIBUTION:")
	t.Logf("  Type         | Original |  Improved  | Difference")
	t.Logf("  -------------|----------|------------|------------")
	totalOrigWins := 0
	totalNewWins := 0
	for handType := range origWinningHandTypes {
		totalOrigWins += origWinningHandTypes[handType]
	}
	for handType := range newWinningHandTypes {
		totalNewWins += newWinningHandTypes[handType]
	}

	for _, handType := range handTypes {
		origWinCount := 0
		newWinCount := 0

		// Count all matching types (prefix matching)
		for ht, count := range origWinningHandTypes {
			if strings.HasPrefix(ht, handType) {
				origWinCount += count
			}
		}
		for ht, count := range newWinningHandTypes {
			if strings.HasPrefix(ht, handType) {
				newWinCount += count
			}
		}

		// Skip if no wins of this type
		if origWinCount == 0 && newWinCount == 0 {
			continue
		}

		origWinPct := float64(origWinCount) * 100.0 / float64(totalOrigWins)
		newWinPct := float64(newWinCount) * 100.0 / float64(totalNewWins)
		winDiff := newWinPct - origWinPct

		t.Logf("  %-12s| %6.2f%% | %6.2f%%   | %+6.2f%%",
			handType, origWinPct, newWinPct, winDiff)
	}

	// Assert improvements - allow up to 10% randomness fluctuation
	if newSimilarHandPercentage > origSimilarHandPercentage*1.1 {
		t.Errorf("Improved shuffle significantly increased similar hands by more than 10%%")
	}

	if float64(newPatternPerDeck) > float64(origPatternPerDeck)*1.1 {
		t.Errorf("Improved shuffle significantly increased consecutive patterns by more than 10%%")
	}

	if float64(newPositionalDeviations) > float64(origPositionalDeviations)*1.1 {
		t.Errorf("Improved shuffle significantly increased position bias by more than 10%%")
	}
}

// getHandStrength returns a numerical strength for hand comparison
func getHandStrength(handType string) float64 {
	// Hand type strengths (higher value = stronger hand)
	handStrengths := map[string]float64{
		"High Card":       1.0,
		"Pair":            2.0,
		"Two Pair":        3.0,
		"Three of a Kind": 4.0,
		"Straight":        5.0,
		"Flush":           6.0,
		"Full House":      7.0,
		"Four of a Kind":  8.0,
		"Straight Flush":  9.0,
		"Royal Flush":     10.0,
	}

	// Extract base hand type without specifics
	baseHandType := handType
	if strings.Contains(handType, "Three of a Kind") {
		baseHandType = "Three of a Kind"
	} else if strings.Contains(handType, "Four of a Kind") {
		baseHandType = "Four of a Kind"
	} else if strings.Contains(handType, "Full House") {
		baseHandType = "Full House"
	} else if strings.Contains(handType, "Two Pair") {
		baseHandType = "Two Pair"
	} else if strings.Contains(handType, "Pair") {
		baseHandType = "Pair"
	} else if strings.Contains(handType, "High Card") {
		baseHandType = "High Card"
	}

	// Get the hand strength
	strength := handStrengths[baseHandType]
	if strength == 0 {
		// Default to lowest strength if not found
		strength = 1.0
	}

	// Add further strength based on specific card ranks
	// This is a simplified approach; a real poker engine would be more detailed
	if strings.Contains(handType, "Ace") {
		strength += 0.14
	} else if strings.Contains(handType, "King") {
		strength += 0.13
	} else if strings.Contains(handType, "Queen") {
		strength += 0.12
	} else if strings.Contains(handType, "Jack") {
		strength += 0.11
	} else if strings.Contains(handType, "Ten") {
		strength += 0.10
	}

	return strength
}

// findWinners identifies the indices of players with the highest hand strength
func findWinners(handStrengths []float64) []int {
	winners := []int{}
	highestStrength := -1.0

	// Find the highest hand strength
	for _, strength := range handStrengths {
		if strength > highestStrength {
			highestStrength = strength
		}
	}

	// Find all players with the highest strength
	for idx, strength := range handStrengths {
		if strength == highestStrength {
			winners = append(winners, idx)
		}
	}

	return winners
}
