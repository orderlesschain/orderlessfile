package transaction

import (
	"gitlab.lrz.de/orderless/orderlessfile/internal/config"
	"gitlab.lrz.de/orderless/orderlessfile/protos/goprotos"
	"sync"
	"time"
)

type DequeuedProposals struct {
	DequeuedProposals []*protos.ProposalRequest
}

type DequeuedTransactions struct {
	DequeuedTransactions []*protos.Transaction
}

type NodeTransactionJournal struct {
	journal                        map[string]bool
	journalNodes                   map[string]bool
	journalLock                    *sync.Mutex
	journalNodesLock               *sync.Mutex
	proposalsQueue                 []*protos.ProposalRequest
	transactionsQueue              []*protos.Transaction
	blockQueue                     []*protos.Block
	DequeuedProposalsChan          chan *DequeuedProposals
	DequeuedTransactionsChan       chan *DequeuedTransactions
	RequestBlockChan               chan bool
	proposalsQueueLock             *sync.Mutex
	transactionsQueueLock          *sync.Mutex
	blockQueueLock                 *sync.Mutex
	tickerDuration                 time.Duration
	proposalBatchSizePertTicker    int
	transactionBatchSizePertTicker int
}

func InitTransactionJournal() *NodeTransactionJournal {
	tempJournal := &NodeTransactionJournal{
		journal:                  map[string]bool{},
		journalNodes:             map[string]bool{},
		journalLock:              &sync.Mutex{},
		journalNodesLock:         &sync.Mutex{},
		proposalsQueue:           []*protos.ProposalRequest{},
		transactionsQueue:        []*protos.Transaction{},
		blockQueue:               []*protos.Block{},
		DequeuedProposalsChan:    make(chan *DequeuedProposals),
		DequeuedTransactionsChan: make(chan *DequeuedTransactions),
		RequestBlockChan:         make(chan bool),
		proposalsQueueLock:       &sync.Mutex{},
		transactionsQueueLock:    &sync.Mutex{},
		blockQueueLock:           &sync.Mutex{},
	}
	tickerDurationMS := config.Config.QueueTickerDurationMS
	tempJournal.tickerDuration = time.Duration(tickerDurationMS) * time.Millisecond
	tempJournal.proposalBatchSizePertTicker = config.Config.ProposalQueueConsumptionRateTPS / (1000 / tickerDurationMS)
	tempJournal.transactionBatchSizePertTicker = config.Config.TransactionQueueConsumptionRateTPS / (1000 / tickerDurationMS)

	return tempJournal
}

func (tj *NodeTransactionJournal) AddTransactionToJournalIfNotExist(transaction *protos.Transaction) bool {
	tj.journalLock.Lock()
	if _, ok := tj.journal[transaction.TransactionId]; !ok {
		tj.journal[transaction.TransactionId] = true
		tj.journalLock.Unlock()
		return true
	}
	tj.journalLock.Unlock()
	return false
}

func (tj *NodeTransactionJournal) IsTransactionInJournal(transaction *protos.Transaction) bool {
	tj.journalLock.Lock()
	_, ok := tj.journal[transaction.TransactionId]
	tj.journalLock.Unlock()
	return ok
}

func (tj *NodeTransactionJournal) AddTransactionToJournalNodeIfNotExist(transaction *protos.Transaction) bool {
	tj.journalNodesLock.Lock()
	if _, ok := tj.journalNodes[transaction.TransactionId]; !ok {
		tj.journalNodes[transaction.TransactionId] = true
		tj.journalNodesLock.Unlock()
		return true
	}
	tj.journalNodesLock.Unlock()
	return false
}

func (tj *NodeTransactionJournal) AddProposalToQueue(proposal *protos.ProposalRequest) {
	tj.proposalsQueueLock.Lock()
	tj.proposalsQueue = append(tj.proposalsQueue, proposal)
	tj.proposalsQueueLock.Unlock()
}

func (tj *NodeTransactionJournal) AddTransactionToQueue(transaction *protos.Transaction) {
	tj.transactionsQueueLock.Lock()
	tj.transactionsQueue = append(tj.transactionsQueue, transaction)
	tj.transactionsQueueLock.Unlock()
}

func (tj *NodeTransactionJournal) RunProposalQueueProcessorTicker() {
	ticker := time.NewTicker(tj.tickerDuration)
	for range ticker.C {
		tj.proposalsQueueLock.Lock()
		toDequeueLength := len(tj.proposalsQueue)
		if toDequeueLength == 0 {
			tj.proposalsQueueLock.Unlock()
			continue
		}
		if toDequeueLength > tj.proposalBatchSizePertTicker {
			toDequeueLength = tj.proposalBatchSizePertTicker
		}
		tempDequeuedProposals := &DequeuedProposals{
			DequeuedProposals: make([]*protos.ProposalRequest, 0, toDequeueLength),
		}
		for i := 0; i < toDequeueLength; i++ {
			tempDequeuedProposals.DequeuedProposals = append(tempDequeuedProposals.DequeuedProposals, tj.proposalsQueue[i])
			tj.proposalsQueue[i] = nil
		}
		tj.proposalsQueue = tj.proposalsQueue[toDequeueLength:]
		tj.proposalsQueueLock.Unlock()
		tj.DequeuedProposalsChan <- tempDequeuedProposals
	}
}

func (tj *NodeTransactionJournal) RunTransactionsQueueProcessorTicker() {
	ticker := time.NewTicker(tj.tickerDuration)
	for range ticker.C {
		tj.transactionsQueueLock.Lock()
		toDequeueLength := len(tj.transactionsQueue)
		if toDequeueLength == 0 {
			tj.transactionsQueueLock.Unlock()
			continue
		}
		if toDequeueLength > tj.transactionBatchSizePertTicker {
			toDequeueLength = tj.transactionBatchSizePertTicker
		}
		tempDequeuedTransactions := &DequeuedTransactions{
			DequeuedTransactions: make([]*protos.Transaction, 0, toDequeueLength),
		}
		for i := 0; i < toDequeueLength; i++ {
			tempDequeuedTransactions.DequeuedTransactions = append(tempDequeuedTransactions.DequeuedTransactions, tj.transactionsQueue[i])
			tj.transactionsQueue[i] = nil
		}
		tj.transactionsQueue = tj.transactionsQueue[toDequeueLength:]
		tj.transactionsQueueLock.Unlock()
		tj.DequeuedTransactionsChan <- tempDequeuedTransactions
	}
}
