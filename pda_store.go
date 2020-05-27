package main

import (
	//"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// Store ... Stores PDAs in a key value store
type Store struct {
	c         *cache.Cache
	c_replica *cache.Cache
}

func (s *Store) init() {
	s.c = cache.New(cache.NoExpiration, 10*time.Minute)
	s.c_replica = cache.New(cache.NoExpiration, 10*time.Minute)
}

func (s *Store) add(pdaid string, pdaaddress *PushDownAutomata) {
	s.c.Set(pdaid, pdaaddress, cache.NoExpiration)
}

func (s *Store) add_replica(gid string, rp interface{}) {

	s.c_replica.Set(gid, rp, cache.NoExpiration)

}

func (s *Store) setWorkingPda(gid string, pdaid string) {

	s.c_replica.Set(gid, pdaid, cache.NoExpiration)

}

func (s *Store) getWorkingPda(gid string, pdaid string) string {

	v, _ := s.c_replica.Get(gid)
	value, _ := v.(string)
	return value

}


func (s *Store) remove(pdaid string) {
	s.c.Delete(pdaid)
}

func (s *Store) removeReplica(gid string) {
	s.c_replica.Delete(gid)
}

func (s *Store) getAll() map[string]*PushDownAutomata {
	allKvs := s.c.Items()
	n := make(map[string]*PushDownAutomata, len(allKvs))
	for k, v := range allKvs {
		value, _ := v.Object.(*PushDownAutomata)
		n[k] = value
	}
	return n
}

func (s *Store) getAllReplicas() []string {
	allKvs := s.c_replica.Items()
	var n []string
	var i int = 0
	for k, _ := range allKvs {
		n = append(n, k)
		i = i + 1
	}
	return n
}

func (s *Store) removeAll() {
	s.c.Flush()
}

func (s *Store) get(pdaid string) *PushDownAutomata {
	v, _ := s.c.Get(pdaid)
	value, _ := v.(*PushDownAutomata)
	return value
}

func (s *Store) getReplicaMembers(gid string) []string {
	v, _ := s.c_replica.Get(gid)
	value, _ := v.(replica_json)
	return value.Group_members
}

func (s *Store) getReplica(gid string) replica_json {
	v, _ := s.c_replica.Get(gid)
	value, _ := v.(replica_json)
	return value
}
