package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fattya/core"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	server = "http://80.98.39.90:6070"
	port   = 6071
	//DrawDeck from this you draw
	reader      = bufio.NewReader(os.Stdin)
	myTable     core.Table
	DrawDeck    []core.Card
	stack       []core.Card
	pHands      = make(map[string][]core.Card)
	endCards    = make(map[string][]core.Card)
	hiddenCards = make(map[string][]core.Card)
	i           = 1
	comp        core.Card
	turn        = 0
	gs          = 1
	//Asked player number need to be in request
	askP    = 0
	ready   = 0
	client  http.Client
	empty   core.Card
	lastTen = false
)

//
//
//
//Game logic

func autoBurn(c core.Card) bool {
	if c.Type == comp.Type {
		i++
	} else {
		comp = c
		i = 1
	}
	if i >= 4 {
		stack = nil
		return true
	}
	return false
}

//Function to check 9 on stack
func limit(c core.Card) bool {
	if c.Type >= 3 && c.Type <= 7 {
		stack = append(stack, c)
		return true
	} else {
		return false
	}
}

//Function to check 8 on stack
func glass(c core.Card) bool {
	if c.Type == 8 {
		stack = append(stack, c)
		gs++
		fmt.Println(gs)
		return true
	} else if len(stack) != 0 {
		if stack[len(stack)-gs].Type <= c.Type {
			stack = append(stack, c)
			gs = 1
			return true
		}
	} else {
		stack = append(stack, c)
	}
	return false
}

//Function to check 1 on stack
func isMagicCard(c core.Card) bool {
	switch {
	case c.Type == 2:
		stack = append(stack, c)
		return true
	case c.Type == 8:
		stack = append(stack, c)
		return true
	case c.Type == 10:
		stack = append(stack, c)
		burn()
		lastTen = true
		return false
	case c.Type == 1:
		stack = append(stack, c)
		ask(askP)
		return true
	default:
		return false
	}
}

func magicCard(c core.Card) bool {
	switch {
	case c.Type == 2:
		return true
	case c.Type == 8:
		return true
	case c.Type == 10:
		return true
	case c.Type == 1:
		return true
	default:
		return false
	}
}

//Execute 1
func ask(p int) {
	turn = p - 1
}

//Function to execute burn on stack(10)
func burn() bool {
	stack = nil
	return false
}

//Regular play logic check
func valid(c core.Card) bool {
	if magicCard(c) {
		return isMagicCard(c)
	} else {
		if stack[len(stack)-1].Type <= c.Type {
			stack = append(stack, c)
			return true
		} else {
			return false
		}
	}
}

func validate(c core.Card) bool {
	switch {
	case stack[len(stack)-1].Type == 9:
		return limit(c)
	case stack[len(stack)-1].Type == 8:
		return glass(c)
	case stack[len(stack)-1].Type == 1:
		return isMagicCard(c)
	default:
		t := valid(c)
		if autoBurn(c) {
			return false
		}
		return t
	}
}

func invalidCard(c core.Card, w http.ResponseWriter) {
	body, err := json.Marshal(c)

	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

//Game
func pickUp(p core.Player) {
	all := len(stack)
	if all > 0 {
		for i := 0; i < all; i++ {
			pHands[p.Password] = append(pHands[p.Password], stack[i])
		}
		stack = nil
	}
}

func setupTable(t core.Table) {
	body, err := json.Marshal(&t)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(server + "/setuptable")
	request, err := http.NewRequest("PUT", server+"/setuptable", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	defer response.Body.Close()
}

func joinTable(w http.ResponseWriter, r *http.Request) {
	var p core.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	p.Seat = turn
	turn++
	myTable.Players = append(myTable.Players, p)
	myTable.VPlayers = append(myTable.VPlayers, p.Name)
	body, err := json.Marshal(&p.Seat)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func start(w http.ResponseWriter, r *http.Request) {
	var p core.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	for k := 0; k < 3; k++ {
		hiddenCards[p.Password] = append(hiddenCards[p.Password], core.Draw(DrawDeck))
	}
	for l := 0; l < 6; l++ {
		pHands[p.Password] = append(pHands[p.Password], core.Draw(DrawDeck))
	}
	selectCards(p)

	if ready == len(myTable.Players) {
		myTable.Status = "Ingame"
		turn = 0
		statusP(p)
		body, err := json.Marshal("Starting...")
		if err != nil {
			log.Error(err)
		}
		w.Write(body)
		play(myTable.Players[turn])
	} else {
		statusP(p)
		body, err := json.Marshal("Waiting for " + strconv.Itoa(len(myTable.Players)-ready) + " more players...")
		if err != nil {
			log.Error(err)
		}
		w.Write(body)
	}
}

func remove(s []core.Card, i int) []core.Card {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func selectCards(p core.Player) {
	statusP(p)
	body, err := json.Marshal(&p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", server+"/selectcards", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	tempC, err := ioutil.ReadAll(response.Body)
	var cardN []int
	json.Unmarshal(tempC, &cardN)
	fmt.Println(cardN)
	for i := 0; i < len(cardN); i++ {
		endCards[p.Password] = append(endCards[p.Password], pHands[p.Password][cardN[i]])
	}
	sort.Sort(sort.Reverse(sort.IntSlice(cardN)))
	for j := 0; j < 3; j++ {
		pHands[p.Password] = remove(pHands[p.Password], cardN[j])
	}
	ready++
}

func wrongcard(p core.Player) {
	body, err := json.Marshal(&p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", server+"/wrongcard", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	defer response.Body.Close()
	play(p)
}

func yeet(p core.Player) {
	for j := 0; j < len(pHands[p.Password]); j++ {
		if len(stack) == 0 && lastTen {
			var msg string
			body, err := json.Marshal(&p)
			if err != nil {
				log.Error(err)
			}
			request, err := http.NewRequest("GET", server+"/yeet", bytes.NewBuffer(body))
			if err != nil {
				log.Error(err)
			}
			request.Header.Set("Content-type", "application/json")
			response, err := client.Do(request)
			if err != nil {
				log.Error(err)
			}
			tempC, err := ioutil.ReadAll(response.Body)
			json.Unmarshal(tempC, &msg)
			if msg == "yes" {
				stack = append(stack, pHands[p.Password][j])
				pHands[p.Password] = remove(pHands[p.Password], j)
				if len(DrawDeck) != 0 {
					if len(pHands[p.Password]) < 3 {
						i := 3 - len(pHands[p.Password])
						for j := 0; j < i; j++ {
							card := core.Draw(DrawDeck)
							fmt.Println(card)
							pHands[p.Password] = append(pHands[p.Password], card)
						}
					}
				}
				yeet(p)
			}
		} else if stack[len(stack)-1].Type == pHands[p.Password][j].Type {
			var msg string
			body, err := json.Marshal(&p)
			if err != nil {
				log.Error(err)
			}
			request, err := http.NewRequest("GET", server+"/yeet", bytes.NewBuffer(body))
			if err != nil {
				log.Error(err)
			}
			request.Header.Set("Content-type", "application/json")
			response, err := client.Do(request)
			if err != nil {
				log.Error(err)
			}
			tempC, err := ioutil.ReadAll(response.Body)
			json.Unmarshal(tempC, &msg)
			if msg == "yes" {
				stack = append(stack, pHands[p.Password][j])
				pHands[p.Password] = remove(pHands[p.Password], j)
				if len(DrawDeck) != 0 {
					if len(pHands[p.Password]) < 3 {
						i := 3 - len(pHands[p.Password])
						for j := 0; j < i; j++ {
							card := core.Draw(DrawDeck)
							fmt.Println(card)
							pHands[p.Password] = append(pHands[p.Password], card)
						}
					}
				}
				yeet(p)
			}
		}
	}
}
func play(p core.Player) {
	var c []int
	body, err := json.Marshal(&p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", server+"/play", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	temp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(temp, &c)
	if c[0] != -1 {
		if len(c) > 1 {
			askP = c[1]
		}
		//Logic
		//
		//First card
		if len(stack) == 0 {
			stack = append(stack, pHands[p.Password][c[0]])
			pHands[p.Password] = remove(pHands[p.Password], c[0])
			if turn == len(myTable.Players)-1 {
				turn = 0
			} else {
				turn++
			}
			//Other cards
		} else if validate(pHands[p.Password][c[0]]) {
			if turn == len(myTable.Players)-1 {
				turn = 0
			} else {
				turn++
			}
			pHands[p.Password] = remove(pHands[p.Password], c[0])
		} else if len(stack) != 0 {
			wrongcard(p)
		} else {
			pHands[p.Password] = remove(pHands[p.Password], c[0])
		}
		//Draw check and draw
		if len(pHands[p.Password]) < 3 {
			i := 3 - len(pHands[p.Password])
			for j := 0; j < i; j++ {
				card := core.Draw(DrawDeck)
				fmt.Println(card)
				pHands[p.Password] = append(pHands[p.Password], card)
			}
		}
		yeet(p)
	} else {
		pickUp(p)
		if turn == len(myTable.Players)-1 {
			turn = 0
		} else {
			turn++
		}
	}
	for _, element := range myTable.Players {
		statusP(element)
	}
	if askP != -1 {
		if len(DrawDeck) != 0 {
			fmt.Println(turn)
			play(myTable.Players[turn])
		} else {
			endGame(myTable.Players[turn])
		}
	} else {
		if len(DrawDeck) != 0 {
			fmt.Println(turn)
			play(myTable.Players[askP])
		} else {
			endGame(myTable.Players[askP])
		}
	}
	lastTen = false
	askP = -1
}

func endGame(p core.Player) {
	var c []int
	body, err := json.Marshal(&p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", server+"/play", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	temp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(temp, &c)
	if c[0] != -1 {
		if len(c) > 1 {
			askP = c[1]
		}
		//Logic
		//
		//First card
		if len(pHands[p.Password]) != 0 {
			if len(stack) == 0 {
				stack = append(stack, pHands[p.Password][c[0]])
				pHands[p.Password] = remove(pHands[p.Password], c[0])
				//Other cards
			} else if validate(pHands[p.Password][c[0]]) {
				if turn == len(myTable.Players)-1 {
					turn = 0
				} else {
					turn++
				}
				pHands[p.Password] = remove(pHands[p.Password], c[0])
			} else if len(stack) != 0 {
				wrongcard(p)
			} else {
				pHands[p.Password] = remove(pHands[p.Password], c[0])
			}
		} else if len(endCards[p.Password]) != 0 {
			if len(stack) == 0 {
				stack = append(stack, endCards[p.Password][c[0]])
				endCards[p.Password] = remove(endCards[p.Password], c[0])
			} else if validate(endCards[p.Password][c[0]]) {
				if turn == len(myTable.Players)-1 {
					turn = 0
				} else {
					turn++
				}
				endCards[p.Password] = remove(endCards[p.Password], c[0])
			} else if len(stack) != 0 {
				wrongcard(p)
			} else {
				endCards[p.Password] = remove(endCards[p.Password], c[0])
			}
		} else if len(hiddenCards[p.Password]) != 0 {
			if hiddenCards[p.Password][c[0]].Type == 1 {
				var ask int
				request, err := http.NewRequest("GET", server+"/hiddenace", bytes.NewBuffer(nil))
				if err != nil {
					log.Error(err)
				}
				request.Header.Set("Content-type", "application/json")
				response, err := client.Do(request)
				if err != nil {
					log.Error(err)
				}
				temp, err := ioutil.ReadAll(response.Body)
				if err != nil {
					log.Error(err)
				}
				json.Unmarshal(temp, &ask)

			}
			if len(stack) == 0 {
				stack = append(stack, hiddenCards[p.Password][c[0]])
				hiddenCards[p.Password] = remove(hiddenCards[p.Password], c[0])
				//Other cards
			} else if validate(hiddenCards[p.Password][c[0]]) {
				if turn == len(myTable.Players)-1 {
					turn = 0
				} else {
					turn++
				}
				hiddenCards[p.Password] = remove(hiddenCards[p.Password], c[0])
			} else if len(stack) != 0 {
				wrongcard(p)
			} else {
				hiddenCards[p.Password] = remove(hiddenCards[p.Password], c[0])
			}
			if len(hiddenCards[p.Password]) != 0 {
				endGame(myTable.Players[turn])
			} else {
				winner(p)
			}
			yeet(p)
		}
	} else {
		pickUp(p)
		if turn == len(myTable.Players)-1 {
			turn = 0
		} else {
			turn++
		}
		endGame(myTable.Players[turn])
	}
	for _, element := range myTable.Players {
		statusP(element)
	}
}

func winner(p core.Player) {
	for _, element := range myTable.Players {
		temp := hiddenCards[element.Password]
		body, err := json.Marshal(&temp)
		if err != nil {
			log.Error(err)
		}
		request, err := http.NewRequest("POST", server+"/winner/"+element.Password+p.Name, bytes.NewBuffer(body))
		if err != nil {
			log.Error(err)
		}
		request.Header.Set("Content-type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			log.Error(err)
		}
		defer response.Body.Close()
	}
}

func statusP(p core.Player) {
	var stat core.Status
	stat.EndCards = endCards[p.Password]
	stat.StackCards = stack
	stat.HandCards = pHands[p.Password]
	body, err := json.Marshal(&stat)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("POST", server+"/status/"+p.Password, bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}
	defer response.Body.Close()
}

func myip() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "ouf"
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	//Set up logger
	log.SetLevel(log.DebugLevel)
	myTable.Address = "http://" + myip() + ":" + strconv.Itoa(port)
	//Create router and endpoints
	/*
		fmt.Println("Main Server IP:(example: http://localhost:port):")
		server, err := reader.ReadString('\n')
		if err != nil {
			log.Error(err)
		}
		server = strings.Trim(server, "\r\n")

		fmt.Println("Server port:")
		temp, err := reader.ReadString('\n')
		if err != nil {
			log.Error(err)
		}
		temp = temp[0 : len(temp)-2]
		port, err = strconv.Atoi(temp)
		if err != nil {
			log.Error(err)
		}
	*/
	router := mux.NewRouter()

	DrawDeck = core.New(2)
	core.Shuffle(DrawDeck)

	fmt.Println("Tablename:")
	temp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("This is not a valid name")
	}
	temp = temp[0 : len(temp)-2]
	myTable.ID = temp

	fmt.Println("Max player:")
	temp, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error")
	}

	temp = temp[0 : len(temp)-2]
	myTable.SeatNum, err = strconv.Atoi(temp)
	if err != nil {
		fmt.Println("This is not a correct value")
	}

	fmt.Println("Deck size:")
	temp, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error")
	}

	temp = temp[0 : len(temp)-2]
	myTable.Decks, err = strconv.Atoi(temp)
	if err != nil {
		fmt.Println("This is not a correct value")
	}
	myTable.Status = "Online"
	setupTable(myTable)
	fmt.Println(myTable.Address)

	//Game
	router.HandleFunc("/start", start).Methods("POST")
	//Table
	router.HandleFunc("/jointable", joinTable).Methods("PUT")

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
	log.Info("FattyaServer listening on port: ", port)

}
