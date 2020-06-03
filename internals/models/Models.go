package models

type Error struct {
	Message string `json:"message"`
}

type Forum struct {
	Posts   int    `json:"posts"`
	Slug    string `json:"slug"`
	Threads int    `json:"threads"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

type Post struct {
	Author   string `json:"author"`
	Created  string `json:"created"`
	Forum    string `json:"forum"`
	Id       int    `json:"id"`
	IsEdited bool   `json:"isEdited"`
	Message  string `json:"message"` // updated
	Parent   int    `json:"parent"`
	Thread   int    `json:"thread"`
}

type Status struct {
	Forum  uint `json:"forum"`
	Post   uint `json:"post"`
	Thread uint `json:"thread"`
	User   uint `json:"user"`
}

type Thread struct {
	Author  string `json:"author"`
	Created string `json:"created"`
	Forum   string `json:"forum"`
	Id      int    `json:"id"`
	Message string `json:"message"` // updated
	Slug    string `json:"slug"`
	Title   string `json:"title"` // updated
	Votes   int    `json:"votes"`
}

type User struct {
	About    string `json:"about"`    // updated
	Email    string `json:"email"`    // updated
	FullName string `json:"fullname"` // updated
	NickName string `json:"nickname"`
}

type Vote struct {
	NickName string `json:"nickname"`
	Voice    int    `json:"voice"`
}

type PostFull struct {
	Author *User   `json:"author"`
	Forum  *Forum  `json:"forum"`
	Post   *Post   `json:"post"`
	Thread *Thread `json:"thread"`
}
