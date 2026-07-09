package idgen

import (
	"fmt"
	"sync"
	"time"
)

const (
	maxMachineID = 1023
	sequenceMask = 4095
)

var defaultEpoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// SnowflakeOptions 配置 Snowflake 生成器。
type SnowflakeOptions struct {
	MachineID int64
	Epoch     time.Time
}

// Snowflake 生成可排序的唯一 int64 ID。
type Snowflake struct {
	mu            sync.Mutex
	epochMillis   int64
	machineID     int64
	sequence      int64
	lastTimestamp int64
	now           func() int64
}

// NewSnowflake 使用显式机器配置创建 Snowflake 生成器。
func NewSnowflake(options SnowflakeOptions) (*Snowflake, error) {
	if options.MachineID < 0 || options.MachineID > maxMachineID {
		return nil, fmt.Errorf("machine id must be between 0 and %d", maxMachineID)
	}
	if options.Epoch.IsZero() {
		options.Epoch = defaultEpoch
	}

	return &Snowflake{
		epochMillis:   options.Epoch.UnixMilli(),
		machineID:     options.MachineID,
		lastTimestamp: -1,
		now: func() int64 {
			return time.Now().UnixMilli()
		},
	}, nil
}

// NextID 返回下一个唯一 ID。
func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := s.now()
	if timestamp < s.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards from %d to %d", s.lastTimestamp, timestamp)
	}

	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			timestamp = s.waitNextMillis(s.lastTimestamp)
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = timestamp
	return ((timestamp - s.epochMillis) << 22) | (s.machineID << 12) | s.sequence, nil
}

func (s *Snowflake) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := s.now()
	for timestamp <= lastTimestamp {
		timestamp = s.now()
	}
	return timestamp
}
