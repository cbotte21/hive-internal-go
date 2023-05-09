package playerbase

import (
	"errors"
	"github.com/cbotte21/hive-go/pb"
)

type Player struct {
	Comm *pb.HiveService_ConnectServer
	Id   string
}

type PlayerBase map[string]Player

// AppendUnique follows a delete existing policy
func (playerBase *PlayerBase) AppendUnique(jwt, _id string, comm *pb.HiveService_ConnectServer) {
	playerBase.Disconnect(_id)
	(*playerBase)[jwt] = Player{Comm: comm, Id: _id}
}

// Disconnect removes a player from the active players map
func (playerBase *PlayerBase) Disconnect(jwt string) {
	delete(*playerBase, jwt)
}

// ForceDisconnect removes a player from the active players map
func (playerBase *PlayerBase) ForceDisconnect(_id string) {
	jwt := ""
	for token, player := range *playerBase {
		if player.Id == _id {
			_ = (*player.Comm).Send(&pb.ConnectionStatus{Status: 0})
			jwt = token
		}
	}
	playerBase.Disconnect(jwt)
}

// Online returns true if a player is online
func (playerBase *PlayerBase) Online(_id string) error {
	for k := range *playerBase {
		if (*playerBase)[k].Id == _id {
			return nil
		}
	}
	return errors.New("player is not online")
}

// GetId returns the XID belonging to a player
func (playerBase *PlayerBase) GetId(jwt string) (string, error) {
	player, found := (*playerBase)[jwt]
	if found {
		return player.Id, nil
	}
	return "", errors.New("player is not online")
}
