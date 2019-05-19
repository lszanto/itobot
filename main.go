package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/lszanto/itobot/status"
	"github.com/nlopes/slack"
)

type msg struct {
	body        string
	forBot      bool
	hasWhere    bool
	hasTomorrow bool
	hasToday    bool
}

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
				msg := parseMessagetext(messageText, botID)

				user, err := api.GetUserInfo(ev.User)
				if err != nil {
					continue
				}

				outGoingMessage := ""

				if msg.forBot && msg.hasWhere {
					bucket := time.Now().Format("02.01.2006")
					if msg.hasTomorrow {
						bucket = time.Now().AddDate(0, 0, 1).Format("02.01.2006")
					}

					outGoingMessage = status.GetLocationsFromBucket(db, bucket)
				} else if msg.hasTomorrow {
					status.AddStatusTomorrow(db, user.Profile.DisplayNameNormalized, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "tomorrow", "")))
				} else if msg.hasToday {
					status.AddStatusToday(db, user.Profile.DisplayNameNormalized, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "today", "")))
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

func parseMessagetext(messageText string, botID string) msg {
	return msg{
		body:        messageText,
		forBot:      strings.HasPrefix(messageText, botID),
		hasWhere:    strings.Contains(messageText, "where"),
		hasTomorrow: strings.Contains(messageText, "tomorrow"),
		hasToday:    strings.Contains(messageText, "today"),
	}
}
