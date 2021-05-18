package main

import (
	spirit "Spirit"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

var (
	reader  = bufio.NewReader(os.Stdin)
	port    int
	log     = logrus.CreateLogger()
	reqS    = http.Client{}
	players = map[string]*spirit.Player{}
	lobby   = map[string]*spirit.Table{}
	msg     string
)

//Game
func startGame(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	body, err := json.Marshal(&p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("POST", lobby[p.Table].Address+"/start", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	temp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(temp, &msg)
	body, err = json.Marshal(&msg)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func selectCards(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", p.Address+"/selectcards", bytes.NewBuffer(nil))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	tempC, err := ioutil.ReadAll(response.Body)
	var cardN []int
	json.Unmarshal(tempC, &cardN)
	body, err := json.Marshal(&cardN)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func wrongcard(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", p.Address+"/wrongcard", bytes.NewBuffer(nil))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	tempC, err := ioutil.ReadAll(response.Body)
	w.Write(tempC)
}

func yeet(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", p.Address+"/yeet", bytes.NewBuffer(nil))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	tempC, err := ioutil.ReadAll(response.Body)
	w.Write(tempC)
}

func play(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player
	var c []int
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("GET", p.Address+"/play", bytes.NewBuffer(nil))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	asd, err := ioutil.ReadAll(response.Body)
	if asd == nil {
		players[p.Password] = nil
		fmt.Println(p.Name + " disconnected")
	}
	json.Unmarshal(asd, &c)
	body, err := json.Marshal(&c)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}
func winner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var stat spirit.Status
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &stat)
	if err != nil {
		log.Error(err)
	}
	body, err := json.Marshal(&stat)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("POST", players[vars["player"]].Address+"/status/"+vars["winner"], bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	temp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(temp, &msg)
	body, err = json.Marshal(&msg)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func pickUp(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	temp2, _ := json.Marshal(&temp)
	t := p.Table
	request, err := http.NewRequest("POST", lobby[t].Address+"/pickup", bytes.NewBuffer(temp2))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	var msg string
	temp, err = ioutil.ReadAll(response.Body)
	json.Unmarshal(temp, &msg)
	body, err := json.Marshal(&msg)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func statusP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var stat spirit.Status
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &stat)
	if err != nil {
		log.Error(err)
	}
	body, err := json.Marshal(&stat)
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("POST", players[vars["player"]].Address+"/status", bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	temp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(temp, &msg)
	body, err = json.Marshal(&msg)
	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

//Lobby
func login(w http.ResponseWriter, r *http.Request) {
	var p spirit.Player

	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &p)
	if err != nil {
		log.Error(err)
	}
	players[p.Password] = &p
	msg, err := json.Marshal("Welcome to Fattya " + p.Name)
	if err != nil {
		log.Error(err)
	}
	w.Write(msg)
}

func lobbyTables(w http.ResponseWriter, r *http.Request) {
	body, err := json.Marshal(&lobby)

	if err != nil {
		log.Error(err)
	}
	w.Write(body)
}

func setupTable(w http.ResponseWriter, r *http.Request) {
	var body spirit.Table
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &body)
	if err != nil {
		log.Error(err)
	}
	lobby[body.ID] = &body
	fmt.Println("tables")
	fmt.Println(lobby)
	fmt.Println(lobby[body.ID].Address)
}

func joinTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var body spirit.Table
	temp, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(temp, &body)
	if err != nil {
		log.Error(err)
	}
	if lobby[body.ID].Status == "Online" {
		lobby[body.ID].Players = append(lobby[body.ID].Players, body.Players[0])
		lobby[body.ID].VPlayers = append(lobby[body.ID].VPlayers, body.Players[0].Name)
	} else {
		temp, err := json.Marshal("The lobby is not available")
		if err != nil {
			log.Error(err)
		}
		w.Write(temp)
		return
	}
	p, err := json.Marshal(players[vars["player"]])
	if err != nil {
		log.Error(err)
	}
	request, err := http.NewRequest("PUT", lobby[body.ID].Address+"/jointable", bytes.NewBuffer(p))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-type", "application/json")
	response, err := reqS.Do(request)
	if err != nil {
		log.Error(err)
	}
	var seat int

	temp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(temp, &seat)
	body2, err := json.Marshal(&seat)
	if err != nil {
		log.Error(err)
	}
	w.Write(body2)
}

func main() {
	fmt.Println("Open port:")
	temp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("This is not a valid username")
	}
	temp = temp[0 : len(temp)-2]
	port, err = strconv.Atoi(temp)
	if err != nil {
		log.Error(err)
	}

	log.SetLevel(logger.DEBUG)
	//Create router
	router := mux.NewRouter()

	//Game
	router.HandleFunc("/start", startGame).Methods("POST") //Start a table
	router.HandleFunc("/selectcards", selectCards).Methods("GET")
	router.HandleFunc("/play", play).Methods("GET") //Play a card
	router.HandleFunc("/pickup", pickUp).Methods("POST")
	router.HandleFunc("/status/{player}", statusP).Methods("POST")
	router.HandleFunc("/wrongcard", wrongcard).Methods("GET")
	router.HandleFunc("/yeet", yeet).Methods("GET")
	router.HandleFunc("/winner/{player}/{winner}", winner).Methods("POST")

	//Lobby handlers
	router.HandleFunc("/login", login).Methods("PUT")                  //Login with existing name:password
	router.HandleFunc("/tables", lobbyTables).Methods("GET")           //Shows tables and statuses of tables
	router.HandleFunc("/setuptable", setupTable).Methods("PUT")        //Make a table
	router.HandleFunc("/jointable/{player}", joinTable).Methods("PUT") //Sit to an already made table

	log.Info("FattyaWeb listening on port: ", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
