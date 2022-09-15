package filestoragecontractorderlessfile

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlessfile/internal/contract/contractinterface"
	protos "gitlab.lrz.de/orderless/orderlessfile/protos/goprotos"
	"strings"
)

type FileStorageContract struct {
}

func NewContract() *FileStorageContract {
	return &FileStorageContract{}
}

func (t *FileStorageContract) Invoke(shim contractinterface.ShimInterface, proposal *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	proposal.MethodName = strings.ToLower(proposal.MethodName)
	if proposal.MethodName == "uploadfile" {
		return t.uploadFile(shim, proposal.MethodParams, proposal.FileShardHashedData)
	}
	if proposal.MethodName == "downloadfile" {
		return t.downLoad(shim, proposal.MethodParams)
	}
	return shim.Error(), errors.New("method name not found")
}

func (t *FileStorageContract) uploadFile(shim contractinterface.ShimInterface, args []string, fileShardData []byte) (*protos.ProposalResponse, error) {
	fileIndex := args[0]
	shim.PutKeyValue(fileIndex, fileShardData)
	return shim.Success(), nil
}

func (t *FileStorageContract) downLoad(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	fileIndex := args[0]

	fileCRDT, errCRDT := shim.GetCRDTObjectV2Warm(fileIndex)
	if errCRDT != nil {
		return shim.Error(), errCRDT
	}
	fileCRDT.Lock.RLock()
	fileBinary, _ := json.Marshal(fileCRDT)
	fileCRDT.Lock.RUnlock()
	shim.PutKeyValue(fileIndex, fileBinary)
	return shim.Success(), nil
}
