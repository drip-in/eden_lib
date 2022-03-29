package id_gen

var (
	IdgeneratorImpl IIdGenerator
)

func InitIdGeneratorImpl(impl IIdGenerator) {
	IdgeneratorImpl = impl
}

type IIdGenerator interface {
	Get() (int64, error)
}
