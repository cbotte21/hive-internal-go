package internal

import (
	"errors"
	"fmt"
	"github.com/cbotte21/hive-go/pb"
	schema2 "github.com/cbotte21/hive-go/schema"
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/datastore"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"golang.org/x/net/context"
	"time"
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

// Connect appends the player to the active players, and disconnects upon stream close
func (hive *Hive) Connect(connectRequest *pb.ConnectRequest, stream pb.HiveService_ConnectServer) error {
	// Check JWT authenticity
	res, err := hive.JwtRedeemer.Redeem(connectRequest.GetJwt())
	if err != nil {
		return err
	}

	// Check if player is banned
	integrity, err := (*hive.JudicialClient).Integrity(context.Background(), &judicial.IntegrityRequest{XId: res.Id})
	if err != nil {
		return err
	}
	if !integrity.Status {
		return errors.New("player is banned")
	}

	// All checks passed, connect

	user := schema2.ActiveUser{Id: res.Id, Jwt: connectRequest.GetJwt(), Role: res.Role}
	//Create user in redis cache
	err = hive.RedisClient.Create(user)
	if err != nil {
		return err
	}

	fmt.Println("[+] " + res.Id)

	kicked := 0
	go func(kicked *int) {
		sub := hive.RedisClient.Subscribe("kicks")
		ch := sub.Channel()
		for msg := range ch {
			if msg.Payload == res.Id {
				*kicked = 1
				break
			}
		}
	}(&kicked)

	// While connected loop
	for stream.Send(&pb.ConnectionStatus{Status: 1}) == nil && kicked == 0 {
		time.Sleep(PollTimeSeconds * time.Second)
	}

	//Send disconnect message if kicked
	if kicked == 1 {
		_ = stream.Send(&pb.ConnectionStatus{Status: 0})
	}

	//Remove user from redis cache
	_ = hive.RedisClient.Delete(user)
	fmt.Println("[-] " + res.Id)

	return nil
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
