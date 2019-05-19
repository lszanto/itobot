package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/nlopes/slack"
)

func main() {
	dbFile := getenv("DBFILE")

	db, err := bolt.Open(dbFile, 0600, nil)
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

				user, err := api.GetUserInfo(ev.User)
				if err != nil {
					continue
				}

				outGoingMessage := ""

				if ev.User != info.User.ID && strings.HasPrefix(messageText, botID) && strings.Contains(messageText, "where") {
					bucket := time.Now().Format("02.01.2006")
					if strings.Contains(messageText, "tomorrow") {
						bucket = time.Now().AddDate(0, 0, 1).Format("02.01.2006")
					}

					outGoingMessage = status.GetLocationsFromBucket(db, bucket)
				} else if strings.Contains(messageText, "tomorrow") {
					status.AddStatusTomorrow(db, user.Profile.DisplayNameNormalized, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "tomorrow", "")))

					outGoingMessage = "Thanks for letting us know <@" + ev.User + ">"
				} else if strings.Contains(messageText, "today") {
					status.AddStatusToday(db, user.Profile.DisplayNameNormalized, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "today", "")))

					outGoingMessage = "Thanks for letting us know <@" + ev.User + ">"
				} else {
					continue
				}

				if outGoingMessage != "" {
					rtm.SendMessage(rtm.NewOutgoingMessage(outGoingMessage, ev.Channel))
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
