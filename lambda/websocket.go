package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

type WebSocketMessage struct {
	Type    string      `json:"type"`
	GameID  int         `json:"game_id,omitempty"`
	UserID  int         `json:"user_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

var apiGatewayClient *apigatewaymanagementapi.ApiGatewayManagementApi

func init() {
	sess := session.Must(session.NewSession())
	apiGatewayClient = apigatewaymanagementapi.New(sess)
}

func handleConnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("WebSocket connection established: %s", request.RequestContext.ConnectionID)
	
	// Store connection in DynamoDB or other persistent storage
	// This is a simplified implementation
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Connected",
	}, nil
}

func handleDisconnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("WebSocket connection closed: %s", request.RequestContext.ConnectionID)
	
	// Remove connection from storage
	// This is a simplified implementation
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Disconnected",
	}, nil
}

func handleDefault(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("WebSocket message received from %s: %s", request.RequestContext.ConnectionID, request.Body)
	
	var message WebSocketMessage
	if err := json.Unmarshal([]byte(request.Body), &message); err != nil {
		log.Printf("Error parsing message: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid message format",
		}, nil
	}
	
	// Handle different message types
	switch message.Type {
	case "chat":
		return handleChatMessage(ctx, request, message)
	case "move":
		return handleGameMove(ctx, request, message)
	case "join_game":
		return handleJoinGame(ctx, request, message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Message processed",
	}, nil
}

func handleChatMessage(ctx context.Context, request events.APIGatewayWebsocketProxyRequest, message WebSocketMessage) (events.APIGatewayProxyResponse, error) {
	// Broadcast chat message to all connections in the same game
	// This would require querying the database for game participants
	// and sending the message to their connections
	
	log.Printf("Chat message in game %d: %s", message.GameID, message.Message)
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Chat message sent",
	}, nil
}

func handleGameMove(ctx context.Context, request events.APIGatewayWebsocketProxyRequest, message WebSocketMessage) (events.APIGatewayProxyResponse, error) {
	// Process game move and broadcast to opponent
	log.Printf("Game move in game %d: %+v", message.GameID, message.Data)
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Move processed",
	}, nil
}

func handleJoinGame(ctx context.Context, request events.APIGatewayWebsocketProxyRequest, message WebSocketMessage) (events.APIGatewayProxyResponse, error) {
	// Associate connection with game
	log.Printf("Player joining game %d", message.GameID)
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Joined game",
	}, nil
}

func sendMessageToConnection(connectionID string, message interface{}) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	_, err = apiGatewayClient.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         messageBytes,
	})
	
	return err
}

func Handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.RequestContext.RouteKey {
	case "$connect":
		return handleConnect(ctx, request)
	case "$disconnect":
		return handleDisconnect(ctx, request)
	case "$default":
		return handleDefault(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Unknown route",
		}, nil
	}
}

func main() {
	lambda.Start(Handler)
}
