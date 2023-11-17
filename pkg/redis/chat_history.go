package redis

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	// DefaultLatestK is the default number of latest messages to retrieve
	DefaultLatestK = 10
)

type Message struct {
	Role     string                  `json:"role"`
	Content  string                  `json:"content"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

type MessageWithTime struct {
	Message
	Timestamp int64 `json:"timestamp"`
}

type ChatMessageWriteInput struct {
	SessionID string `json:"session_id"`
	Message
}

type ChatMessageWriteOutput struct {
	Status bool `json:"status"`
}

type ChatHistoryRetrieveInput struct {
	SessionID string `json:"session_id"`
	LatestK   *int   `json:"latest_k,omitempty"`
}

// ChatHistoryReadOutput is a wrapper struct for the messages associated with a session ID
type ChatHistoryRetrieveOutput struct {
	Messages []*Message `json:"messages"`
	Status   bool       `json:"status"`
}

func WriteMessage(client *goredis.Client, input ChatMessageWriteInput) ChatMessageWriteOutput {
	// Current time
	currTime := time.Now().Unix()
	key := input.SessionID

	// Create a MessageWithTime struct with the provided input and timestamp
	messageWithTime := MessageWithTime{
		Message: Message{
			Role:     input.Role,
			Content:  input.Content,
			Metadata: input.Metadata,
		},
		Timestamp: currTime,
	}

	// Marshal the MessageWithTime struct to JSON
	messageJSON, err := json.Marshal(messageWithTime)
	if err != nil {
		return ChatMessageWriteOutput{Status: false}
	}

	// Append chat message to the Redis list
	err = client.RPush(context.Background(), key, messageJSON).Err()
	if err != nil {
		return ChatMessageWriteOutput{Status: false}
	}

	return ChatMessageWriteOutput{Status: true}
}

// RetrieveSessionMessages retrieves the latest K messages from the Redis list for the given session ID
func RetrieveSessionMessages(client *goredis.Client, input ChatHistoryRetrieveInput) ChatHistoryRetrieveOutput {
	if input.LatestK == nil || *input.LatestK <= 0 {
		input.LatestK = &DefaultLatestK
	}

	messagesWithTime := []MessageWithTime{}
	messages := []*Message{}

	// Determine the start and stop indexes for retrieving the latest k messages
	startIndex := int64(0)
	stopIndex := int64(*input.LatestK - 1) // The stop index is k-1 to fetch the latest k messages

	// Retrieve the latest k messages associated with the sessionID
	messageWithTimeJSONs, err := client.LRange(context.Background(), input.SessionID, startIndex, stopIndex).Result()
	if err != nil {
		// Handle the error, e.g., log it or return an error response
		return ChatHistoryRetrieveOutput{
			Messages: messages,
			Status:   false,
		}
	}

	// Unmarshal retrieved JSON messages into MessageWithTime structs
	for _, m := range messageWithTimeJSONs {
		var messageWithTime MessageWithTime
		if err := json.Unmarshal([]byte(m), &messageWithTime); err != nil {
			// Handle the error, e.g., log it or skip the invalid message
			continue
		}
		messagesWithTime = append(messagesWithTime, messageWithTime)
	}

	// Sort the messages by timestamp in ascending order (earliest first)
	sort.SliceStable(messagesWithTime, func(i, j int) bool {
		return messagesWithTime[i].Timestamp < messagesWithTime[j].Timestamp
	})

	// Convert the MessageWithTime structs to Message structs
	for _, m := range messagesWithTime {
		messages = append(messages, &Message{
			Role:     m.Role,
			Content:  m.Content,
			Metadata: m.Metadata,
		})
	}
	return ChatHistoryRetrieveOutput{
		Messages: messages,
		Status:   true,
	}
}
