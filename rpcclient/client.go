package rpcclient

import (
	"context"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
	"github.com/qiaopengjun5162/web3-multichain-sync/rpcclient/chain-account/account"
	"github.com/qiaopengjun5162/web3-multichain-sync/rpcclient/chain-account/common"
)

type WalletChainAccountClient struct {
	Ctx             context.Context
	ChainName       string
	AccountRpClient account.WalletAccountServiceClient
}

func NewWalletChainAccountClient(ctx context.Context, rpc account.WalletAccountServiceClient, chainName string) (*WalletChainAccountClient, error) {
	log.Info("New account chain rpc client", "chainName", chainName)
	return &WalletChainAccountClient{Ctx: ctx, AccountRpClient: rpc, ChainName: chainName}, nil
}

func (wac *WalletChainAccountClient) ExportAddressByPubKey(typeOrVersion, publicKey string) string {
	req := &account.ConvertAddressRequest{
		Chain:     wac.ChainName,
		Type:      typeOrVersion,
		PublicKey: publicKey,
	}
	address, err := wac.AccountRpClient.ConvertAddress(wac.Ctx, req)
	if err != nil {
		log.Error("covert address fail", "err", err)
		return ""
	}
	if address.Code == common.ReturnCode_ERROR {
		log.Error("covert address fail", "err", err)
		return ""
	}
	return address.Address
}

func (wac *WalletChainAccountClient) GetBlockInfo(blockNumber *big.Int) ([]*account.BlockInfoTransactionList, error) {
	req := &account.BlockNumberRequest{
		Chain:  wac.ChainName,
		Height: blockNumber.Int64(),
		ViewTx: true,
	}
	blockInfo, err := wac.AccountRpClient.GetBlockByNumber(wac.Ctx, req)
	if err != nil {
		log.Error("get block GetBlockByNumber fail", "err", err)
		return nil, err
	}
	if blockInfo.Code == common.ReturnCode_ERROR {
		log.Error("get block info fail", "err", err)
		return nil, err
	}
	return blockInfo.Transactions, nil
}

func (wac *WalletChainAccountClient) GetAccount(address string) (int, int, int) {
	req := &account.AccountRequest{
		Chain:           wac.ChainName,
		Network:         "mainnet",
		Address:         address,
		ContractAddress: "0x00",
	}

	accountInfo, err := wac.AccountRpClient.GetAccount(wac.Ctx, req)
	if err != nil {
		log.Info("GetAccount fail", "err", err)
		return 0, 0, 0
	}

	if accountInfo.Code == common.ReturnCode_ERROR {
		log.Info("get account info fail", "msg", accountInfo.Msg)
		return 0, 0, 0
	}

	accountNumber, err := strconv.Atoi(accountInfo.AccountNumber)
	if err != nil {
		log.Info("failed to convert account number", "err", err)
		return 0, 0, 0
	}

	sequence, err := strconv.Atoi(accountInfo.Sequence)
	if err != nil {
		log.Info("failed to convert sequence", "err", err)
		return 0, 0, 0
	}

	balance, err := strconv.Atoi(accountInfo.Balance)
	if err != nil {
		log.Info("failed to convert balance", "err", err)
		return 0, 0, 0
	}

	return accountNumber, sequence, balance
}
