package telegram

type Update struct {
	UpdateID     int            `json:"update_id"`
	Message      *Message       `json:"message"`
	MyChatMember *ChatMemberUpd `json:"my_chat_member"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	From      *User  `json:"from"`
	Text      string `json:"text"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type ChatMemberUpd struct {
	Chat      Chat       `json:"chat"`
	NewMember ChatMember `json:"new_chat_member"`
}

type ChatMember struct {
	Status string `json:"status"` // "left", "kicked", "member", "administrator"
}
