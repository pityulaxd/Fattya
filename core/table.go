package core

//Table is where you group players
type Table struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Players  []Player `json:"player"`
	Address  string   `json:"address"`
	VPlayers []string `json:"vplayers"` //Visible players only for screen purposes
	SeatNum  int      `json:"seatnum"`
	Decks    int      `json:"decks"` //Number of starting decks
}

//SetPlayerStatus setting all sitting players status
func (t Table) SetPlayerStatus(s string) {
	for i := range t.Players {
		t.Players[i].Status = s
	}
}

//SetupTable make table
//Table name,Table status, Player,Starting deck
func (t *Table) SetupTable(tableID string, status string, player Player, seatNum int, numOfDecks int) *Table {
	t.ID = tableID
	t.Status = status
	t.Players = append(t.Players, player)
	t.VPlayers = append(t.VPlayers, player.Name)
	t.SeatNum = seatNum
	t.Decks = numOfDecks
	return t
}
