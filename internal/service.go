package internal

import (
	"fmt"
	"github.com/cbotte21/hive-go/internal/playerbase"
	"github.com/cbotte21/hive-go/pb"
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"golang.org/x/net/context"
	"time"
)

const PollTimeSeconds = 3

type Hive struct {
	PlayerBase     *playerbase.PlayerBase
	JwtRedeemer    *jwtParser.JwtSecret
	JudicialClient *judicial.JudicialServiceClient
	pb.UnimplementedHiveServiceServer
}

func NewHive(playerBase *playerbase.PlayerBase, jwtRedeemer *jwtParser.JwtSecret, judicialClient *judicial.JudicialServiceClient) Hive {
	return Hive{PlayerBase: playerBase, JwtRedeemer: jwtRedeemer, JudicialClient: judicialClient}
}

// Connect appends the player to the active players, and disconnects upon stream close
func (hive *Hive) Connect(connectRequest *pb.ConnectRequest, stream pb.HiveService_ConnectServer) error {
	res, err := hive.JwtRedeemer.Redeem(connectRequest.GetJwt())
	if err == nil {
		integrity, err := (*hive.JudicialClient).Integrity(context.Background(), &judicial.IntegrityRequest{XId: res.Id})
		if err == nil {
			if integrity.Status {
				hive.PlayerBase.AppendUnique(connectRequest.GetJwt(), res.Id, &stream)
				fmt.Println("[+] " + res.Id)
				// While connected loop
				for hive.PlayerBase.Online(res.Id) == nil && stream.Send(&pb.ConnectionStatus{Status: 1}) == nil {
					time.Sleep(PollTimeSeconds * time.Second)
				}
			}
		} else {
			return err
		}
	} else {
		return err
	}
	hive.PlayerBase.Disconnect(res.Id)
	fmt.Println("[-] " + res.Id)
	return nil
}

// ForceDisconnect removes the player from the active PlayerBase
func (hive *Hive) ForceDisconnect(ctx context.Context, disconnectRequest *pb.DisconnectRequest) (*pb.DisconnectResponse, error) {
	hive.PlayerBase.ForceDisconnect(disconnectRequest.GetId())
	return &pb.DisconnectResponse{}, nil
}

// Online returns true if a player is online
func (hive *Hive) Online(ctx context.Context, onlineRequest *pb.OnlineRequest) (*pb.OnlineResponse, error) {
	err := hive.PlayerBase.Online(onlineRequest.XId)
	status := int32(0)
	if err == nil {
		status = 1
	}
	return &pb.OnlineResponse{Status: status}, err
}

// Redeem returns the _id pertaining to a player
func (hive *Hive) Redeem(ctx context.Context, redeemRequest *pb.RedeemRequest) (*pb.RedeemResponse, error) {
	id, err := hive.PlayerBase.GetId(redeemRequest.GetJwt())
	return &pb.RedeemResponse{XId: id}, err
}
