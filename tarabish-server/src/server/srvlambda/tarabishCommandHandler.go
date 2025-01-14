package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go-tarabish/src/game-logic/deck"
	"go-tarabish/src/game-logic/player"
	"go-tarabish/src/game-logic/tarabish"
	server "go-tarabish/src/server/messages"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

// holds the api gateway for the entire lifespan of the lambda function
var apigateway *apigatewaymanagementapi.ApiGatewayManagementApi

func buildApigateway(event events.APIGatewayWebsocketProxyRequest) *apigatewaymanagementapi.ApiGatewayManagementApi {
	if apigateway == nil {
		sess, err := session.NewSession()
		if err != nil {
			log.Fatalln("Unable to create AWS session", err.Error())
		}
		dname := event.RequestContext.DomainName
		stage := event.RequestContext.Stage
		endpoint := fmt.Sprintf("https://%v/%v", dname, stage)
		apigateway = apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(endpoint))
	}
	return apigateway
}

func handleCommand(ctx context.Context, event events.APIGatewayWebsocketProxyRequest,
	connectionStore connectionStorer, playerStore tarabish.PlayerWriter, gameStore tarabish.GameReadWriter) error {

	tarabish := tarabish.New(playerStore, gameStore)
	adjustPlayers(ctx, tarabish)
	setGamesStatus(tarabish)

	buildApigateway(event)
	connectionID := event.RequestContext.ConnectionID

	// convert data received to MessageFromPlayer struct
	var msg server.MessageFromPlayer
	if err := json.Unmarshal([]byte(event.Body), &msg); err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Message received", msg)

	gameName := msg.GameName
	playerName := msg.PlayerName

	switch msg.ID {
	case "playerEntersOsteria":
		err := connectionStore.AddPlayerToConnectionID(ctx, connectionID, playerName)
		if err != nil {
			log.Fatalf("Player %v could not be added to its connection", playerName)
		}
		handViewForPlayers, alreadyIn := tarabish.PlayerEnters(playerName)
		if alreadyIn {
			// Player is already in the osteria
			err := connectionStore.MarkConnectionIDDisconnected(ctx, connectionID)
			if err != nil {
				panic(err)
			}
			resp := server.NewMessageToOnePlayer(server.PlayerIsAlreadyInOsteria, playerName)
			resp.Error = fmt.Sprintf("Player \"%v\" is already in the Osteria", playerName)
			sendMessage(ctx, resp, &connectionID)
		} else {
			if handViewForPlayers == nil {
				// if there are no handViews to be sent to Players it means that the Player is entering for the fist time in the Osteria
				// or he is re-entering but was not playing any game previously
				respTo := "playerEntersOsteria - no handViews"
				sendPlayers(ctx, tarabish, respTo, connectionStore)
				sendGames(ctx, tarabish, respTo, connectionStore)
			} else {
				// on the contrary if the handViews are defined it means that the Player is re-entering the Osteria
				// and that he was previously playing a game, so we return the handViews to all Players for them
				// to resume the game
				respTo := fmt.Sprintf("playerEntersOsteria \"%v\"", playerName)
				sendGames(ctx, tarabish, respTo, connectionStore)
				sendPlayers(ctx, tarabish, respTo, connectionStore)
				sendPlayerViews(ctx, tarabish, handViewForPlayers, respTo, connectionStore)
			}
		}
	case "newGame":
		_, e := tarabish.NewGame(gameName)
		if e != nil {
			// There is already a game with the same name
			resp := server.NewMessageToOnePlayer(server.GameWithSameNamePresent, playerName)
			resp.Error = fmt.Sprintf("Game \"%v\" with the same name already created", gameName)
			resp.GameName = gameName
			sendMessage(ctx, resp, &connectionID)
		}
		respTo := fmt.Sprintf("newGame \"%v\"", gameName)
		sendGames(ctx, tarabish, respTo, connectionStore)
	case "addPlayerToGame":
		e := tarabish.AddPlayerToGame(playerName, gameName)
		if e != nil {
			resp := server.NewMessageToOnePlayer(server.ErrorAddingPlayerToGameMsgID, playerName)
			resp.Error = e.Error()
			sendMessage(ctx, resp, &connectionID)
		} else {
			respTo := fmt.Sprintf("addPlayerToGame - game \"%v\"", gameName)
			sendGames(ctx, tarabish, respTo, connectionStore)
		}
	case "addObserverToGame":
		playerName := msg.PlayerName
		gameName := msg.GameName
		hv, err := tarabish.AddObserverToGame(playerName, gameName)
		if err != nil {
			resp := server.NewMessageToOnePlayer(server.ErrorAddingObserverToGameMsgID, playerName)
			resp.Error = err.Error()
			sendMessage(ctx, resp, &connectionID)
		} else {
			respTo := fmt.Sprintf("addObserverToGame - game \"%v\"", gameName)
			sendGames(ctx, tarabish, respTo, connectionStore)
			game := tarabish.Games[gameName]
			sendObserverUpdates(ctx, tarabish, hv, respTo, game, connectionStore)
		}
	case "newHand":
		game := tarabish.Games[gameName]
		_, handViewForPlayers, handCreated := tarabish.NewHand(game)
		if handCreated {
			respTo := fmt.Sprintf("newHand - game \"%v\"", gameName)
			fmt.Println("NewHand", gameName, len(handViewForPlayers))
			sendGames(ctx, tarabish, respTo, connectionStore)
			sendPlayerViews(ctx, tarabish, handViewForPlayers, respTo, connectionStore)
			sendObserverUpdates(ctx, tarabish, handViewForPlayers, respTo, game, connectionStore)
		}
	case "playCard":
		handViewForPlayers, finalTableTake, g := tarabish.Play(msg.PlayerName, msg.CardPlayed, msg.CardsTaken)
		// if handViewForPlayers is nil it means something anomalous happened while playing the card and so
		// there is no message sent to clients
		if handViewForPlayers != nil {
			respTo := fmt.Sprintf("playCard \"%v\"", playerName)
			sendCardsPlayedAndTaken(ctx, msg.CardPlayed, msg.CardsTaken, finalTableTake, g, playerName, respTo, connectionStore)
			sendPlayerViews(ctx, tarabish, handViewForPlayers, respTo, connectionStore)
			sendObserverUpdates(ctx, tarabish, handViewForPlayers, respTo, g, connectionStore)
		}
	case "closeGame":
		gameName := msg.GameName
		tarabish.Close(gameName, playerName)
		respTo := fmt.Sprintf("Game \"%v\" closed", gameName)
		sendGames(ctx, tarabish, respTo, connectionStore)
	default:
		panicMessage := fmt.Sprintf("Unexpected messageId %v arrived from player %v\n", msg.ID, playerName)
		log.Fatal(panicMessage)
	}
	return nil
}

func buildMessage(msg interface{}) []byte {
	msgB, e := json.Marshal(msg)
	if e != nil {
		panicMessage := fmt.Sprintf("Marshalling to json of %v failed with error %v\n", msg, e)
		log.Fatalln(panicMessage)
	}
	return msgB
}

func sendPlayers(ctx context.Context, tarabish *tarabish.Tarabish, responseTo string, store connectionStorer) {
	msg := server.NewMessageToAllClients(server.PlayersMsgID)
	msg.Players = tarabish.AllPlayers()
	msg.ResponseTo = responseTo
	broadcast(ctx, msg, store)
}

func sendGames(ctx context.Context, tarabish *tarabish.Tarabish, responseTo string, store connectionStorer) {
	allGames := tarabish.AllGames()
	if len(allGames) > 0 {
		msg := server.NewMessageToAllClients(server.GamesMsgID)
		msg.Games = allGames
		msg.ResponseTo = responseTo
		broadcast(ctx, msg, store)
	}
}

func sendPlayerViews(ctx context.Context, tarabish *tarabish.Tarabish,
	handViewForPlayers map[string]tarabish.HandPlayerView, responseTo string, store connectionStorer) {
	for playerName := range handViewForPlayers {
		hView := handViewForPlayers[playerName]
		msg := server.NewMessageToOnePlayer(server.HandView, playerName)
		msg.ResponseTo = responseTo
		msg.HandPlayerView = hView
		connectionID, err := store.ConnectionIDForPlayer(ctx, playerName)
		if err != nil {
			log.Printf("Connection for player %v not found", playerName)
		}

		if tarabish.Players[playerName].Status == player.PlayerPlaying || tarabish.Players[playerName].Status == player.PlayerLookingAtHandResult {
			sendMessage(ctx, msg, &connectionID)
		}
	}
}

func sendObserverUpdates(ctx context.Context, tarabish *tarabish.Tarabish,
	handViewForPlayers map[string]tarabish.HandPlayerView, responseTo string, game *tarabish.Game, store connectionStorer) {
	for observerName := range game.Observers {
		msg := server.NewMessageToOnePlayer(server.HandView, observerName)
		msg.ResponseTo = responseTo
		msg.AllHandPlayerViews = handViewForPlayers
		connectionID, err := store.ConnectionIDForPlayer(ctx, observerName)
		if err != nil {
			log.Printf("Connection for observer %v not found", observerName)
		}
		sendMessage(ctx, msg, &connectionID)
	}
}

func sendCardsPlayedAndTaken(ctx context.Context, cardPlayed deck.Card, cardsTaken []deck.Card,
	finalTableTake tarabish.FinalTableTake, game *tarabish.Game, playerName string, responseTo string, store connectionStorer) {
	playerObservers := make([]string, 0)
	for p := range game.Players {
		playerObservers = append(playerObservers, p)
	}
	for o := range game.Observers {
		playerObservers = append(playerObservers, o)
	}
	for _, playerObserverName := range playerObservers {
		msg := server.NewMessageToOnePlayer(server.CardsPlayedAndTaken, playerObserverName)
		msg.ResponseTo = responseTo
		msg.CardPlayed = cardPlayed
		msg.CardsTaken = cardsTaken
		msg.FinalTableTake = finalTableTake
		msg.CardPlayedByPlayer = playerName
		msg.FinalTableTake = finalTableTake
		connectionID, err := store.ConnectionIDForPlayer(ctx, playerObserverName)
		if err != nil {
			log.Printf("Connection for player %v not found", playerObserverName)
		}
		log.Printf(">>>>>>>>>>>>>>>><<<<<<< Observer %v  with connectionID %v", playerObserverName, connectionID)
		// ATTENTION PLEASE
		sendMessage(ctx, msg, &connectionID)
	}
}

func sendMessage(ctx context.Context, msg server.MessageToOnePlayer, connectionID *string) {

	msgB := buildMessage(msg)

	input := &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: connectionID,
		Data:         msgB,
	}
	_, err := apigateway.PostToConnection(input)
	if err != nil {
		log.Println("ERROR while sending message to a client", err.Error())
	}
}

func broadcast(ctx context.Context, msg server.MessageToAllClients, store connectionStorer) {
	msgB := buildMessage(msg)

	connections, err := store.ActiveConnectionIDs(ctx)
	if err != nil {
		log.Fatalln("Unable to get connections", err.Error())
	}
	for _, conn := range connections {
		input := &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(conn),
			Data:         msgB,
		}

		_, err = apigateway.PostToConnection(input)
		if err != nil {
			log.Println("ERROR while sending message to a client", err.Error())
		}
	}
}

// adjustPlayers adds to tarabish all the players who have already connected
// but who are not yet into any game - these players are not loaded by the GameReadWriter since this
// loads only the players who are playing a game, not those who have just entered the osteria
// Then is also set with status "PlayerPlaying" those players who are actually into a game
func adjustPlayers(ctx context.Context, tarabish *tarabish.Tarabish) {
	connectedPlayers, err := connectionStore.ConnectedPlayers(ctx)
	if err != nil {
		log.Fatalln("Can not read connected players from store")
	}
	for _, p := range connectedPlayers {
		pInGame := tarabish.Players[p]
		if pInGame == nil {
			connectedP := &player.Player{}
			connectedP.Name = p
			connectedP.Status = player.PlayerNotPlaying
			tarabish.Players[p] = connectedP
		} else {
			pInGame.Status = player.PlayerPlaying
		}
	}
}

// setGamesStatus sets the status of the games
func setGamesStatus(tarabish *tarabish.Tarabish) {
	for _, g := range tarabish.Games {
		g.CalculateState()
	}
}
