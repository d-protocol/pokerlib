package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/d-protocol/pokerlib"
)

// handEvaluation contains all the data needed to evaluate a hand
type handEvaluation struct {
	handType    string
	strength    float64
	tiebreakers []float64
	holeCards   []string
}

func main() {
	fmt.Println("\n=== Poker Game Simulation with Enhanced Shuffling ===")
	fmt.Println("This demo shows the improved card randomization in action")

	// Number of players in the game
	playerCount := 9
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &playerCount)
		if playerCount < 2 {
			playerCount = 2
		} else if playerCount > 9 {
			playerCount = 9
		}
	}

	fmt.Printf("Starting a poker game with %d players\n", playerCount)
	fmt.Printf("Playing %d hands, rotating the dealer position each hand\n\n", playerCount)

	// Create a standard deck and show the original order
	originalDeck := pokerlib.NewStandardDeckCards()
	fmt.Println("--- Original Deck Order ---")
	printDeck(originalDeck)

	// Process each hand with the dealer button rotation
	for handNum := 0; handNum < playerCount; handNum++ {
		fmt.Printf("\n======= HAND #%d =======\n", handNum+1)

		// Create standard game options
		gameOptions := pokerlib.NewStardardGameOptions()

		// Setup players
		playerSettings := make([]*pokerlib.PlayerSetting, playerCount)
		for i := 0; i < playerCount; i++ {
			playerSettings[i] = &pokerlib.PlayerSetting{
				PlayerID:  fmt.Sprintf("Player %d", i+1),
				Bankroll:  1000, // Each player starts with 1000 chips
				Positions: []string{},
			}
		}

		// Set dealer, small blind and big blind positions, rotating each hand
		dealerPos := handNum % playerCount
		sbPos := (dealerPos + 1) % playerCount
		bbPos := (dealerPos + 2) % playerCount

		playerSettings[dealerPos].Positions = append(playerSettings[dealerPos].Positions, "dealer")
		playerSettings[sbPos].Positions = append(playerSettings[sbPos].Positions, "sb")
		playerSettings[bbPos].Positions = append(playerSettings[bbPos].Positions, "bb")

		// Update game options
		gameOptions.Players = playerSettings
		gameOptions.Blind = pokerlib.BlindSetting{
			SB: 5,  // Small blind 5 chips
			BB: 10, // Big blind 10 chips
		}
		gameOptions.Ante = 0 // No ante

		// Create a new shuffled deck for this hand
		shuffledDeck := pokerlib.ShuffleCards(originalDeck)
		if handNum == 0 {
			// Only show the shuffled deck on the first hand to avoid too much output
			fmt.Println("\n--- Shuffled Deck ---")
			printDeck(shuffledDeck)
		}

		// Use the shuffled deck for the game
		gameOptions.Deck = shuffledDeck
		gameOptions.HoleCardsCount = 2

		// Create game
		game := pokerlib.NewGame(gameOptions)
		if err := game.Start(); err != nil {
			log.Fatalf("Failed to start game: %v", err)
		}

		fmt.Println("\n--- Game Started ---")
		fmt.Printf("Dealer: Player %d\n", game.Dealer().SeatIndex()+1)
		fmt.Printf("Small Blind: Player %d\n", game.SmallBlind().SeatIndex()+1)
		fmt.Printf("Big Blind: Player %d\n", game.BigBlind().SeatIndex()+1)

		// Initialize game
		if err := game.EmitEvent(pokerlib.GameEvent_Started); err != nil {
			log.Fatalf("Failed to emit start event: %v", err)
		}

		// Pay blinds
		if err := game.EmitEvent(pokerlib.GameEvent_BlindsRequested); err != nil {
			log.Fatalf("Failed to request blinds: %v", err)
		}

		if err := game.PayBlinds(); err != nil {
			log.Fatalf("Failed to pay blinds: %v", err)
		}

		// Deal hole cards
		fmt.Println("\n--- Hole Cards ---")
		deckPosition := 0
		for i := 0; i < game.GetPlayerCount(); i++ {
			player := game.Player(i)
			playerState := player.State()

			// Deal 2 cards to this player
			playerCards := []string{shuffledDeck[deckPosition], shuffledDeck[deckPosition+1]}
			deckPosition += 2

			// Assign the hole cards
			playerState.HoleCards = playerCards
			fmt.Printf("Player %d: %v\n", i+1, playerCards)
		}

		// Initialize game fully
		if err := game.EmitEvent(pokerlib.GameEvent_Initialized); err != nil {
			log.Fatalf("Failed to initialize game: %v", err)
		}

		// Deal community cards (simplified for demo purposes)
		fmt.Println("\n--- Community Cards ---")
		// Flop (3 cards)
		flop := shuffledDeck[deckPosition : deckPosition+3]
		deckPosition += 3
		fmt.Printf("Flop: %v\n", flop)

		// Turn (1 card)
		turn := shuffledDeck[deckPosition : deckPosition+1]
		deckPosition += 1
		fmt.Printf("Turn: %v\n", turn)

		// River (1 card)
		river := shuffledDeck[deckPosition : deckPosition+1]
		deckPosition += 1
		fmt.Printf("River: %v\n", river)

		// Combine all community cards
		communityCards := append(append(flop, turn...), river...)
		game.GetState().Status.Board = communityCards

		// Perform hand evaluation for all players
		fmt.Println("\n--- Final Hands ---")
		handTypes := make(map[int]string)
		handStrengths := make(map[int]float64)
		handDetails := make(map[int]handEvaluation)

		for i := 0; i < game.GetPlayerCount(); i++ {
			player := game.Player(i)
			playerState := player.State()

			// Get hole cards
			holeCards := playerState.HoleCards

			// Evaluate the hand
			handType, usingCommunity := evaluateHandWithSource(holeCards, communityCards)
			handTypes[i] = handType

			// Store detailed hand evaluation for winner determination
			strength, tiebreakers := calculateHandStrength(holeCards, communityCards)
			handStrengths[i] = strength
			handDetails[i] = handEvaluation{
				handType:    handType,
				strength:    strength,
				tiebreakers: tiebreakers,
				holeCards:   holeCards,
			}

			if usingCommunity {
				fmt.Printf("Player %d: %v - %s (using community cards)\n", i+1, holeCards, handType)
			} else {
				fmt.Printf("Player %d: %v - %s\n", i+1, holeCards, handType)
			}
		}

		// Check for duplicate hand types
		fmt.Println("\n--- Hand Type Analysis ---")
		checkDuplicateHandTypes(handTypes)

		// Determine the winner(s)
		winners := determineWinners(handDetails)
		fmt.Println("\n--- Winner Determination ---")
		if len(winners) == 1 {
			winnerIdx := winners[0]
			fmt.Printf("Winner: Player %d with %s\n", winnerIdx+1, handTypes[winnerIdx])
		} else {
			fmt.Println("Tie between multiple players:")
			for _, winnerIdx := range winners {
				fmt.Printf("- Player %d with %s\n", winnerIdx+1, handTypes[winnerIdx])
			}
		}
	}

	fmt.Println("\nSimulation completed successfully!")
}

// Helper functions

// printDeck displays cards in a grid format
func printDeck(cards []string) {
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

// evaluateHandWithSource determines the best poker hand and indicates if community cards are used
func evaluateHandWithSource(holeCards, communityCards []string) (string, bool) {
	// Combine hole cards and community cards
	allCards := append([]string{}, holeCards...)
	allCards = append(allCards, communityCards...)

	// Determine if the hand uses community cards
	usingCommunity := false

	// Count suits for flush detection
	suits := make(map[string]int)
	holeSuits := make(map[string]int)
	for _, card := range allCards {
		if len(card) > 0 {
			suit := string(card[0])
			suits[suit]++
		}
	}
	for _, card := range holeCards {
		if len(card) > 0 {
			suit := string(card[0])
			holeSuits[suit]++
		}
	}

	// Check for flush
	hasFlush := false
	for suit, count := range suits {
		if count >= 5 {
			hasFlush = true
			// Check if flush uses community cards
			if holeSuits[suit] < 2 {
				usingCommunity = true
			}
			break
		}
	}

	// Count ranks for pairs, etc.
	ranks := make(map[string]int)
	holeRanks := make(map[string]int)
	for _, card := range allCards {
		if len(card) >= 2 {
			rank := string(card[1])
			ranks[rank]++
		}
	}
	for _, card := range holeCards {
		if len(card) >= 2 {
			rank := string(card[1])
			holeRanks[rank]++
		}
	}

	// Check for different hand types
	fourOfAKind := ""
	threeOfAKindRank := ""
	pairRanks := []string{}

	for rank, count := range ranks {
		if count == 4 {
			fourOfAKind = rank
			// Check if four of a kind uses community cards
			if holeRanks[rank] < 2 {
				usingCommunity = true
			}
		} else if count == 3 {
			threeOfAKindRank = rank
			// Check if three of a kind uses community cards
			if holeRanks[rank] < 2 {
				usingCommunity = true
			}
		} else if count == 2 {
			pairRanks = append(pairRanks, rank)
			// Check if pair uses community cards
			if holeRanks[rank] < 2 {
				usingCommunity = true
			}
		}
	}

	// Determine hand type
	if hasFlush {
		return "Flush", usingCommunity
	}

	if fourOfAKind != "" {
		return fmt.Sprintf("Four of a Kind - %ss", cardName(fourOfAKind)), usingCommunity
	}

	if threeOfAKindRank != "" && len(pairRanks) > 0 {
		return fmt.Sprintf("Full House - %ss over %ss",
			cardName(threeOfAKindRank), cardName(pairRanks[0])), usingCommunity
	}

	if threeOfAKindRank != "" {
		return fmt.Sprintf("Three of a Kind - %ss", cardName(threeOfAKindRank)), usingCommunity
	}

	if len(pairRanks) >= 2 {
		return fmt.Sprintf("Two Pair - %ss and %ss",
			cardName(pairRanks[0]), cardName(pairRanks[1])), usingCommunity
	}

	if len(pairRanks) == 1 {
		return fmt.Sprintf("Pair of %ss", cardName(pairRanks[0])), usingCommunity
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

	// If high card is not in hole cards, we're using community cards
	if holeRanks[highest] == 0 {
		usingCommunity = true
	}

	return fmt.Sprintf("High Card %s", cardName(highest)), usingCommunity
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

// checkDuplicateHandTypes reports on duplicate hand types
func checkDuplicateHandTypes(handTypes map[int]string) {
	// Create a frequency map
	handFrequency := make(map[string]int)
	for _, handType := range handTypes {
		handFrequency[handType]++
	}

	// Count duplicates and report
	duplicateCount := 0
	duplicateHands := make(map[string][]int)

	for playerIdx, handType := range handTypes {
		if handFrequency[handType] > 1 {
			// This is a duplicate hand
			list, ok := duplicateHands[handType]
			if !ok {
				list = make([]int, 0)
			}
			duplicateHands[handType] = append(list, playerIdx+1) // +1 for display
		}
	}

	// Count total duplicate pairs
	for _, count := range handFrequency {
		if count > 1 {
			// Calculate number of pairs
			duplicateCount += (count * (count - 1)) / 2
		}
	}

	// Report results
	playerCount := len(handTypes)
	totalPairs := (playerCount * (playerCount - 1)) / 2
	duplicatePercentage := float64(duplicateCount) * 100.0 / float64(totalPairs)

	fmt.Printf("Players: %d, Possible pairs: %d\n", playerCount, totalPairs)
	fmt.Printf("Duplicate hand types: %d (%.2f%%)\n", duplicateCount, duplicatePercentage)

	if duplicateCount > 0 {
		fmt.Println("Players with the same hand types:")
		for handType, players := range duplicateHands {
			fmt.Printf("  %s: Players %v\n", handType, players)
		}
	}
}

// calculateHandStrength returns a numerical value for hand strength and tiebreaker info
func calculateHandStrength(holeCards, communityCards []string) (float64, []float64) {
	// Combine hole cards and community cards
	allCards := append([]string{}, holeCards...)
	allCards = append(allCards, communityCards...)

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

	// Get the hand type
	handType, _ := evaluateHandWithSource(holeCards, communityCards)

	// Extract base hand type without specifics
	baseHandType := handType
	if strings.Contains(handType, " of ") {
		baseHandType = handType[:strings.Index(handType, " of ")]
	} else if strings.Contains(handType, "Pair of") {
		baseHandType = "Pair"
	} else if strings.Contains(handType, "Two Pair") {
		baseHandType = "Two Pair"
	} else if strings.Contains(handType, "High Card") {
		baseHandType = "High Card"
	}

	// Get the hand strength
	strength := handStrengths[baseHandType]
	if strength == 0 {
		// Default to lowest strength if not found
		strength = 1.0
	}

	// Parse tiebreakers from the hand type
	tiebreakers := extractTiebreakers(handType)

	return strength, tiebreakers
}

// extractTiebreakers extracts numeric tiebreaker values from hand description
func extractTiebreakers(handType string) []float64 {
	tiebreakers := []float64{}

	// Get rank values for tiebreakers
	rankValues := map[string]float64{
		"Two":   2.0,
		"Three": 3.0,
		"Four":  4.0,
		"Five":  5.0,
		"Six":   6.0,
		"Seven": 7.0,
		"Eight": 8.0,
		"Nine":  9.0,
		"Ten":   10.0,
		"Jack":  11.0,
		"Queen": 12.0,
		"King":  13.0,
		"Ace":   14.0,
	}

	// Extract rank information for different hand types
	if strings.Contains(handType, "Four of a Kind") {
		rankText := handType[strings.Index(handType, "-")+2 : len(handType)-1] // remove last 's'
		tiebreakers = append(tiebreakers, rankValues[rankText])
	} else if strings.Contains(handType, "Full House") {
		parts := strings.Split(handType, " over ")
		rankText1 := parts[0][strings.Index(parts[0], "-")+2 : len(parts[0])-1] // remove last 's'
		rankText2 := parts[1][:len(parts[1])-1]                                 // remove last 's'
		tiebreakers = append(tiebreakers, rankValues[rankText1], rankValues[rankText2])
	} else if strings.Contains(handType, "Three of a Kind") {
		rankText := handType[strings.Index(handType, "-")+2 : len(handType)-1] // remove last 's'
		tiebreakers = append(tiebreakers, rankValues[rankText])
	} else if strings.Contains(handType, "Two Pair") {
		parts := strings.Split(handType, " and ")
		rankText1 := parts[0][strings.Index(parts[0], "-")+2 : len(parts[0])-1] // remove last 's'
		rankText2 := parts[1][:len(parts[1])-1]                                 // remove last 's'
		tiebreakers = append(tiebreakers, rankValues[rankText1], rankValues[rankText2])
	} else if strings.Contains(handType, "Pair of") {
		rankText := handType[strings.Index(handType, "of ")+3 : len(handType)-1] // remove last 's'
		tiebreakers = append(tiebreakers, rankValues[rankText])
	} else if strings.Contains(handType, "High Card") {
		rankText := handType[strings.Index(handType, "Card ")+5:]
		tiebreakers = append(tiebreakers, rankValues[rankText])
	}

	return tiebreakers
}

// determineWinners returns indices of players with winning hands
func determineWinners(handDetails map[int]handEvaluation) []int {
	winners := []int{}
	maxStrength := -1.0

	// First find the highest hand strength
	for _, details := range handDetails {
		if details.strength > maxStrength {
			maxStrength = details.strength
		}
	}

	// Find all players with the max hand strength
	potentialWinners := []int{}
	for playerIdx, details := range handDetails {
		if details.strength == maxStrength {
			potentialWinners = append(potentialWinners, playerIdx)
		}
	}

	// If only one player has the max hand strength, they're the winner
	if len(potentialWinners) == 1 {
		return potentialWinners
	}

	// Need to resolve ties using tiebreakers
	winners = resolveHandTies(potentialWinners, handDetails)
	return winners
}

// resolveHandTies resolves ties based on tiebreakers
func resolveHandTies(candidates []int, handDetails map[int]handEvaluation) []int {
	if len(candidates) == 0 {
		return []int{}
	}

	// Compare tiebreakers in order
	for tiebreakerIdx := 0; tiebreakerIdx < 5; tiebreakerIdx++ {
		maxValue := -1.0

		// Find the highest value for this tiebreaker
		for _, playerIdx := range candidates {
			details := handDetails[playerIdx]
			if tiebreakerIdx < len(details.tiebreakers) && details.tiebreakers[tiebreakerIdx] > maxValue {
				maxValue = details.tiebreakers[tiebreakerIdx]
			}
		}

		// If this tiebreaker isn't available or all remaining players tie, continue
		if maxValue == -1.0 {
			continue
		}

		// Keep only players with the max value for this tiebreaker
		newCandidates := []int{}
		for _, playerIdx := range candidates {
			details := handDetails[playerIdx]
			if tiebreakerIdx < len(details.tiebreakers) && details.tiebreakers[tiebreakerIdx] == maxValue {
				newCandidates = append(newCandidates, playerIdx)
			}
		}

		// If we've narrowed down the candidates, either return them or continue with remaining
		if len(newCandidates) < len(candidates) {
			if len(newCandidates) == 1 {
				return newCandidates // Single winner
			}
			candidates = newCandidates // Continue with remaining candidates
		}
	}

	// If we get here, the remaining candidates are all tied
	return candidates
}
