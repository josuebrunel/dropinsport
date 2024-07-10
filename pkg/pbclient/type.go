package pbclient

type RequestAuth struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type UserRecord struct {
	ID              string `json:"id"`
	CollectionID    string `json:"collectionId"`
	CollectionName  string `json:"collectionName"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Verified        bool   `json:"verified"`
	EmailVisibility bool   `json:"emailVisibility"`
}

type ResponseAuth struct {
	Token  string     `json:"token"`
	Record UserRecord `json:"record"`
}

type RecordBase struct {
	ID             string `json:"id"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Updated        string `json:"updated"`
	Created        string `json:"created"`
}

type Records[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}
