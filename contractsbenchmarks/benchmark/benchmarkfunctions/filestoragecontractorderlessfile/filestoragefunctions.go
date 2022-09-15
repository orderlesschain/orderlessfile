package filestoragecontractorderlessfile

import (
	"gitlab.lrz.de/orderless/orderlessfile/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlessfile/internal/crdtmanagerv2"
	"gitlab.lrz.de/orderless/orderlessfile/internal/customcrypto/hasher"
	protos "gitlab.lrz.de/orderless/orderlessfile/protos/goprotos"
	"strconv"
)

func UploadFile(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	clientFileIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	clientFileId := AllClientsIds[clientFileIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[clientFileIdIndex]++
	counter := options.SingleFunctionCounter.Counter[clientFileIdIndex]
	if _, ok := options.SingleFunctionCounter.UsedClientsMap[clientFileIdIndex]; !ok {
		options.SingleFunctionCounter.UsedClientsMap[clientFileIdIndex] = true
		options.SingleFunctionCounter.UsedClientsList = append(options.SingleFunctionCounter.UsedClientsList, clientFileIdIndex)
	}
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(clientFileId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	fileShardBytes := textGeneratorHashed(options.NumberOfKeysSecond)
	shardId := counter % TotalFileShards
	fileVersion := counter / TotalFileShards
	fileShardHashed := hasher.Hash(append(fileShardBytes, []byte(strconv.Itoa(fileVersion))...))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs: []string{clientFileId + "#", strconv.Itoa(fileVersion),
			strconv.Itoa(shardId), strconv.Itoa(options.NumberOfKeysSecond), crdtmanagerv2.ClockToString(clock)},
		FileShardHashedByte: fileShardHashed,
		ContractMethodName:  "uploadfile",
		WriteReadType:       protos.ProposalRequest_Write,
	}
}

func DownloadFile(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	options.SingleFunctionCounter.Lock.Lock()
	clientFileRandomIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(len(options.SingleFunctionCounter.UsedClientsList))
	clientFileId := AllClientsIds[options.SingleFunctionCounter.UsedClientsList[clientFileRandomIndex]]
	options.SingleFunctionCounter.Lock.Unlock()
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{clientFileId + "#"},
		ContractMethodName: "downloadfile",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func ReadWriteUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return DownloadFile(options)
	} else {
		return UploadFile(options)
	}
}

func WriteUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	return UploadFile(options)
}

func ReadUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	return DownloadFile(options)
}

func TextGenerator(shardSize string) string {
	if shardSize == "10" {
		return shard10k
	}
	if shardSize == "25" {
		return shard25k
	}
	if shardSize == "50" {
		return shard50k
	}
	if shardSize == "75" {
		return shard75k
	}
	if shardSize == "100" {
		return shard100k
	}
	if shardSize == "200" {
		return shard200k
	}
	if shardSize == "300" {
		return shard300k
	}
	if shardSize == "400" {
		return shard400k
	}
	if shardSize == "500" {
		return shard500k
	}
	if shardSize == "1000" {
		return shard1000k
	}
	return ""
}

func textGeneratorHashed(shardSize int) []byte {
	if shardSize == 10 {
		return shard10kByte
	}
	if shardSize == 25 {
		return shard25kByte
	}
	if shardSize == 50 {
		return shard50kByte
	}
	if shardSize == 75 {
		return shard75kByte
	}
	if shardSize == 100 {
		return shard100kByte
	}
	if shardSize == 200 {
		return shard200kByte
	}
	if shardSize == 300 {
		return shard300kByte
	}
	if shardSize == 400 {
		return shard400kByte
	}
	if shardSize == 500 {
		return shard500kByte
	}
	if shardSize == 1000 {
		return shard1000kByte
	}
	return nil
}
