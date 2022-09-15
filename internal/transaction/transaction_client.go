package transaction

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"gitlab.lrz.de/orderless/orderlessfile/internal/config"
	"gitlab.lrz.de/orderless/orderlessfile/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlessfile/internal/customcrypto/signer"
	"gitlab.lrz.de/orderless/orderlessfile/internal/logger"
	"gitlab.lrz.de/orderless/orderlessfile/protos/goprotos"
	"sort"
)

type ClientTransaction struct {
	TransactionId     string
	ProposalResponses map[string]*protos.ProposalResponse
	Status            protos.TransactionStatus
	signer            *signer.Signer
}

func NewClientTransaction(signer *signer.Signer) *ClientTransaction {
	return &ClientTransaction{
		TransactionId:     uuid.NewString(),
		Status:            protos.TransactionStatus_RUNNING,
		ProposalResponses: map[string]*protos.ProposalResponse{},
		signer:            signer,
	}
}

func (t *ClientTransaction) MakeProposalRequestBenchmarkExecutor(counter int, baseOptions *contractinterface.BaseContractOptions) *protos.ProposalRequest {
	proposalParams := baseOptions.BenchmarkFunction(&contractinterface.BenchmarkFunctionOptions{
		Counter:                     counter,
		CurrentClientPseudoId:       baseOptions.CurrentClientPseudoId,
		BenchmarkUtils:              baseOptions.BenchmarkUtils,
		TotalTransactions:           baseOptions.TotalTransactions,
		SingleFunctionCounter:       baseOptions.SingleFunctionCounter,
		CrdtOperationPerObjectCount: baseOptions.Bconfig.CrdtOperationPerObjectCount,
		CrdtObjectType:              baseOptions.Bconfig.CrdtObjectType,
		CrdtObjectCount:             baseOptions.Bconfig.CrdtObjectCount,
		NumberOfKeys:                int(baseOptions.Bconfig.NumberOfKeys),
		NumberOfKeysSecond:          int(baseOptions.Bconfig.NumberOfKeysSecond),
	})
	return &protos.ProposalRequest{
		TargetSystem:         baseOptions.Bconfig.TargetSystem,
		ProposalId:           t.TransactionId,
		ClientId:             config.Config.UUID,
		ContractName:         baseOptions.Bconfig.ContractName,
		MethodName:           proposalParams.ContractMethodName,
		MethodParams:         proposalParams.Outputs,
		WriteReadTransaction: proposalParams.WriteReadType,
		FileShardHashedData:  proposalParams.FileShardHashedByte,
	}
}

func (t *ClientTransaction) PrintFileData() {
	for _, proposal := range t.ProposalResponses {
		for _, readWriteSet := range proposal.ReadWriteSet.WriteKeyValues.WriteKeyValues {
			logger.InfoLogger.Println("received file:", string(readWriteSet.Value))
		}
	}
}

func (t *ClientTransaction) MakeTransactionBenchmarkExecutorWithClientFullSign(bconfig *protos.BenchmarkConfig, endorsementPolicy int) (*protos.Transaction, error) {
	if t.Status == protos.TransactionStatus_FAILED_GENERAL {
		return nil, errors.New("transaction failed")
	}
	tempTransaction := &protos.Transaction{
		TargetSystem:      bconfig.TargetSystem,
		TransactionId:     t.TransactionId,
		ClientId:          config.Config.UUID,
		ContractName:      bconfig.ContractName,
		NodeSignatures:    map[string][]byte{},
		EndorsementPolicy: int32(endorsementPolicy),
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.NodeSignatures[proposal.NodeId] = proposal.NodeSignature
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.ReadWriteSet = proposal.ReadWriteSet
		break
	}
	sort.Slice(tempTransaction.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	marshalledReadWriteSet, _ := proto.Marshal(tempTransaction.ReadWriteSet)
	tempTransaction.ClientSignature = t.signer.Sign(marshalledReadWriteSet)

	return tempTransaction, nil
}

func (t *ClientTransaction) MakeTransactionBenchmarkExecutorWithClientBaseSign(bconfig *protos.BenchmarkConfig, endorsementPolicy int) (*protos.Transaction, error) {
	if t.Status == protos.TransactionStatus_FAILED_GENERAL {
		return nil, errors.New("transaction failed")
	}
	tempTransaction := &protos.Transaction{
		TargetSystem:      bconfig.TargetSystem,
		TransactionId:     t.TransactionId,
		ClientId:          config.Config.UUID,
		ContractName:      bconfig.ContractName,
		NodeSignatures:    map[string][]byte{},
		EndorsementPolicy: int32(endorsementPolicy),
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.NodeSignatures[proposal.NodeId] = proposal.NodeSignature
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.ReadWriteSet = proposal.ReadWriteSet
		break
	}
	sort.Slice(tempTransaction.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	marshalledReadSet, _ := proto.Marshal(tempTransaction.ReadWriteSet.ReadKeys)
	tempTransaction.ClientSignature = t.signer.Sign(marshalledReadSet)

	return tempTransaction, nil
}
