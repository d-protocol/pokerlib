package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/d-protocol/pokerlib"
)

func main() {
	fmt.Println("=== Poker Shuffling Algorithm Test ===")
	fmt.Println()

	// Run a single game to showcase the improved shuffling
	runSampleGame()

	// Run simulation to compare hand distributions
	if len(os.Args) > 1 && os.Args[1] == "--analyze" {
		fmt.Println("\n=== Running Statistical Analysis ===")
		runAnalysis(1000)
	} else {
		fmt.Println("\nRun with '--analyze' flag for statistical analysis.")
	}
}

// runSampleGame demonstrates a single poker game with shuffled cards
func runSampleGame() {
	// Create and shuffle a standard deck
	deck := pokerlib.NewStandardDeckCards()
	fmt.Println("Original deck (unshuffled):")
	printCardGrid(deck)

	// Apply the improved shuffling algorithm
	fmt.Println("\nShuffled deck:")
	shuffled := pokerlib.ShuffleCards(deck)
	printCardGrid(shuffled)

	// Deal cards to 9 players (as in a full poker table)
	fmt.Println("\n=== Dealing Cards to Players ===")
	playerHands := make([][]string, 9)
	for i := 0; i < 9; i++ {
		// Each player gets 2 cards
		playerHands[i] = []string{
			shuffled[i*2],
			shuffled[i*2+1],
		}
		fmt.Printf("Player %d: %v\n", i+1, playerHands[i])
	}

	// Deal community cards
	fmt.Println("\n=== Community Cards ===")
	communityCards := []string{
		shuffled[18], // burn a card then deal flop
		shuffled[19],
		shuffled[20],
		shuffled[21], // burn a card then deal turn
		shuffled[22], // burn a card then deal river
	}
	fmt.Printf("Flop: %v %v %v\n", communityCards[0], communityCards[1], communityCards[2])
	fmt.Printf("Turn: %v\n", communityCards[3])
	fmt.Printf("River: %v\n", communityCards[4])

	// Evaluate each player's hand
	fmt.Println("\n=== Hand Evaluations ===")
	for i := 0; i < 9; i++ {
		handType := evaluateHand(playerHands[i], communityCards)
		fmt.Printf("Player %d: %v - %s\n", i+1, playerHands[i], handType)
	}

	// Check for duplicated hand types
	fmt.Println("\n=== Hand Type Analysis ===")
	checkDuplicateHands(playerHands, communityCards)
}

// runAnalysis performs a statistical analysis of the shuffling algorithm
func runAnalysis(games int) {
	start := time.Now()
	fmt.Printf("Running %d simulated games...\n", games)

	// Track statistics
	totalGames := 0
	totalPlayers := 0
	totalDuplicateHands := 0
	maxDuplicatesInGame := 0

	// Track patterns
	consecutiveSuitCount := 0
	consecutiveRankCount := 0
	totalCards := 0

	for i := 0; i < games; i++ {
		// Create and shuffle a deck
		deck := pokerlib.NewStandardDeckCards()
		shuffled := pokerlib.ShuffleCards(deck)

		// Check for consecutive cards with same suit or rank
		for j := 0; j < len(shuffled)-1; j++ {
			if shuffled[j][0] == shuffled[j+1][0] {
				consecutiveSuitCount++
			}
			if len(shuffled[j]) >= 2 && len(shuffled[j+1]) >= 2 && shuffled[j][1] == shuffled[j+1][1] {
				consecutiveRankCount++
			}
			totalCards++
		}

		// Deal cards to players
		playerCount := 9
		playerHands := make([][]string, playerCount)
		for p := 0; p < playerCount; p++ {
			playerHands[p] = []string{
				shuffled[p*2],
				shuffled[p*2+1],
			}
		}

		// Deal community cards
		communityCards := []string{
			shuffled[18],
			shuffled[19],
			shuffled[20],
			shuffled[21],
			shuffled[22],
		}

		// Count duplicate hands in this game
		duplicates := countDuplicateHands(playerHands, communityCards)
		totalDuplicateHands += duplicates
		if duplicates > maxDuplicatesInGame {
			maxDuplicatesInGame = duplicates
		}

		totalGames++
		totalPlayers += playerCount
	}

	// Calculate statistics
	duration := time.Since(start)
	playerPairs := totalPlayers * (totalPlayers - 1) / 2 * totalGames
	duplicatePercentage := float64(totalDuplicateHands) * 100.0 / float64(playerPairs)
	suitPatternPercentage := float64(consecutiveSuitCount) * 100.0 / float64(totalCards)
	rankPatternPercentage := float64(consecutiveRankCount) * 100.0 / float64(totalCards)

	// Report results
	fmt.Printf("\nAnalysis completed in %v\n", duration)
	fmt.Printf("Games simulated: %d\n", totalGames)
	fmt.Printf("Total player pairs compared: %d\n", playerPairs)
	fmt.Printf("Total duplicate hands: %d\n", totalDuplicateHands)
	fmt.Printf("Duplicate hand percentage: %.2f%%\n", duplicatePercentage)
	fmt.Printf("Maximum duplicates in a single game: %d\n", maxDuplicatesInGame)
	fmt.Printf("Consecutive same suit percentage: %.2f%%\n", suitPatternPercentage)
	fmt.Printf("Consecutive same rank percentage: %.2f%%\n", rankPatternPercentage)

	// Interpret results
	fmt.Println("\nInterpretation:")
	if duplicatePercentage < 20.0 {
		fmt.Println("✅ Excellent shuffling randomness (duplicate hands < 20%)")
	} else if duplicatePercentage < 25.0 {
		fmt.Println("✓ Good shuffling randomness (duplicate hands < 25%)")
	} else if duplicatePercentage < 30.0 {
		fmt.Println("⚠️ Fair shuffling randomness (duplicate hands < 30%)")
	} else {
		fmt.Println("❌ Poor shuffling randomness (duplicate hands >= 30%)")
	}

	fmt.Printf("\nPattern analysis: %.2f%% same suit, %.2f%% same rank in consecutive positions\n",
		suitPatternPercentage, rankPatternPercentage)
}

// printCardGrid displays cards in a grid format for better visualization
func printCardGrid(cards []string) {
	for i, card := range cards {
		fmt.Printf("%s ", card)
		if (i+1)%13 == 0 {
			fmt.Println()
		}
	}
	if len(cards)%13 != 0 {
		fmt.Println()
	}
}

// evaluateHand determines the poker hand type
func evaluateHand(holeCards, communityCards []string) string {
	// Combine hole cards and community cards
	allCards := append([]string{}, holeCards...)
	allCards = append(allCards, communityCards...)

	// Count suits for flush detection
	suits := make(map[string]int)
	for _, card := range allCards {
		if len(card) > 0 {
			suit := string(card[0])
			suits[suit]++
		}
	}

	// Check for flush
	hasFlush := false
	for _, count := range suits {
		if count >= 5 {
			hasFlush = true
			break
		}
	}

	// Count ranks for pairs, etc.
	ranks := make(map[string]int)
	for _, card := range allCards {
		if len(card) >= 2 {
			rank := string(card[1])
			ranks[rank]++
		}
	}

	// Check for different hand types
	fourOfAKind := ""
	threeOfAKindRank := ""
	pairRanks := []string{}

	for rank, count := range ranks {
		if count == 4 {
			fourOfAKind = rank
		} else if count == 3 {
			threeOfAKindRank = rank
		} else if count == 2 {
			pairRanks = append(pairRanks, rank)
		}
	}

	// Determine hand type
	if hasFlush {
		return "Flush"
	}

	if fourOfAKind != "" {
		return fmt.Sprintf("Four of a Kind - %ss", cardName(fourOfAKind))
	}

	if threeOfAKindRank != "" && len(pairRanks) > 0 {
		return fmt.Sprintf("Full House - %ss over %ss",
			cardName(threeOfAKindRank), cardName(pairRanks[0]))
	}

	if threeOfAKindRank != "" {
		return fmt.Sprintf("Three of a Kind - %ss", cardName(threeOfAKindRank))
	}

	if len(pairRanks) >= 2 {
		return fmt.Sprintf("Two Pair - %ss and %ss",
			cardName(pairRanks[0]), cardName(pairRanks[1]))
	}

	if len(pairRanks) == 1 {
		return fmt.Sprintf("Pair of %ss", cardName(pairRanks[0]))
	}

	// Find highest card
	highest := ""
	highestValue := -1
	rankValues := map[string]int{
		"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8,
		"9": 9, "T": 10, "J": 11, "Q": 12, "K": 13, "A": 14,
	}

	for rank := range ranks {
		if rankValues[rank] > highestValue {
			highest = rank
			highestValue = rankValues[rank]
		}
	}

	return fmt.Sprintf("High Card %s", cardName(highest))
}

// cardName returns the full name of a card rank
func cardName(rank string) string {
	switch rank {
	case "A":
		return "Ace"
	case "K":
		return "King"
	case "Q":
		return "Queen"
	case "J":
		return "Jack"
	case "T":
		return "Ten"
	case "9":
		return "Nine"
	case "8":
		return "Eight"
	case "7":
		return "Seven"
	case "6":
		return "Six"
	case "5":
		return "Five"
	case "4":
		return "Four"
	case "3":
		return "Three"
	case "2":
		return "Two"
	default:
		return strings.ToUpper(rank)
	}
}

// checkDuplicateHands analyzes and reports on duplicate hand types
func checkDuplicateHands(playerHands [][]string, communityCards []string) {
	playerCount := len(playerHands)
	handTypes := make([]string, playerCount)

	// Evaluate each player's hand
	for i := 0; i < playerCount; i++ {
		handTypes[i] = evaluateHand(playerHands[i], communityCards)
	}

	// Check for duplicates
	duplicateCount := 0
	duplicatePairs := []string{}

	for i := 0; i < playerCount; i++ {
		for j := i + 1; j < playerCount; j++ {
			if handTypes[i] == handTypes[j] {
				duplicateCount++
				pair := fmt.Sprintf("Players %d and %d both have: %s",
					i+1, j+1, handTypes[i])
				duplicatePairs = append(duplicatePairs, pair)
			}
		}
	}

	// Report results
	totalPairs := playerCount * (playerCount - 1) / 2
	duplicatePercentage := float64(duplicateCount) * 100.0 / float64(totalPairs)

	fmt.Printf("Players: %d, Possible pairs: %d\n", playerCount, totalPairs)
	fmt.Printf("Duplicate hand types: %d (%.2f%%)\n",
		duplicateCount, duplicatePercentage)

	if duplicateCount > 0 {
		fmt.Println("Duplicate pairs:")
		for _, pair := range duplicatePairs {
			fmt.Println("  - " + pair)
		}
	}
}

// countDuplicateHands counts the number of duplicate hand types
func countDuplicateHands(playerHands [][]string, communityCards []string) int {
	playerCount := len(playerHands)
	handTypes := make([]string, playerCount)

	for i := 0; i < playerCount; i++ {
		handTypes[i] = evaluateHand(playerHands[i], communityCards)
	}

	duplicateCount := 0
	for i := 0; i < playerCount; i++ {
		for j := i + 1; j < playerCount; j++ {
			if handTypes[i] == handTypes[j] {
				duplicateCount++
			}
		}
	}

	return duplicateCount
}
