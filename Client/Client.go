package main

import (
	"MultiplaterDemo/SharedData"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/codecat/go-enet"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"log"
	"path"
)

//go:embed assets/*
var assets embed.FS

type gameClient struct {
	server           enet.Peer
	serverConn       enet.Host
	serverGameState  SharedData.MPgame
	playerPict       *ebiten.Image
	goldPict         *ebiten.Image
	desiredDirection int
}

func (client *gameClient) Update() error {
	client.desiredDirection = getPlayerInput()
	client.server.SendString(fmt.Sprintf("%d", client.desiredDirection), 0, enet.PacketFlagReliable)
	// Wait until the next event; param is timeout time
	ev := client.serverConn.Service(1000)
	switch ev.GetType() {
	case enet.EventConnect: // We connected to the server
		log.Println("Connected to the server! event", ev)

	case enet.EventDisconnect: // We disconnected from the server
		log.Println("Lost connection to the server!")

	case enet.EventReceive: // The server sent us data
		packet := ev.GetPacket()
		log.Println("got Raw Data from server:", packet.GetData())
		json.Unmarshal(packet.GetData(), &client.serverGameState)
		log.Println("got info from server, updated gamestate to:", client.serverGameState)
		packet.Destroy()
	}
	return nil
}

func getPlayerInput() int {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		return SharedData.LEFT
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
		return SharedData.RIGHT
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
		return SharedData.DOWN
	} else if ebiten.IsKeyPressed(ebiten.KeyUp) {
		return SharedData.UP
	} else {
		return SharedData.STILL
	}
}

func (client *gameClient) Draw(screen *ebiten.Image) {
	drawOps := ebiten.DrawImageOptions{}
	for _, treasure := range client.serverGameState.Gold {
		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(float64(treasure.Xloc), float64(treasure.Yloc))
		screen.DrawImage(client.goldPict, &drawOps)
	}
	for _, player := range client.serverGameState.Players {
		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(float64(player.Xloc), float64(player.Yloc))
		screen.DrawImage(client.playerPict, &drawOps)
	}
}

func (client *gameClient) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowSize(SharedData.WINDOW_WIDTH, SharedData.WINDOW_WIDTH)
	ebiten.SetWindowTitle("Multi Player Demo")

	// Initialize enet
	enet.Initialize()

	// Create a client host
	client, err := enet.NewHost(nil, 1, 1, 0, 0)
	if err != nil {
		log.Fatal("Couldn't create host: %s", err)
		return
	}

	// Connect the client host to the server
	peer, err := client.Connect(enet.NewAddress("127.0.0.1", 8095), 1, 0)
	if err != nil {
		log.Fatal("Couldn't connect: %s", err)
		return
	}
	game := gameClient{
		server:     peer,
		serverConn: client,
		playerPict: LoadEmbeddedImage("", "goblin.png"),
		goldPict:   LoadEmbeddedImage("", "coins.png"),
	}
	defer client.Destroy() // Destroy the host when we're done with it
	// Uninitialize enet
	defer enet.Deinitialize()
	err = ebiten.RunGame(&game)
	if err != nil {
		fmt.Println("Error running game", err)
	}
}

func LoadEmbeddedImage(folderName string, imageName string) *ebiten.Image {
	embeddedFile, err := assets.Open(path.Join("assets", folderName, imageName))
	if err != nil {
		log.Fatal("failed to load embedded image ", imageName, err)
	}
	ebitenImage, _, err := ebitenutil.NewImageFromReader(embeddedFile)
	if err != nil {
		fmt.Println("Error loading tile image:", imageName, err)
	}
	return ebitenImage
}
