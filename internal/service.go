package internal

import (
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
	res, err := hive.JwtRedeemer.Redeem(connectRequest.GetJwt())
	if err == nil {
		integrity, err := (*hive.JudicialClient).Integrity(context.Background(), &judicial.IntegrityRequest{XId: res.Id})
		if err == nil {
			if integrity.Status {
				user := schema2.ActiveUser{Id: res.Id, Jwt: connectRequest.GetJwt(), Role: res.Role}
				//Create user in redis cache
				err = hive.RedisClient.Create(user)
				if err != nil {
					return err
				}

				fmt.Println("[+] " + res.Id)

				// While connected loop
				for err == nil && stream.Send(&pb.ConnectionStatus{Status: 1}) == nil {
					time.Sleep(PollTimeSeconds * time.Second)
					_, err = hive.RedisClient.Find(user)
				}

				//Remove user from redis cache
				_ = hive.RedisClient.Delete(user)
				fmt.Println("[-] " + res.Id)
			}
		}
	}
	return err
}

// ForceDisconnect removes the player from the active PlayerBase
func (hive *Hive) ForceDisconnect(ctx context.Context, disconnectRequest *pb.DisconnectRequest) (*pb.DisconnectResponse, error) {
	user := schema2.ActiveUser{Id: disconnectRequest.GetId()}
	err := hive.RedisClient.Delete(user)
	return &pb.DisconnectResponse{}, err
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
