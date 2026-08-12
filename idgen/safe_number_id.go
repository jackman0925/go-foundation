package idgen

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	defaultSafeNumberIDSequenceDigits = 3
	maxSafeNumberIDSequenceDigits     = 4
	maxJSSafeInteger                  = int64(9007199254740991)
	safeNumberIDPrefixLayout          = "06002150405"
)

// SafeNumberIDOptions 配置前端安全数字 ID 生成器。
type SafeNumberIDOptions struct {
	// SequenceDigits 控制同一秒内的序列位数，默认 3，最大 4。
	SequenceDigits int
	// Location 控制 ID 中日期时间前缀使用的时区，默认使用本地时区。
	Location *time.Location
}

// SafeNumberID 生成 JS Number 可安全表示的本地数字 ID。
type SafeNumberID struct {
	mu             sync.Mutex
	sequenceDigits int
	maxSequence    int64
	location       *time.Location
	lastSecond     string
	sequence       int64
	now            func() time.Time
}

// NewSafeNumberID 创建前端安全数字 ID 生成器。
func NewSafeNumberID(options SafeNumberIDOptions) (*SafeNumberID, error) {
	if options.SequenceDigits == 0 {
		options.SequenceDigits = defaultSafeNumberIDSequenceDigits
	}
	if options.SequenceDigits < 1 || options.SequenceDigits > maxSafeNumberIDSequenceDigits {
		return nil, fmt.Errorf("sequence digits must be between 1 and %d", maxSafeNumberIDSequenceDigits)
	}
	if options.Location == nil {
		options.Location = time.Local
	}

	maxSequence := int64(1)
	for i := 0; i < options.SequenceDigits; i++ {
		maxSequence *= 10
	}

	return &SafeNumberID{
		sequenceDigits: options.SequenceDigits,
		maxSequence:    maxSequence - 1,
		location:       options.Location,
		now:            time.Now,
	}, nil
}

// Next 返回下一个 JS Number 可安全表示的 int64 ID。
func (s *SafeNumberID) Next() (int64, error) {
	id, err := s.NextString()
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse safe number id: %w", err)
	}
	if number > maxJSSafeInteger {
		return 0, fmt.Errorf("safe number id exceeds JS safe integer: %d", number)
	}
	return number, nil
}

// NextString 返回下一个数字字符串形式的 ID。
func (s *SafeNumberID) NextString() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentSecond := s.now().In(s.location).Format(safeNumberIDPrefixLayout)
	if currentSecond == s.lastSecond {
		if s.sequence >= s.maxSequence {
			return "", fmt.Errorf("safe number id sequence exhausted in second %s", currentSecond)
		}
		s.sequence++
	} else {
		s.lastSecond = currentSecond
		s.sequence = 0
	}

	return fmt.Sprintf("%s%0*d", currentSecond, s.sequenceDigits, s.sequence), nil
}
