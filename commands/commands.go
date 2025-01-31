package commands

import (
	"fmt"
	"math/rand"
	"minecraftgo/wrapper"
	"strconv"
	"strings"
)

func SummonMob(wpr *wrapper.Wrapper, player_name string, mob_name Mob) {
	//get the location of the player
	cmd := fmt.Sprintf("/data get entity %s Pos", player_name)
	fmt.Println("Running command --> ", cmd)
	res := wpr.SendCommand(cmd)
	fmt.Println("Response to command <--", res)

	//parse the information
	pos_start := strings.LastIndex(res, "[")
	pos_end := strings.LastIndex(res, "]")
	pos_string := strings.ReplaceAll(res[(pos_start+1):(pos_end-1)], "d", "") //clear out the double modifier
	pos := strings.Split(pos_string, ",")

	fmt.Println("Running command --> ", cmd)
	cmd = fmt.Sprintf("/summon %s %s%s%s", mob_name, pos[0], pos[1], pos[2])
	fmt.Println("Response to command <--", res)
	wpr.SendCommand(cmd)
}

func Tell(wpr *wrapper.Wrapper, player_name string, message string) {
	tell := fmt.Sprintf("/tell %s \"%s\"", player_name, message)
	wpr.SendCommand(tell)
}

func SetWeather(wpr *wrapper.Wrapper, weather Weather) {
	cmd := fmt.Sprintf("/weather %s", weather)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func Damage(wpr *wrapper.Wrapper, player_name string, amount int) {
	cmd := fmt.Sprintf("/damage %s %d minecraft:fireball by %s", player_name, amount, player_name)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func Attribute(wpr *wrapper.Wrapper, player_name string, attribute AttributeName, uuid string, name string, modifier float64) {
	cmd := fmt.Sprintf("/attribute %s %s modifier add %s %s %.2f", player_name, attribute, uuid, name, modifier)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func SetDifficulty(wpr *wrapper.Wrapper, diff Difficulty) {
	cmd := fmt.Sprintf("/difficulty %s", diff)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func SetEffect(wpr *wrapper.Wrapper, player_name string, effect Effect, seconds int, amplifier int, hideParticles bool) {
	cmd := fmt.Sprintf("/effect give %s %s %d %d %t", player_name, effect, seconds, amplifier, hideParticles)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func Enchant(wpr *wrapper.Wrapper, player_name string, enchantment Enchantment, level int) {
	cmd := fmt.Sprintf("/enchant %s minecraft:enchantment:%s %d", player_name, enchantment, level)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func AddLevels(wpr *wrapper.Wrapper, player_name string, amount int) {
	cmd := fmt.Sprintf("/experience add %s %d levels", player_name, amount)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func Kill(wpr *wrapper.Wrapper, player_name string) {
	cmd := fmt.Sprintf("/kill %s", player_name)
	fmt.Println("Running command --> ", cmd)
	wpr.SendCommand(cmd)
}

func TeleportRandom(wpr wrapper.Wrapper, player_name string, x_max float64, y_max float64, z_max float64) {
	//get the location of the player
	cmd := fmt.Sprintf("/data get entity %s Pos", player_name)
	fmt.Println("Running command --> ", cmd)
	res := wpr.SendCommand(cmd)
	fmt.Println("Response to command <--", res)

	//parse the information
	pos_start := strings.LastIndex(res, "[")
	pos_end := strings.LastIndex(res, "]")
	pos_string := strings.ReplaceAll(res[(pos_start+1):(pos_end-1)], "d", "") //clear out the double modifier
	pos_string = strings.ReplaceAll(pos_string, " ", "")                      //remove spaces
	pos := strings.Split(pos_string, ",")
	fpos := make([]float64, len(pos))
	for idx, p := range pos {
		fp, err := strconv.ParseFloat(p, 64)
		if err != nil {
			fmt.Println("problem parsing positions, aborting")
			return
		}

		fpos[idx] = fp
	}

	//add randomness
	dir := randDirection()
	fpos[0] = fpos[0] + (rand.Float64() * x_max * dir)
	dir = randDirection()
	fpos[1] = fpos[1] + (rand.Float64() * y_max * dir)
	dir = randDirection()
	fpos[2] = fpos[2] + (rand.Float64() * z_max * dir)

	fmt.Println("Running command --> ", cmd)
	cmd = fmt.Sprintf("/teleport %s %.6f %.6f %.6f", player_name, fpos[0], fpos[1], fpos[2])
	fmt.Println("Response to command <--", res)
	wpr.SendCommand(cmd)

}

// returns either 1 or -1, used for multiplying directions randomly
func randDirection() float64 {
	dir := rand.Intn(2)
	if dir == 0 {
		return 1
	} else {
		return -1
	}
}
