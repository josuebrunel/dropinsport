package models

type Collection[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalPages int `json:"totalPages"`
	TotalItems int `json:"totalItems"`
	Items      []T `json:"items"`
}

type User struct {
	ID              string `json:"id,omitempty"`
	CollectionID    string `json:"collectionId"`
	CollectionName  string `json:"collectionName"`
	Username        string `json:"username"`
	Verified        bool   `json:"verified"`
	EmailVisibility bool   `json:"emailVisibility"`
	Email           string `json:"email"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
}

type UserExpandGroup struct {
	User
	Expand struct {
		Groups []Group `json:"groups_via_user"`
	} `json:"expand,omitempty"`
}

type SportStat struct {
	Abbr string `json:"abbr"`
	Name string `json:"name"`
	Step string `json:"step"`
	Type string `json:"type"`
}
type SportData struct {
	Icon  string      `json:"icon"`
	Stats []SportStat `json:"stats"`
}
type Sport struct {
	ID             string    `json:"id,omitempty"`
	CollectionID   string    `json:"collectionId"`
	CollectionName string    `json:"collectionName"`
	Created        string    `json:"created"`
	Data           SportData `json:"data"`
	Icon           string    `json:"icon"`
	Name           string    `json:"name"`
	Updated        string    `json:"updated"`
}

type Group struct {
	ID             string `json:"id,omitempty"`
	User           string `json:"user"`
	City           string `json:"city"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Country        string `json:"country"`
	Created        string `json:"created"`
	Description    string `json:"description"`
	Name           string `json:"name"`
	Sport          string `json:"sport"`
	Street         string `json:"street"`
	Updated        string `json:"updated"`
	Expand         struct {
		Members []Member `json:"members_via_group"`
		Seasons []Season `json:"seasons_via_group"`
		Sport   Sport    `json:"sport"`
	} `json:"expand,omitempty"`
}

type Season struct {
	ID             string `json:"id,omitempty"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	Group          string `json:"group"`
	Expand         struct {
		Group Group `json:"group"`
	} `json:"expand,omitempty"`
}

type Member struct {
	ID             string `json:"id,omitempty"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
	Group          string `json:"group"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Expand         struct {
		Group Group `json:"group"`
	} `json:"expand,omitempty"`
}

type MemberStat struct {
	ID             string    `json:"id,omitempty"`
	CollectionID   string    `json:"collectionId"`
	CollectionName string    `json:"collectionName"`
	Created        string    `json:"created"`
	Updated        string    `json:"updated"`
	Group          string    `json:"group"`
	Member         string    `json:"member"`
	Season         string    `json:"season"`
	Stats          SportStat `json:"stats"`
	Expand         struct {
		Group  Group  `json:"group"`
		Member Member `json:"member"`
		Season Season `json:"season"`
	} `json:"expand,omitempty"`
}
