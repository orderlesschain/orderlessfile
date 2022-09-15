package benchmarkfunctions

import (
	"errors"
	"gitlab.lrz.de/orderless/orderlessfile/contractsbenchmarks/benchmark/benchmarkfunctions/filestoragecontractorderlessfile"
	"gitlab.lrz.de/orderless/orderlessfile/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlessfile/internal/logger"
	"strings"
)

func GetBenchmarkFunctions(contactName string, benchmarkFunctionName string) (contractinterface.BenchmarkFunction, error) {
	contactName = strings.ToLower(contactName)
	benchmarkFunctionName = strings.ToLower(benchmarkFunctionName)
	switch contactName {
	case "filestoragecontractorderlessfile":
		switch benchmarkFunctionName {
		case "uploadfile":
			return filestoragecontractorderlessfile.UploadFile, nil
		case "downloadfile":
			return filestoragecontractorderlessfile.DownloadFile, nil
		case "readwritetransactionwarm":
			return filestoragecontractorderlessfile.ReadWriteUniformTransactionWarm, nil
		case "writetransactionwarm":
			return filestoragecontractorderlessfile.WriteUniformTransactionWarm, nil
		case "readtransactionwarm":
			return filestoragecontractorderlessfile.ReadUniformTransactionWarm, nil
		}
	}
	logger.ErrorLogger.Println("smart contract function was not found")
	return nil, errors.New("functions not found")
}
