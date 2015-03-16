package main

import (
	"log"
	"regexp"
	"strings"
	"time"
)

func Post(msg string) {
	log.Printf(">>> %s", msg)

	if err := session.PostMessage(msg); err != nil {
		log.Fatalf("Could not post message to RobustIRC: %v", err)
	}
}

func Privmsg(user string, msg string) {
	Post("PRIVMSG " + user + " :" + msg)
}

func Topic(channel string, topic string) {
	Post("TOPIC " + channel + " :" + topic)
}

func TopicGet(channel string) string {
	received := make(chan string)

	ListenerAdd(func(parsed Message) bool {
		// Example Topic:
		// PREFIX=robustirc.net COMMAND=332 PARAMS=[frank #test]

		p := parsed.Params

		if len(p) < 2 || p[1] != channel {
			// not the channel we're interested in
			return true
		}

		if parsed.Command == RPL_TOPIC {
			received <- parsed.Trailing
			return false
		}

		if parsed.Command == RPL_NOTOPIC {
			received <- ""
			return false
		}

		return true
	})

	Post("TOPIC " + channel)

	select {
	case topic := <-received:
		return topic
	case <-time.After(10 * time.Second):
	}
	return ""
}

func IsPrivateQuery(p Message) bool {
	return p.Command == "PRIVMSG" && Target(p) == *nick
}

func Join(channel string) {
	channel = strings.TrimSpace(channel)
	channel = strings.TrimPrefix(channel, "#")

	if channel == "" {
		return
	}

	log.Printf("joining #%s", channel)
	if *nickserv_password != "" {
		Privmsg("chanserv", "invite #"+channel)
	}
	Post("JOIN #" + channel)
}

func Nick(p Message) string {
	return strings.SplitN(p.Prefix, "!", 2)[0]
}

func Target(parsed Message) string {
	p := parsed.Params
	if len(p) == 0 {
		return ""
	} else {
		return p[0]
	}
}

func IsNickAdmin(p Message) bool {
	nick := Nick(p)
	admins := regexp.MustCompile("\\s+").Split(*admins, -1)

	for _, admin := range admins {
		if *verbose {
			log.Printf("debug admin: checking if |%s|==|%s| (=%v)", nick, admin, nick == admin)
		}
		if nick == admin {
			return true
		}
	}
	return false
}

func IsIn(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}