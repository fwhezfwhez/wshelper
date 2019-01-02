package wshelper

// Commands
const (
	SEND_ONE   = 1 + iota // send to a user
	SEND_MANY             // send to many users
	SEND_GROUP            // send to a chat group
	SEND_ROOM             // send to a room

	CREATE_GROUP  // create a chat group
	CREATE_ROOM   // create a chat room

	ADD_ONE     // add a user as friend
	ADD_MANY    // add many users as friends
	JOIN_ROOM   // join a chat room
	JOIN_GROUP  // join a chat group

	DELETE_GROUP  // delete a chat group
	DELETE_ROOM   // delete a chat room
	DELETE_ONE    // delete a friend
	DELETE_MANY   // delete many friends

	CREATE_LISTGROUP  // create a friend group,like 'family', 'workmate'...
	CREATE_ROOMGROUP  // create a room group, like 'family', 'workmate', 'game' ...

	ADD_ONE_REMARK    // remark a friend
	ADD_ROOM_REMARK   //remark a room
	ADD_GROUP_REMART  //remart a group
)

// SubCommands
// send types
const(
	TEXT = 1 + iota
	VOICE
	IMAGE
	FILE
	FILEFOLDER
	VIDEO
	URL
)
