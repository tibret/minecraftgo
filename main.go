package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"minecraftgo/commands"
	"minecraftgo/secrets"
	"minecraftgo/twitch"
	"minecraftgo/wrapper"
	"net/http"
)

func main() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/startGame", startGame)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

func getRoot(res http.ResponseWriter, req *http.Request) {
	url := "https://id.twitch.tv/oauth2/authorize?client_id=" + secrets.ClientID + "&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2FstartGame&response_type=code&scope=channel%3Abot%20user%3Aread%3Achat%20user%3Abot"
	io.WriteString(res, "<html><body><a href=\""+url+"\">Click here to start game</a></body></html>")
}

func startGame(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Starting game...")
	params := req.URL.Query()

	code := params.Get("code")
	fmt.Println("Using Token", code)

	wpr := setupMinecraftServer()
	go setupWebsocket(code, wpr)

	io.WriteString(res, "<html><body><div>Server is starting</div></body></html>")
}

func setupMinecraftServer() *wrapper.Wrapper {
	cmd := wrapper.JavaExecCmd("server.jar", 1024, 1024)
	//cmd.Stdout = os.Stdout
	console := wrapper.NewConsole(cmd)
	wpr := wrapper.NewWrapper(console)

	return wpr
}

func setupWebsocket(code string, wpr *wrapper.Wrapper) {
	conn := twitch.NewConnection()
	defer conn.Cancel()

	wpr.Start()
	defer wpr.Stop()

	fmt.Println("!! Server loaded")

	authToken := twitch.Auth(code)
	fmt.Println("!! Got Auth Token", authToken)

	twitch.SubscribeToEvent(conn, "channel.chat.message", authToken)

	gameOver := false
	player_name := "tibretS"

	for !gameOver {
		_, data, err := conn.Conn.Read(conn.Context)
		if err != nil {
			panic(err)
		}

		// process metadata
		var metadata twitch.MessageMetadata
		err = json.Unmarshal(data, &metadata)
		if err != nil {
			break
		}

		if metadata.Metadata.MessageType == twitch.KeepAlive {
			continue
		} else if metadata.Metadata.MessageType == twitch.Notification {
			payload := twitch.GetMessageText(data)

			commands.Tell(wpr, player_name, payload)

			if payload == "skeleton" {
				commands.SummonMob(wpr, player_name, commands.Skeleton)
			} else if payload == "teleport" {
				commands.TeleportRandom(wpr, player_name, commands.NewVec3(50, 10, 50))
			} else if payload == "clearskies" {
				commands.SetWeather(wpr, commands.Clear)
			} else if payload == "rain" {
				commands.SetWeather(wpr, commands.Rain)
			} else if payload == "damage" {
				commands.Damage(wpr, player_name, 10)
			} else if payload == "gofast" {
				attribute_id := uuid.NewString()
				commands.Attribute(wpr, player_name, commands.MovementSpeed, attribute_id, 2)
			} else if payload == "slowdown" {
				attribute_id := uuid.NewString()
				commands.Attribute(wpr, player_name, commands.MovementSpeed, attribute_id, 0.5)
			} else if payload == "levelup" {
				commands.AddLevels(wpr, player_name, 10)
			} else if payload == "glow" {
				commands.SetEffect(wpr, player_name, commands.Glowing, 10, 1, false)
			} else if payload == "silktouch" {
				commands.Enchant(wpr, player_name, commands.SilkTouch, 1)
			} else if payload == "kill" {
				commands.Kill(wpr, player_name)
			} else if payload == "suitup" {
				commands.Give(wpr, player_name, []string{"minecraft:diamond_pickaxe",
					"minecraft:diamond_boots",
					"minecraft:diamond_helmet",
					"minecraft:diamond_shovel",
					"minecraft:diamond_axe",
					"minecraft:diamond_sword",
					"minecraft:diamond_chestplate",
					"minecraft:diamond_leggings"})
			}

			if payload == "quit" {
				gameOver = true
			}
		}
	}

	fmt.Println("Game ended, connection closed")
}
