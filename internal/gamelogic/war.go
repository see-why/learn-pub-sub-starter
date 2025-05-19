package gamelogic

import (
	"fmt"
)

type WarOutcome int

const (
	WarOutcomeNotInvolved WarOutcome = iota
	WarOutcomeNoUnits
	WarOutcomeYouWon
	WarOutcomeOpponentWon
	WarOutcomeDraw
)

// UnitPower represents the power level of different unit types
var UnitPower = map[UnitRank]int{
	RankArtillery: 10,
	RankCavalry:   5,
	RankInfantry:  1,
}

func (gs *GameState) HandleWar(rw RecognitionOfWar) (outcome WarOutcome, winner string, loser string) {
	defer fmt.Println("------------------------")
	fmt.Println()
	fmt.Println("==== War Declared ====")
	fmt.Printf("%s has declared war on %s!\n", rw.Attacker.Username, rw.Defender.Username)

	player := gs.GetPlayerSnap()

	// Validate war participants
	if !isValidWarParticipant(player, rw) {
		return WarOutcomeNotInvolved, "", ""
	}

	// Find overlapping location
	overlappingLocation := getOverlappingLocation(rw.Attacker, rw.Defender)
	if overlappingLocation == "" {
		fmt.Printf("Error! No units are in the same location. No war will be fought.\n")
		return WarOutcomeNoUnits, "", ""
	}

	// Get units at the overlapping location
	attackerUnits := getUnitsAtLocation(rw.Attacker.Units, Location(overlappingLocation))
	defenderUnits := getUnitsAtLocation(rw.Defender.Units, Location(overlappingLocation))

	// Display unit information
	displayWarUnits(rw.Attacker.Username, attackerUnits)
	displayWarUnits(rw.Defender.Username, defenderUnits)

	// Calculate power levels
	attackerPower := calculatePowerLevel(attackerUnits)
	defenderPower := calculatePowerLevel(defenderUnits)
	fmt.Printf("Attacker has a power level of %v\n", attackerPower)
	fmt.Printf("Defender has a power level of %v\n", defenderPower)

	// Determine war outcome
	outcome, winner, loser = determineWarOutcome(attackerPower, defenderPower, rw, player.Username)

	// Handle unit removal based on outcome
	if outcome == WarOutcomeDraw || (outcome == WarOutcomeOpponentWon && player.Username == rw.Attacker.Username) {
		gs.removeUnitsInLocation(overlappingLocation)
		fmt.Printf("Your units in %s have been killed.\n", overlappingLocation)
	}

	return outcome, winner, loser
}

func isValidWarParticipant(player Player, rw RecognitionOfWar) bool {
	if player.Username == rw.Defender.Username {
		fmt.Printf("%s, you published the war.\n", player.Username)
		return false
	}

	if player.Username != rw.Attacker.Username {
		fmt.Printf("%s, you are not involved in this war.\n", player.Username)
		return false
	}

	return true
}

func getUnitsAtLocation(units map[int]Unit, location Location) []Unit {
	var locationUnits []Unit
	for _, unit := range units {
		if unit.Location == location {
			locationUnits = append(locationUnits, unit)
		}
	}
	return locationUnits
}

func displayWarUnits(username string, units []Unit) {
	fmt.Printf("%s's units:\n", username)
	for _, unit := range units {
		fmt.Printf("  * %v\n", unit.Rank)
	}
}

func calculatePowerLevel(units []Unit) int {
	power := 0
	for _, unit := range units {
		power += UnitPower[unit.Rank]
	}
	return power
}

func determineWarOutcome(attackerPower, defenderPower int, rw RecognitionOfWar, currentPlayer string) (WarOutcome, string, string) {
	if attackerPower > defenderPower {
		fmt.Printf("%s has won the war!\n", rw.Attacker.Username)
		if currentPlayer == rw.Defender.Username {
			fmt.Println("You have lost the war!")
			return WarOutcomeOpponentWon, rw.Attacker.Username, rw.Defender.Username
		}
		return WarOutcomeYouWon, rw.Attacker.Username, rw.Defender.Username
	} else if defenderPower > attackerPower {
		fmt.Printf("%s has won the war!\n", rw.Defender.Username)
		if currentPlayer == rw.Attacker.Username {
			fmt.Println("You have lost the war!")
			return WarOutcomeOpponentWon, rw.Defender.Username, rw.Attacker.Username
		}
		return WarOutcomeYouWon, rw.Defender.Username, rw.Attacker.Username
	}

	fmt.Println("The war ended in a draw!")
	return WarOutcomeDraw, rw.Attacker.Username, rw.Defender.Username
}
