package contractinterface

import (
	"gitlab.lrz.de/orderless/orderlessfile/internal/benchmark/benchmarkutils"
	"gitlab.lrz.de/orderless/orderlessfile/protos/goprotos"
)

type ContractInterface interface {
	Invoke(ShimInterface, *protos.ProposalRequest) (*protos.ProposalResponse, error)
}

type BaseContractOptions struct {
	Bconfig               *protos.BenchmarkConfig
	BenchmarkFunction     BenchmarkFunction
	BenchmarkUtils        *benchmarkutils.BenchmarkUtils
	CurrentClientPseudoId int
	TotalTransactions     int
	SingleFunctionCounter *SingleFunctionCounter
}
