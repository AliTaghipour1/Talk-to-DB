package ai

import "log"

type AIModule struct {
	avalaiClient *avalaiClient
}

func NewAIModule(apikey string) *AIModule {
	return &AIModule{avalaiClient: newAvalaiClient(apikey)}
}

func (m *AIModule) GetQuery(databaseContext string, nlq string) string {
	log.Println("NLQ", nlq)
	ask, err := m.avalaiClient.ask(databaseContext, nlq)
	if err != nil {
		log.Println("failed to ask:", err)
		return ""
	}

	log.Println("ask result:", ask)
	return ask
}
