package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fattya/core"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/zenazn/goji/graceful"
)

var (
	router    = mux.NewRouter()
	port      = 8080
	player    core.Player
	server    = "http://80.98.39.90:6070"
	client    = http.Client{}
	reader    = bufio.NewReader(os.Stdin)
	table     core.Table
	err       error
	playTable core.Table
	msg       string
	lobby     = map[string]*core.Table{}
	hand      []core.Card
)

//Account
func login() {
	body, err := json.Marshal(&player)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("PUT", server+"/login", bytes.NewBuffer(body))
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
	json.Unmarshal(temp, &msg)
	fmt.Println(msg)
	player.Status = "Online"
}

//Game
func start() {
	fmt.Println("Start")
	player.Status = "At table"
	body, err := json.Marshal(&player)
	if err != nil {
		log.Error(err)
	}

	request, err := http.NewRequest("POST", server+"/start", bytes.NewBuffer(body))
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
	json.Unmarshal(temp, &msg)
	fmt.Println(msg)
}

func selectCards(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SELECT")
	fmt.Println("Please select your endgame cards(1 by 1 with enter)")
	var c []int
	for i := 0; i < 3; i++ {
		temp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error")
		}
		temp = temp[0 : len(temp)-2]
		tempC, err := strconv.Atoi(temp)
		tempC--
		c = append(c, tempC)
	}
	body, err := json.Marshal(&c)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func wrongcard(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Invalid card try another")
}

func yeet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("You've have a same type of card")
	fmt.Println("Would you like to play it?")
	temp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error")
	}
	temp = temp[0 : len(temp)-2]
	body, err := json.Marshal(temp)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func play(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Play")
	var c []int
	fmt.Println("It`s your turn!")
	temp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error")
	}
	temp = temp[0 : len(temp)-2]
	if temp == "pickup" {
		c = append(c, -1)
		body, err := json.Marshal(&c)
		if err != nil {
			log.Error(err)
		}
		w.Write(body)
		return
	} else {
		tempN, err := strconv.Atoi(temp)
		if err != nil {
			log.Error(err)
		}
		c = append(c, tempN)
		c[0]--
		if hand[c[0]].Type == 1 {
			var ask int
			fmt.Println("Ask:")
			temp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Input error")
			}
			temp = temp[0 : len(temp)-2]
			ask, err = strconv.Atoi(temp)
			c = append(c, ask)
		}

		body, err := json.Marshal(c)
		if err != nil {
			log.Error(err)
		}
		w.Write(body)
	}
}
func statusP(w http.ResponseWriter, r *http.Request) {
	var stat core.Status
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &stat)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("Board:")
	fmt.Println(stat.StackCards)
	if len(stat.HandCards) != 0 {
		fmt.Println("Hand:")
		fmt.Println(stat.HandCards)
		hand = stat.HandCards
	} else if len(stat.EndCards) != 0 {
		fmt.Println("End cards:")
		fmt.Println(stat.EndCards)
		hand = stat.EndCards
	} else {
		fmt.Println("Hidden cards:")
	}
}

//Table
func tables() {
	request, err := http.NewRequest("GET", server+"/tables", bytes.NewBuffer(nil))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}

	temp, err := ioutil.ReadAll(response.Body)

	json.Unmarshal(temp, &lobby)
	if err != nil {
		log.Error(err)
	}
	for _, element := range lobby {
		fmt.Println("\n", "Name:", element.ID, "\n", "Status:", element.Status, "\n", "Decks:", element.Decks, "\n", "Players:", element.VPlayers, "\n")
	}
}

func jointable() {
	fmt.Println("Table name:")
	var id string
	id, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("This is not a valid name")
	}
	id = id[0 : len(id)-2]

	table.SetupTable(id, "Online", player, 0, 0)

	body, err := json.Marshal(&table)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("PUT", server+"/jointable/"+player.Password, bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
	}

	temp, err := ioutil.ReadAll(response.Body)
	json.Unmarshal(temp, &player.Seat)
	if err != nil {
		log.Error(err)
	}
	player.Table = table.ID
	fmt.Println("Sitting at table " + table.ID)
}

func winner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var hidden []core.Card
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &hidden)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("Your hidden cards:")
	fmt.Println(hidden)
	fmt.Println("Winner:")
	fmt.Println(vars["winner"])
	graceful.ListenAndServe(":"+strconv.Itoa(port), router)
	player.Status = "Online"
}

func hiddenAce(w http.ResponseWriter, r *http.Request) {
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &msg)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("Your hidden card was an Ace, who do u wanna ask?")
	var answer int
	temp2, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("This is not a valid username")
	}
	temp2 = temp2[0 : len(temp2)-2]
	answer, err = strconv.Atoi(temp2)

	body, err := json.Marshal(answer)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func cServer() {
	router.HandleFunc("/play", play).Methods("GET")
	router.HandleFunc("/selectcards", selectCards).Methods("GET")
	router.HandleFunc("/status", statusP).Methods("POST")
	router.HandleFunc("/yeet", yeet).Methods("GET")
	router.HandleFunc("/wrongcard", wrongcard).Methods("GET")
	router.HandleFunc("/winner/{winner}", winner).Methods("POST")
	router.HandleFunc("/hiddenace", hiddenAce).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))

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

func main() {

	fmt.Println("Username:")
	player.Name, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("This is not a valid username")
	}
	player.Name = player.Name[0 : len(player.Name)-2]

	fmt.Println("ID:")
	player.Password, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("This is not a valid ID")
	}
	player.Password = player.Password[0 : len(player.Password)-2]
	player.Address = "http://" + myip() + ":" + strconv.Itoa(port)
	/*
		fmt.Println("Main Server IP:(example: http://localhost:port):")
		server, err = reader.ReadString('\n')
		if err != nil {
			log.Error(err)
		}
		server = server[0 : len(server)-2]

		fmt.Println("This users open port for communication:")
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
	go cServer()

	login()

	var input string
	fmt.Println(player.Status)
	for input != "quit" && player.Status == "Online" {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error")
		}
		input = input[0 : len(input)-2]
		switch {
		case input == "join" || input == "joint":
			jointable()
		case input == "lobby":
			tables()
		case input == "start":
			start()
		default:
			fmt.Println("U fucked up")
		}
	}
	for true {
	}
}
