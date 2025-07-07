package pkg

import (
	"log"
	"time"
	"user-service/internal/conf"

	"github.com/sony/sonyflake"
)

type IDGenerator struct {
	sf *sonyflake.Sonyflake
}

// NewIDGenerator 创建一个带机器ID的Sonyflake实例
func NewIDGenerator(cfg *conf.IDGen) *IDGenerator {
	startTime, err := time.Parse(time.RFC3339, cfg.StartTime)
	if err != nil {
		log.Fatalf("invalid start time in config: %v", err)
	}
	settings := sonyflake.Settings{
		StartTime: startTime,
		MachineID: func() (uint16, error) {
			return uint16(cfg.MachineId), nil
		},
	}
	sf := sonyflake.NewSonyflake(settings)
	if sf == nil {
		log.Fatalf("sonyflake not created")
	}
	return &IDGenerator{sf: sf}
}

func (g *IDGenerator) Generate() int64 {
	id, err := g.sf.NextID()
	if err != nil {
		log.Fatalf("failed to generate ID: %v", err)
	}
	return int64(id)
}
