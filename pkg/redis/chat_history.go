package redis

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	// DefaultLatestK is the default number of latest conversation turns to retrieve
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
	SessionID            string `json:"session_id"`
	LatestK              *int   `json:"latest_k,omitempty"`
	IncludeSystemMessage bool   `json:"include_system_message"`
}

// ChatHistoryReadOutput is a wrapper struct for the messages associated with a session ID
type ChatHistoryRetrieveOutput struct {
	Messages []*Message `json:"messages"`
	Status   bool       `json:"status"`
}

// WriteSystemMessage writes system message for a given session ID
func WriteSystemMessage(client *goredis.Client, sessionID string, message MessageWithTime) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Store in a hash with a unique SessionID
	return client.HSet(context.Background(), "system_messages", sessionID, messageJSON).Err()
}

func WriteNonSystemMessage(client *goredis.Client, sessionID string, message MessageWithTime) error {
	// Marshal the MessageWithTime struct to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Index by Timestamp: Add to the Sorted Set
	return client.ZAdd(context.Background(), sessionID+":timestamps", goredis.Z{
		Score:  float64(message.Timestamp),
		Member: string(messageJSON),
	}).Err()
}

// RetrieveSystemMessage gets system message based on a given session ID
func RetrieveSystemMessage(client *goredis.Client, sessionID string) (bool, *MessageWithTime, error) {
	serializedMessage, err := client.HGet(context.Background(), "system_messages", sessionID).Result()

	// Check if the messageID does not exist
	if err == goredis.Nil {
		// Handle the case where the message does not exist
		return false, nil, nil
	} else if err != nil {
		// Handle other types of errors
		return false, nil, err
	}

	var message MessageWithTime
	if err := json.Unmarshal([]byte(serializedMessage), &message); err != nil {
		return false, nil, err
	}

	return true, &message, nil
}

func WriteMessage(client *goredis.Client, input ChatMessageWriteInput) ChatMessageWriteOutput {
	// Current time
	currTime := time.Now().Unix()

	// Create a MessageWithTime struct with the provided input and timestamp
	messageWithTime := MessageWithTime{
		Message: Message{
			Role:     input.Role,
			Content:  input.Content,
			Metadata: input.Metadata,
		},
		Timestamp: currTime,
	}

	// Treat system message differently
	if input.Role == "system" {
		err := WriteSystemMessage(client, input.SessionID, messageWithTime)
		if err != nil {
			return ChatMessageWriteOutput{Status: false}
		} else {
			return ChatMessageWriteOutput{Status: true}
		}
	}

	err := WriteNonSystemMessage(client, input.SessionID, messageWithTime)
	if err != nil {
		return ChatMessageWriteOutput{Status: false}
	} else {
		return ChatMessageWriteOutput{Status: true}
	}
}

// RetrieveSessionMessages retrieves the latest K conversation turns from the Redis list for the given session ID
func RetrieveSessionMessages(client *goredis.Client, input ChatHistoryRetrieveInput) ChatHistoryRetrieveOutput {
	if input.LatestK == nil || *input.LatestK <= 0 {
		input.LatestK = &DefaultLatestK
	}
	key := input.SessionID

	messagesWithTime := []MessageWithTime{}
	messages := []*Message{}
	ctx := context.Background()

	// Retrieve the latest K conversation turns associated with the session ID by descending timestamp order
	messagesNum := *input.LatestK * 2
	timestampMessages, err := client.ZRevRange(ctx, key+":timestamps", 0, int64(messagesNum-1)).Result()
	if err != nil {
		return ChatHistoryRetrieveOutput{
			Messages: messages,
			Status:   false,
		}
	}

	// Iterate through the members and deserialize them into MessageWithTime
	for _, member := range timestampMessages {
		var messageWithTime MessageWithTime
		if err := json.Unmarshal([]byte(member), &messageWithTime); err != nil {
			return ChatHistoryRetrieveOutput{
				Messages: messages,
				Status:   false,
			}
		}
		messagesWithTime = append(messagesWithTime, messageWithTime)
	}

	// Sort the messages by timestamp in ascending order (earliest first)
	sort.SliceStable(messagesWithTime, func(i, j int) bool {
		return messagesWithTime[i].Timestamp < messagesWithTime[j].Timestamp
	})

	// Add System message if exist
	if input.IncludeSystemMessage {
		exist, sysMessage, err := RetrieveSystemMessage(client, input.SessionID)
		if err != nil {
			return ChatHistoryRetrieveOutput{
				Messages: messages,
				Status:   false,
			}
		}
		if exist {
			messages = append(messages, &Message{
				Role:     sysMessage.Role,
				Content:  sysMessage.Content,
				Metadata: sysMessage.Metadata,
			})
		}
	}

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
