package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"io/ioutil"
	"bytes"
	"encoding/json"
)

func main() {
	fmt.Println("Procurando variavel de ambiente SLACKTOKEN")
	token := os.Getenv("SLACKTOKEN")
	if token == "" {
		fmt.Println("Cadê o token")
		return
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(token)
	fmt.Println("mybot ready, ^C exits")

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		if m.Type == "message" && (strings.HasPrefix(m.Text, "<@"+id+">") || strings.HasPrefix(m.Text, "momo")){
			parts := strings.Fields(m.Text)
			if parts[1] == "stock" {
				go func(m Message) {
					m.Text = getQuote(parts[2])
					postMessage(ws, m)
				}(m)
			} else if parts[1] == "ajuda" {
				go func(m Message) {
					m.Text = getHelp()
					postMessage(ws, m)
				}(m)
			} else if parts[1] == "wiki" {
				var wikiArg string = ""
				go func(m Message) {
					if len(parts) >= 3 {
						wikiArg = parts[2]
					}
					m.Text = getWikiPage(wikiArg)
					postMessage(ws, m)
				}(m)
			} else if parts[1] == "rodando" {
				go func(m Message) {
					m.Text = isSystemRunning(parts[2])
					postMessage(ws, m)
				}(m)
			} else if parts[1] == "trello" {
				go func(m Message) {
					m.Text = trello(parts[2])
					postMessage(ws, m)
				}(m)
			} else {
				m.Text = fmt.Sprintf("Você acha isso engraçado cara?\n")
				postMessage(ws, m)
			}
		}
	}
}

func getQuote(sym string) string {
	sym = strings.ToUpper(sym)
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1op&e=.csv", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if len(rows) >= 1 && len(rows[0]) == 5 {
		return fmt.Sprintf("%s (%s) is trading at $%s", rows[0][0], rows[0][1], rows[0][2])
	}
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}

func getHelp() string {
	return fmt.Sprintf("ajuda - Vou te falar isto que já estou te falando cara! \nwiki {palavra} - Procura por palavra nos arquivos do Wiki da BRQ\nwiki list - lista nome dos arquivos do wiki da BRQ\ntrello {palavra} - procura por palavra nos cards do Trello")
}

func getWikiPage(title string) string {
	if title == "" {
		return "https://github.com/kaiservog/brq-wiki/wiki"
	}

	//./src/github.com/rapidloop/mybot
	var wikiDir string = "wiki/brq-wiki.wiki"
	files, _ := ioutil.ReadDir(wikiDir)

	if title == "list" {
		var buffer bytes.Buffer
		buffer.WriteString("Arquivos\n")
		for _, f := range files {
			buffer.WriteString(f.Name())
			buffer.WriteString(", \n")
		}

		return buffer.String()
	}

    for _, f := range files {
    	fileName := strings.ToLower(f.Name())

		if strings.Contains(fileName, title) {
			dat, err := ioutil.ReadFile(wikiDir + "/" + f.Name())
			if err != nil {
				return "Que isso @#1c**!!, cadê as constraints!!"			
			}

			return string(dat)
		}
    }

    return "Não achei nada disso no Wiki, cadê as constraints!!"
}

func isSystemRunning(system string) string {
	switch system {
		case "moc":
			return isRunning("http://10.2.1.170:8080/simoc")
		case "gms":
			return isRunning("http://10.2.78.72:9098/sigms")
	}
	return "Nunca vi esse sistema cara???"
}

func isRunning(url string) string {
	resp, err := http.Get(url)
	notRunning := "Não ta rodando cara!"
	if err != nil || resp == nil {
		return notRunning
	}

	if resp.StatusCode != 200 {
		return notRunning
	}

	return "Tá rodando sim cara!"
}

func trello(stringQuery string) string {
	url := "https://api.trello.com/1/search?query=" + stringQuery + "&cards_limit=3&key=3db4f1a528d7d9159e603b88da5825f6&token=66d6f7006babc6c67cda870748d5cc335ea606f255edce390adacb9588356fdd"

	resp, err := http.Get(url)
	if err != nil || resp == nil {
		return "Oooops!!"
	}

	if resp.StatusCode != 200 {
		return "Ooooops!!"
	}

	defer resp.Body.Close()

	fmt.Println("card", url)

	type card struct {
		Name   string      `json:"name"`
		ShortUrl string    `json:"shortUrl"`
	}

	type TrelloResponse struct {
		Cards   []card      `json:"cards"`
	}

	res := TrelloResponse{}
	json.NewDecoder(resp.Body).Decode(&res)
    fmt.Println(res)

    var buffer bytes.Buffer
    for _, c := range res.Cards {
 		buffer.WriteString(c.Name)
 		buffer.WriteString(" - ")
 		buffer.WriteString(c.ShortUrl)
 		buffer.WriteString(" \n")
 	}

	return buffer.String()
}