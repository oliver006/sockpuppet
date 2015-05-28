package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
)

var (
	wsAddr = flag.String("ws_addr", "", "Address of the host to connect to")
)

type ServerMessage []struct {
	UUID        string `json:"uuid"`
	Timestamp   string `json:"timestamp"`
	Region      string `json:"region"`
	Zone        string `json:"zone"`
	Product     string `json:"product"`
	Project     string `json:"project"`
	Environment string `json:"environment"`
	Type        string `json:"type"`
	Body        string `json:"body,omitempty"`
}

type MessageBody struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Version      int    `json:"version"`
	SubType      string `json:"sub_type"`
	Label        string `json:"label"`
	StartTime    int    `json:"start_time"`
	EndTime      int    `json:"end_time"`
	LastModified int    `json:"last_modified"`

	Links []struct {
		URL         string `json:"url"`
		Count       int    `json:"count"`
		ContentID   string `json:"content_id"`
		ContentType string `json:"content_type"`
		Offset      int    `json:"offset"`
	} `json:"links"`
}

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()

	if *wsAddr == "" {
		fmt.Println("Need to provide a valid host via --ws_addr=\"\"")
		return
	}

	addr := *wsAddr
	switch {
	case strings.HasPrefix(addr, "ws://"):
		addr = strings.Replace(*wsAddr, ".com./", ".com.:80/", 1)
	case strings.HasPrefix(addr, "wss://"):
		addr = strings.Replace(*wsAddr, ".com./", ".com.:443/", 1)
	}

	ws, err := websocket.Dial(addr, "", "http://www.nytimes.com/")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to %s", addr)

	var msgBuf = make([]byte, 4096)
	for {
		bufLen, err := ws.Read(msgBuf)
		if err != nil {
			log.Printf("read err: %s", err)
			break
		}

		if bufLen < 1 {
			continue
		}

		switch msgBuf[0] {
		case 'o':
			// reply to the login request
			cookie := randCookie()
			msg := fmt.Sprintf(`["{\"action\":\"login\", \"client_app\":\"hermes.push\", \"cookies\":{\"nyt-s\":\"%s\"}}"]`, cookie)
			_, err := ws.Write([]byte(msg))
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Sent cookie: %s\n", cookie)

		case 'h':
			// keep-alive
			log.Println("ping")

		case 'a':
			// some JSON encoded data, let's decode it!
			msg, err := decodeServerMessage(msgBuf[1:bufLen])
			if err != nil {
				log.Printf("no good: %s", err)
				continue
			}

			// response to the login message?
			if msg[0].Product == "core" && msg[0].Project == "standard" {
				log.Printf("Logged in ok")
				continue
			}

			// possibly breaking news?
			if msg[0].Product == "hermes" && msg[0].Project == "push" && len(msg[0].Body) > 0 {
				body, err := decodeMessageBody(msg[0].Body)
				if err != nil {
					log.Printf("decoding err: %s    msg: %s", err, msg)
					continue
				}
				log.Printf("NEW! %s  %s: %s", body.SubType, body.Label, body.Title)
			}

		default:
			log.Printf("No idea what this is: %s\n", msgBuf)
		}
	}

	log.Printf("Exiting")
}
func decodeMessageBody(body string) (res MessageBody, err error) {
	err = json.Unmarshal([]byte(body), &res)
	return
}

func decodeServerMessage(buf []byte) (res ServerMessage, err error) {
	err = json.Unmarshal(buf, &res)
	if err != nil {
		return
	}
	if len(res) == 0 {
		err = fmt.Errorf("shouldn't be empty")
	}
	return
}

func randCookie() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 30)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
