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
	ID             string `json:"id,omitempty" form:"id"`
	User           string `json:"user" form:"user"`
	City           string `json:"city" form:"city"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Country        string `json:"country" form:"country"`
	Created        string `json:"created" form:"created"`
	Description    string `json:"description" form:"description"`
	Name           string `json:"name" form:"name"`
	Sport          string `json:"sport" form:"sport"`
	Street         string `json:"street" form:"street"`
	Updated        string `json:"updated" form:"updated"`
	Expand         struct {
		Members []Member `json:"members_via_group" form:"members_via_group"`
		Seasons []Season `json:"seasons_via_group" form:"seasons_via_group"`
		Sport   Sport    `json:"sport" form:"sport"`
	} `json:"expand,omitempty" form:"expand"`
}

type Season struct {
	ID             string `json:"id,omitempty" form:"id"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Created        string `json:"created" form:"created"`
	Updated        string `json:"updated" form:"updated"`
	Name           string `json:"name" form:"name"`
	Status         string `json:"status" form:"status"`
	StartDate      string `json:"start_date" form:"start_date"`
	EndDate        string `json:"end_date" form:"end_date"`
	Group          string `json:"group" form:"group"`
	Expand         struct {
		Group Group `json:"group" form:"group"`
	} `json:"expand,omitempty" form:"expand"`
}

type Member struct {
	ID             string `json:"id,omitempty" form:"id"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Created        string `json:"created" form:"created"`
	Updated        string `json:"updated" form:"updated"`
	Group          string `json:"group" form:"group"`
	Username       string `json:"username" form:"username"`
	Email          string `json:"email" form:"email"`
	Phone          string `json:"phone" form:"phone"`
	Expand         struct {
		Group Group `json:"group" form:"group"`
	} `json:"expand,omitempty" form:"expand"`
}

type MemberStat struct {
	ID             string    `json:"id,omitempty" form:"id"`
	CollectionID   string    `json:"collectionId"`
	CollectionName string    `json:"collectionName"`
	Created        string    `json:"created" form:"created"`
	Updated        string    `json:"updated" form:"updated"`
	Group          string    `json:"group" form:"group"`
	Member         string    `json:"member" form:"member"`
	Season         string    `json:"season" form:"season"`
	Stats          SportStat `json:"stats" form:"stats"`
	Expand         struct {
		Group  Group  `json:"group" form:"group"`
		Member Member `json:"member" form:"member"`
		Season Season `json:"season" form:"season"`
	} `json:"expand,omitempty" form:"expand"`
}
