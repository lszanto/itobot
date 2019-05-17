package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/nlopes/slack"
)

func main() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	token := getenv("SLACKTOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Print("Event recieved: ")
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				info := rtm.GetInfo()
				botID := "<@" + strings.ToLower(info.User.ID) + ">"
				messageText := strings.TrimSpace(strings.ToLower(ev.Text))

				if ev.User != info.User.ID && strings.HasPrefix(messageText, botID) && strings.Contains(messageText, "tomorrow") {
					user, err := api.GetUserInfo(ev.User)

					if err != nil {
						continue
					}

					userName := user.Profile.DisplayNameNormalized
					userLocation := strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "tomorrow", ""))

					addStatusTomorrow(userName, userLocation)

					rtm.SendMessage(rtm.NewOutgoingMessage("Thanks for letting us know <@"+ev.User+">", ev.Channel))
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				//
			}
		}
	}
}

func getenv(name string) string {
	v := os.Getenv(name)

	if v == "" {
		panic("missing required env variable " + name)
	}

	return v
}

func addStatusTomorrow(user string, location string) {

}
