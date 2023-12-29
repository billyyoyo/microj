package coor

import (
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/server/coordinate"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"sync"
)

const (
	LEADER_ADDRESS_KEY = "leader-address"
)

type CoordinatorHandler struct {
	sync.RWMutex
	cache    map[string]interface{}
	Raft     *raft.Raft
	raftConf coordinate.ConfigRaft
}

func (c *CoordinatorHandler) InitFinished(cr coordinate.ConfigRaft, r *raft.Raft) {
	c.cache = make(map[string]interface{})
	c.Raft = r
	c.raftConf = cr
}

func (c *CoordinatorHandler) LeaderChanged(isLeader bool) {
	if isLeader {
		cmd := Command{
			Action: "set",
			Payload: map[string]interface{}{
				LEADER_ADDRESS_KEY: fmt.Sprintf("%s:%d", c.raftConf.Address, app.Port()),
			},
		}
		bs, _ := json.Marshal(cmd)
		f := c.Raft.Apply(bs, 0)
		if f.Error() == nil {
			logger.Err(errors.Wrap(f.Error(), "can not apply leader host address"))
		}
	}
}

func (c *CoordinatorHandler) CommandApplied(data []byte) interface{} {
	logger.Info("raft app log apply")
	cmd := Command{}
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		return err
	}
	switch cmd.Action {
	case "set":
		c.Set(cmd.Payload)
	case "del":
		c.Del(cmd.Payload["key"].(string))
	}
	return nil
}

func (c *CoordinatorHandler) GetApplicationData() (data []byte, err error) {
	logger.Info("get all raft app data")
	c.RLock()
	defer c.RUnlock()
	data, err = json.Marshal(c.cache)
	if err != nil {
		logger.Error("Serialize application data failed", err)
	}
	return
}

func (c *CoordinatorHandler) SnapshotRestored(data []byte) (err error) {
	logger.Info("restore raft app data from snapshot")
	c.Lock()
	defer c.Unlock()
	d := make(map[string]interface{})
	err = json.Unmarshal(data, &d)
	if err != nil {
		logger.Error("restore snapshot unserialize failed", err)
		return err
	}
	c.cache = d
	return nil
}

func (c *CoordinatorHandler) Get(key string) interface{} {
	logger.Info("get raft app data by key")
	c.RLock()
	defer c.RUnlock()
	return c.cache[key]
}

func (c *CoordinatorHandler) GetLeaderApiAddress() (addr string, err error) {
	d := c.Get(LEADER_ADDRESS_KEY)
	if d == nil {
		err = errors.New("no leader address in data cache")
		return
	}
	addr = d.(string)
	if addr == "" {
		err = errors.New("no leader address in data cache")
	}
	return
}

func (c *CoordinatorHandler) Set(data map[string]interface{}) {
	logger.Info("batch set raft app data")
	c.Lock()
	defer c.Unlock()
	for k, v := range data {
		c.cache[k] = v
	}
}

func (c *CoordinatorHandler) Del(key string) {
	logger.Info("del raft app data by key")
	c.Lock()
	defer c.Unlock()
	delete(c.cache, key)
}

type Command struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}
