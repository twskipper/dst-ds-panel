package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"dst-ds-panel/internal/model"
)

type Store struct {
	mu       sync.RWMutex
	filePath string
	data     storeData
}

type storeData struct {
	Clusters []model.Cluster `json:"clusters"`
}

func New(filePath string) (*Store, error) {
	s := &Store{filePath: filePath}
	if err := s.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		s.data = storeData{Clusters: []model.Cluster{}}
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.data)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Store) ListClusters() []model.Cluster {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Cluster, len(s.data.Clusters))
	copy(result, s.data.Clusters)
	return result
}

func (s *Store) GetCluster(id string) (*model.Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.data.Clusters {
		if s.data.Clusters[i].ID == id {
			c := s.data.Clusters[i]
			return &c, nil
		}
	}
	return nil, fmt.Errorf("cluster %s not found", id)
}

func (s *Store) SaveCluster(c model.Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Clusters {
		if s.data.Clusters[i].ID == c.ID {
			s.data.Clusters[i] = c
			return s.save()
		}
	}
	s.data.Clusters = append(s.data.Clusters, c)
	return s.save()
}

func (s *Store) DeleteCluster(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Clusters {
		if s.data.Clusters[i].ID == id {
			s.data.Clusters = append(s.data.Clusters[:i], s.data.Clusters[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("cluster %s not found", id)
}
