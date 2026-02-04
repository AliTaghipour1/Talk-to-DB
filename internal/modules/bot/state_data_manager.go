package bot

import (
	"fmt"
	"sync"
)

type stateDataManager struct {
	data sync.Map
}

func newStateDataManager() *stateDataManager {
	return &stateDataManager{}
}

type DescriptionData struct {
	Table  string
	Column *string
}

const (
	userDescriptionKey = "description-data-%d"
)

func getDescriptionKey(userID int64) string {
	return fmt.Sprintf(userDescriptionKey, userID)
}

func (s *stateDataManager) GetDescriptionData(userID int64) (DescriptionData, bool) {
	value, ok := s.data.Load(getDescriptionKey(userID))
	if !ok {
		return DescriptionData{}, false
	}

	descriptionData, ok := value.(*DescriptionData)
	if !ok {
		return DescriptionData{}, false
	}
	return *descriptionData, true
}

func (s *stateDataManager) AddDescriptionColumnData(columnName string, userID int64) error {
	value, ok := s.data.Load(getDescriptionKey(userID))
	if !ok {
		return fmt.Errorf("description data does not exist")
	}

	descriptionData, ok := value.(*DescriptionData)
	if !ok {
		return fmt.Errorf("description data is not of type DescriptionData")
	}
	descriptionData.Column = &columnName
	s.data.Store(getDescriptionKey(userID), descriptionData)
	return nil
}

func (s *stateDataManager) AddDescriptionTableData(tableName string, userID int64) error {
	s.data.Store(getDescriptionKey(userID), &DescriptionData{Table: tableName})
	return nil
}

func (s *stateDataManager) EmptyUserStateData(userID int64) {
	s.data.Delete(getDescriptionKey(userID))
}
