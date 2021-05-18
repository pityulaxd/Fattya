package core

//Player model
type Player struct {
	Name     string `json:"Name"`
	Password string `json:"Password"`
	Status   string `json:"Status"`
	Table    string `json:"Table"`
	Seat     int    `json:"seat"`
	Address  string `json:"address"`
}

//SetupPlayer setter
func (p *Player) SetupPlayer(c Player) {
	p.Name = c.Name
	p.Password = c.Password
	p.Status = c.Status
}
