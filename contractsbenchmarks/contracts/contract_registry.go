package contracts

import (
	"errors"
	"gitlab.lrz.de/orderless/orderlessfile/contractsbenchmarks/contracts/filestoragecontractorderlessfile"
	"gitlab.lrz.de/orderless/orderlessfile/internal/contract/contractinterface"
	"strings"
)

func GetContract(contractName string) (contractinterface.ContractInterface, error) {
	contractName = strings.ToLower(contractName)
	switch contractName {
	case "filestoragecontractorderlessfile":
		return filestoragecontractorderlessfile.NewContract(), nil
	default:
		return nil, errors.New("contract not found")
	}
}

func GetContractNames() []string {
	return []string{
		"filestoragecontractorderlessfile",
	}
}
