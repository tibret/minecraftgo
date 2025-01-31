package main

import (
	"encoding/json"
	"fmt"
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
	setupWebsocket(code, wpr)
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

			commands.Tell(wpr, "tibretS", payload)

			if payload == "skeleton" {
				commands.SummonMob(wpr, "tibretS", commands.Skeleton)
			}

			if payload == "quit" {
				gameOver = true
			}
		}
	}

	fmt.Println("Game ended, connection closed")
}
