package message

import (
	"github.com/openai/openai-go/v3"
)

type MessageData struct {
	Age    int
	Gender string
	Year   int
}
type AiMessage []openai.ChatCompletionMessageParamUnion

func (m *AiMessage) AddUserAssistantMap(mapToAdd map[string]string) {
	for k, v := range mapToAdd {
		(*m) = append(*m, openai.UserMessage(k), openai.AssistantMessage(v))
	}
}
func (m *AiMessage) AddUser(userMsg string) {
	(*m) = append(*m, openai.UserMessage(userMsg))
}
