package main

import (
	"MultiplaterDemo/SharedData"
	"encoding/json"
	"github.com/co0p/tankism/lib/collision"
	"github.com/codecat/go-enet"
	"log"
	"math/rand"
	"strconv"
)

func main() {
	// Initialize enet
	enet.Initialize()

	// Create a host listening on  port 8095
	host, err := enet.NewHost(enet.NewListenAddress(8095), 32, 1, 0, 0)
	if err != nil {
		log.Fatal("Couldn't create host: %s", err.Error())
		return
	}
	mainGame := SharedData.MPgame{
		Players: make([]*SharedData.Player, 0, 20), // no more than 20 concurrent players
		Gold:    makeGold(),
	}
	RunServer(host, mainGame)

}

func makeGold() []*SharedData.GoldPile {
	allTreasures := make([]*SharedData.GoldPile, 0, 15)
	for i := 0; i < 10; i++ {
		allTreasures = append(allTreasures, &SharedData.GoldPile{
			Xloc: rand.Intn(SharedData.WINDOW_WIDTH),
			Yloc: rand.Intn(SharedData.WINDOW_HEIGHT),
		})
	}
	return allTreasures
}

func RunServer(host enet.Host, game SharedData.MPgame) {
	for { //For ever
		// Wait until the next event, 1000 is timeout
		ev := host.Service(1000)

		// Do nothing if we didn't get any event
		if ev.GetType() == enet.EventNone {
			continue
		}
		switch ev.GetType() {
		case enet.EventConnect: // A new peer has connected
			log.Println("New peer connected: %s", ev.GetPeer().GetAddress())
			newPlayer := SharedData.Player{
				PlayerID: ev.GetPeer().GetAddress().String(),
				Xloc:     rand.Intn(SharedData.WINDOW_WIDTH - SharedData.PLAYER_WIDTH),
				Yloc:     rand.Intn(SharedData.WINDOW_HEIGHT - SharedData.PLAYER_HEIGHT),
				Score:    0,
			}
			game.Players = append(game.Players, &newPlayer)
		case enet.EventDisconnect: // A connected peer has disconnected
			log.Println("Peer disconnected: %s", ev.GetPeer().GetAddress())
			//remove disconnected player
			var loc int
			disconnectPlayerID := ev.GetPeer().GetAddress().String()
			for i, player := range game.Players {
				if disconnectPlayerID == player.PlayerID {
					loc = i
					break
				}
			}
			game.Players = append(game.Players[:loc], game.Players[loc+1:]...) //the ... will 'unpack' the slice
		case enet.EventReceive: // A peer sent us some data

			processPlayerMove(&game, ev)
			result, err := json.Marshal(&game)
			//log.Println("About to send this json to client", result)
			if err != nil {
				log.Fatal("Big error turning game into json: ", err)
			}
			ev.GetPeer().SendString(string(result), ev.GetChannelID(), enet.PacketFlagReliable)

		}
	}
}

func processPlayerMove(game *SharedData.MPgame, ev enet.Event) {
	// Get the packet
	packet := ev.GetPacket()

	// We must destroy the packet when we're done with it
	defer packet.Destroy()

	// Get the bytes in the packet
	directionAsString := string(packet.GetData())
	direction, err := strconv.Atoi(directionAsString)
	if err != nil {
		log.Fatal("Client side trickery! not sending direction", err)
	}
	playerID := ev.GetPeer().GetAddress().String()
	//find the right player
	for loc, _ := range game.Players {
		if playerID == game.Players[loc].PlayerID {
			if direction == SharedData.UP {
				game.Players[loc].Yloc -= SharedData.TRAVEL_SPEED
			} else if direction == SharedData.DOWN {
				game.Players[loc].Yloc += SharedData.TRAVEL_SPEED
			} else if direction == SharedData.LEFT {
				game.Players[loc].Xloc -= SharedData.TRAVEL_SPEED
			} else if direction == SharedData.RIGHT {
				game.Players[loc].Xloc += SharedData.TRAVEL_SPEED
			}
			//now check to see if the player overlaps any of the gold
			playerBounds := collision.BoundingBox{
				X:      float64(game.Players[loc].Xloc),
				Y:      float64(game.Players[loc].Yloc),
				Width:  SharedData.PLAYER_WIDTH,
				Height: SharedData.PLAYER_HEIGHT,
			}
			//only one can be collected per update
			for treasureNum, treasure := range game.Gold {
				goldBounds := collision.BoundingBox{
					X:      float64(treasure.Xloc),
					Y:      float64(treasure.Yloc),
					Width:  SharedData.GOLD_WIDTH,
					Height: SharedData.GOLD_HEIGHT,
				}
				if collision.AABBCollision(playerBounds, goldBounds) {
					game.Players[loc].Score += 1
					game.Gold = append(game.Gold[:treasureNum], game.Gold[treasureNum+1:]...)
					break
				}
			}
			//	log.Println("processing move for playerID:", game.Players[loc].PlayerID)
		}
	}

}
