package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)

const KuteGoAPIURL = "https://kutego-api-xxxxx-ew.a.run.app"

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

type Gopher struct {
	Name string `json: "name"`
}

const prefix string = ","

func pre(val string) string {
	return prefix + val
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == pre("test") {
		s.ChannelMessageSend(m.ChannelID, "Test!")
	}

	re := regexp.MustCompile(`^` + prefix + `r(?:oll)? (.+)`)
	if re.MatchString(m.Content) {
		matches := re.FindStringSubmatch(m.Content)
		s.ChannelMessageSend(m.ChannelID, Parse(matches[1]))
	}

	if m.Content == pre("imagetest") {

		//Call the KuteGo API and retrieve our cute Dr Who Gopher
		response, err := http.Get(`https://images.vexels.com/media/users/3/150755/raw/a6f9b6779f3f3f6177e638b5af256091-playing-cards-set.jpg`)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()

		if response.StatusCode == 200 {
			_, err = s.ChannelFileSend(m.ChannelID, "card.jpg", response.Body)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Error: Can't get dr-who Gopher! :-(")
		}
	}
}
