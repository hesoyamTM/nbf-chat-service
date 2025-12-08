package redpanda

type MatcherGroupRequestAccepted struct {
	Group struct {
		GroupID    string `json:"id"`
		OwnerID    string `json:"owner_id"`
		Parameters struct {
			Name string `json:"name"`
		} `json:"parameters"`
	}
	Request struct {
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"`
	} `json:"request"`
}

type MatcherGroupUserLeft struct {
	Group struct {
		GroupID string `json:"id"`
		OwnerID string `json:"owner_id"`
	} `json:"group"`
	UserID   string `json:"user_id"`
	Remained int    `json:"remained"`
}

type MatcherGroupUserKicked struct {
	Group struct {
		GroupID string `json:"id"`
		OwnerID string `json:"owner_id"`
	} `json:"group"`
	UserID   string `json:"user_id"`
	Remained int    `json:"remained"`
}
