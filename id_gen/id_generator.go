package id_gen

import (
	"github.com/drip-in/eden_lib/utils"
)

type IdGenerator struct {
	idgen *utils.SnowFlakeIdGenerator
}

func NewIdGenerator() *IdGenerator {
	generator, err := utils.NewIDGenerator().SetWorkerId(100).Init()
	if err != nil {
		panic(err)
	}
	gen := &IdGenerator{idgen: generator}

	return gen
}
