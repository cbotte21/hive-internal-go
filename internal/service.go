package internal

import (
	"github.com/cbotte21/hive-go/pb"
	schema2 "github.com/cbotte21/hive-go/schema"
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/datastore"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"golang.org/x/net/context"
)

const PollTimeSeconds = 3 // Should be higher in production to reduce bandwidth

type Hive struct {
	JwtRedeemer    *jwtParser.JwtSecret
	JudicialClient *judicial.JudicialServiceClient
	RedisClient    *datastore.RedisClient[schema2.ActiveUser]
	pb.UnimplementedHiveServiceServer
}

func NewHive(jwtRedeemer *jwtParser.JwtSecret, judicialClient *judicial.JudicialServiceClient, redisClient *datastore.RedisClient[schema2.ActiveUser]) Hive {
	return Hive{JwtRedeemer: jwtRedeemer, JudicialClient: judicialClient, RedisClient: redisClient}
}

// ForceDisconnect removes the player from the active PlayerBase
func (hive *Hive) ForceDisconnect(ctx context.Context, disconnectRequest *pb.DisconnectRequest) (*pb.DisconnectResponse, error) {
	return &pb.DisconnectResponse{}, hive.RedisClient.Publish("kicks", disconnectRequest.GetId())
}

// Online returns true if a player is online
func (hive *Hive) Online(ctx context.Context, onlineRequest *pb.OnlineRequest) (*pb.OnlineResponse, error) {
	user := schema2.ActiveUser{Id: onlineRequest.GetXId()}
	_, err := hive.RedisClient.Find(user)

	status := int32(0)
	if err == nil {
		status = 1
	}

	return &pb.OnlineResponse{Status: status}, err
}
