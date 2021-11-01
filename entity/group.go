package entity

type Group struct {
	Id           string         `json:"id"`
	Participants []Participants `json:"participants"`
}

type Participants struct {
	Id string `json:"id"`
}
