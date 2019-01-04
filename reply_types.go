package wshelper

const (
	// chat message
	REPLY_MESSAGE = 1 + iota
	// notifications
	REPLY_NOTIFY
	// tips after send some message like 'Tom is offline. message will be received after he/she 's online'
	REPLY_TIPS
	// debug message showed in a specific message box
	REPLY_DEBUG
)
