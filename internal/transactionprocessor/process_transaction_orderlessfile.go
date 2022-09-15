package transactionprocessor

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlessfile/internal/blockprocessor"
	"gitlab.lrz.de/orderless/orderlessfile/internal/logger"
	"gitlab.lrz.de/orderless/orderlessfile/internal/profiling"
	"gitlab.lrz.de/orderless/orderlessfile/internal/transaction"
	protos "gitlab.lrz.de/orderless/orderlessfile/protos/goprotos"
	"reflect"
	"sort"
)

var byzantineTemperedMessage = []byte("tempered")

type OrderlessFileNodeTransactionResponseSubscriber struct {
	stream   protos.TransactionService_SubscribeNodeTransactionsServer
	finished chan<- bool
}

func (p *Processor) signProposalResponseOrderlessFile(proposalResponse *protos.ProposalResponse) (*protos.ProposalResponse, error) {
	sort.Slice(proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return proposalResponse.ReadWriteSet.ReadKeys.ReadKeys[i].Key < proposalResponse.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	marshalledReadWriteSet, err := proto.Marshal(proposalResponse.ReadWriteSet)
	if err != nil {
		return p.makeFailedProposal(proposalResponse.ProposalId), nil
	}
	if p.ShouldFailByzantineTampered() {
		marshalledReadWriteSet = append(marshalledReadWriteSet, byzantineTemperedMessage...)
	}
	proposalResponse.NodeSignature = p.signer.Sign(marshalledReadWriteSet)
	return proposalResponse, nil
}

func (p *Processor) preProcessValidateReadWriteSetOrderlessFile(tx *protos.Transaction) error {
	sort.Slice(tx.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tx.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tx.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tx.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tx.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tx.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	txDigest, err := proto.Marshal(tx.ReadWriteSet)
	if err != nil {
		return err
	}
	passedSignature := int32(0)
	for node, nodeSign := range tx.NodeSignatures {
		err = p.signer.Verify(node, txDigest, nodeSign)
		if err == nil {
			passedSignature++
		}
	}
	if passedSignature < tx.EndorsementPolicy {
		return errors.New("failed signature validation")
	}
	err = p.signer.Verify(tx.ClientId, txDigest, tx.ClientSignature)
	if err != nil {
		return err
	}
	return nil
}

func (p *Processor) ProcessProposalOrderlessFileStream(proposal *protos.ProposalRequest) {
	p.txJournal.AddProposalToQueue(proposal)
}

func (p *Processor) runProposalQueueProcessingOrderlessFile() {
	for {
		proposals := <-p.txJournal.DequeuedProposalsChan
		go p.processDequeuedProposalsOrderlessFile(proposals)
	}
}

func (p *Processor) processDequeuedProposalsOrderlessFile(proposals *transaction.DequeuedProposals) {
	for _, proposal := range proposals.DequeuedProposals {
		go p.processProposalOrderlessFile(proposal)
	}
}

func (p *Processor) processProposalOrderlessFile(proposal *protos.ProposalRequest) {
	response, err := p.executeContract(proposal)
	if err != nil {
		p.sendProposalResponseToSubscriber(proposal.ClientId, response)
		return
	}
	response, err = p.signProposalResponseOrderlessFile(response)
	if err != nil {
		p.sendProposalResponseToSubscriber(proposal.ClientId, response)
		return
	}
	p.sendProposalResponseToSubscriber(proposal.ClientId, response)
}

func (p *Processor) ProcessTransactionOrderlessFileStream(tx *protos.Transaction) {
	p.txJournal.AddTransactionToQueue(tx)
}

func (p *Processor) runTransactionQueueProcessingOrderlessFile() {
	for {
		transactions := <-p.txJournal.DequeuedTransactionsChan
		go p.processDequeuedTransactionsOrderlessFile(transactions)
	}
}

func (p *Processor) processDequeuedTransactionsOrderlessFile(transactions *transaction.DequeuedTransactions) {
	for _, tx := range transactions.DequeuedTransactions {
		go p.processTransactionOrderlessFile(tx)
	}
}

func (p *Processor) processTransactionOrderlessFile(tx *protos.Transaction) {
	if !tx.FromNode {
		if err := p.preProcessValidateReadWriteSetOrderlessFile(tx); err != nil {
			tx.Status = protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION
		}
	}
	shouldBeAdded := p.txJournal.AddTransactionToJournalIfNotExist(tx)
	if shouldBeAdded {
		p.blockProcessor.TransactionChan <- tx
	} else {
		if !tx.FromNode {
			p.finalizeAlreadySubmittedTransactionFromClientOrderlessFile(tx)
		}
	}
}

func (p *Processor) processTransactionFromOtherNodesOrderlessFile(txs []*protos.Transaction) {
	for _, tx := range txs {
		shouldBeAdded := p.txJournal.AddTransactionToJournalNodeIfNotExist(tx)
		if shouldBeAdded {
			alreadyInJournal := p.txJournal.IsTransactionInJournal(tx)
			if !alreadyInJournal {
				tx.FromNode = true
				p.txJournal.AddTransactionToQueue(tx)
			}
		}
	}
}

func (p *Processor) runTransactionProcessorOrderlessFile() {
	for {
		block := <-p.blockProcessor.MinedBlock
		go p.processBlockOrderlessFile(block)
	}
}

func (p *Processor) processBlockOrderlessFile(block *blockprocessor.MinedBlock) {
	for _, tx := range block.Transactions {
		go p.processTransactionInBlockOrderlessFile(tx, block)
	}
}

func (p *Processor) processTransactionInBlockOrderlessFile(tx *protos.Transaction, block *blockprocessor.MinedBlock) {
	if tx.Status == protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
		p.finalizeTransactionResponseOrderlessFile(tx.ClientId,
			p.makeFailedTransactionResponse(tx.TransactionId, tx.Status, block.Block.ThisBlockHash), tx)
	} else {
		response, err := p.commitTransactionOrderlessFile(tx)
		if err == nil {
			response.BlockHeader = block.Block.ThisBlockHash
			p.finalizeTransactionResponseOrderlessFile(tx.ClientId, response, tx)
		} else {
			p.finalizeTransactionResponseOrderlessFile(tx.ClientId,
				p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_DATABASE, block.Block.ThisBlockHash), tx)
		}
	}
}

func (p *Processor) commitTransactionOrderlessFile(tx *protos.Transaction) (*protos.TransactionResponse, error) {
	if len(tx.ShardData) > 0 {
		p.crdtManagerOrderlessFile.ApplyFileOperationsWarm(tx)
		if err := p.addTransactionToDatabaseOrderlessFile(tx); err != nil {
			logger.InfoLogger.Println(err)
			return nil, err
		}
	}
	return p.makeSuccessTransactionResponse(tx.TransactionId, []byte{}), nil
}

func (p *Processor) addTransactionToDatabaseOrderlessFile(tx *protos.Transaction) error {
	dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
	var err error
	if err = dbOp.PutKeyValueNoVersion(tx.FileID+"-"+tx.ShardId, []byte(tx.ShardData+"-"+tx.FileVersion)); err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	return err
}

func (p *Processor) finalizeAlreadySubmittedTransactionFromClientOrderlessFile(transaction *protos.Transaction) {
	var response *protos.TransactionResponse
	if transaction.Status == protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
		response = p.makeFailedTransactionResponse(transaction.TransactionId, transaction.Status, []byte("BlockHeader"))
	} else {
		response = p.makeSuccessTransactionResponse(transaction.TransactionId, []byte("BlockHeader"))
	}
	response.NodeSignature = p.signer.Sign(response.BlockHeader)
	p.sendTransactionResponseToSubscriber(transaction.ClientId, response)
}

func (p *Processor) finalizeTransactionResponseOrderlessFile(clientId string, txResponse *protos.TransactionResponse, transaction *protos.Transaction) {
	if !transaction.FromNode {
		txResponse.NodeSignature = p.signer.Sign(txResponse.BlockHeader)
		p.sendTransactionResponseToSubscriber(clientId, txResponse)
	}
}

func (p *Processor) sendTransactionBatchToSubscribedNodeOrderlessFile(transactions []*protos.Transaction) {
	if p.ShouldFailByzantineNetwork() {
		return
	}
	gossips := &protos.NodeTransactionResponse{
		Transaction: transactions,
	}
	p.nodeSubscribersLock.RLock()
	nodes := reflect.ValueOf(p.OrderlessFileNodeTransactionResponseSubscriber).MapKeys()
	if profiling.IsBandwidthProfiling {
		_ = proto.Size(gossips)
	}
	p.nodeSubscribersLock.RUnlock()
	for _, node := range nodes {
		nodeId := node.String()
		p.nodeSubscribersLock.RLock()
		streamer, ok := p.OrderlessFileNodeTransactionResponseSubscriber[nodeId]
		p.nodeSubscribersLock.RUnlock()
		if !ok {
			logger.ErrorLogger.Println("Node was not found in the subscribers streams.", nodeId)
			return
		}
		if err := streamer.stream.Send(gossips); err != nil {
			streamer.finished <- true
			logger.ErrorLogger.Println("Could not send the response to the node " + nodeId)
			p.nodeSubscribersLock.Lock()
			delete(p.OrderlessFileNodeTransactionResponseSubscriber, nodeId)
			p.nodeSubscribersLock.Unlock()
		}
	}
}

func (p *Processor) NodeTransactionResponseSubscriptionOrderlessFile(subscription *protos.TransactionResponseEventSubscription,
	stream protos.TransactionService_SubscribeNodeTransactionsServer) error {
	finished := make(chan bool)
	p.nodeSubscribersLock.Lock()
	p.OrderlessFileNodeTransactionResponseSubscriber[subscription.ComponentId] = &OrderlessFileNodeTransactionResponseSubscriber{
		stream:   stream,
		finished: finished,
	}
	p.nodeSubscribersLock.Unlock()
	cntx := stream.Context()
	for {
		select {
		case <-finished:
			return nil
		case <-cntx.Done():
			return nil
		}
	}
}
