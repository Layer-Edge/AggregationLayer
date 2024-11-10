package da

import (
    "context"
    "fmt"
    "os/exec"
    "cosmossdk.io/math"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/tx"
    "github.com/cosmos/cosmos-sdk/codec"
    codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
    "github.com/cosmos/cosmos-sdk/crypto/keyring"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/tx/signing"
    banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type CosmosClientConfig struct {
    ChainID        string
    RPCEndpoint    string
    AccountPrefix  string
    KeyringBackend string
    KeyName        string
    HomeDir        string
}

type CosmosClient struct {
    config     CosmosClientConfig
    clientCtx  client.Context
    kr         keyring.Keyring
    senderAddr sdk.AccAddress
}

func (c *CosmosClient) Init(cfg CosmosClientConfig) error {
    c.config = cfg

    clientCtx := client.Context{}
    clientCtx = clientCtx.
        WithChainID(cfg.ChainID).
        WithNodeURI(cfg.RPCEndpoint).
        WithBroadcastMode("block")

    sdkConfig := sdk.GetConfig()
    sdkConfig.SetBech32PrefixForAccount(cfg.AccountPrefix, cfg.AccountPrefix+"pub")

    interfaceRegistry := codecTypes.NewInterfaceRegistry()
    marshaler := codec.NewProtoCodec(interfaceRegistry)

    kr, err := keyring.New(
        "layeredge.info",
        "test",
        cfg.HomeDir,
        nil,
        marshaler,
    )
    if err != nil {
        return fmt.Errorf("failed to create keyring: %v", err)
    }

    senderInfo, err := kr.Key(cfg.KeyName)
    if err != nil {
        return fmt.Errorf("failed to get sender info: %v", err)
    }

    senderAddr, err := senderInfo.GetAddress()
    if err != nil {
        return fmt.Errorf("failed to get sender address: %v", err)
    }

    c.clientCtx = clientCtx.WithFromAddress(senderAddr)
    c.kr = kr
    c.senderAddr = senderAddr

    return nil
}

func (c *CosmosClient) SendData(data string) error {
    amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1)))
    msg := banktypes.NewMsgSend(c.senderAddr, c.senderAddr, amount)

    accountNumber, sequence, err := c.getAccountNumberAndSequence()
    if err != nil {
        return fmt.Errorf("failed to get account number and sequence: %v", err)
    }

    gasPrices := sdk.NewDecCoins(sdk.NewDecCoin("stake", math.NewInt(1)))
    txf := tx.Factory{}.
        WithTxConfig(c.clientCtx.TxConfig).
        WithAccountNumber(accountNumber).
        WithSequence(sequence).
        WithGas(200000).
        WithGasPrices(gasPrices.String()).
        WithChainID(c.config.ChainID).
        WithMemo(data).
        WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

    txBuilder, err := txf.BuildUnsignedTx(msg)
    if err != nil {
        return fmt.Errorf("failed to build unsigned transaction: %v", err)
    }

    err = tx.Sign(context.Background(), txf, c.config.KeyName, txBuilder, true)
    if err != nil {
        return fmt.Errorf("failed to sign transaction: %v", err)
    }

    txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
    if err != nil {
        return fmt.Errorf("failed to encode transaction: %v", err)
    }

    res, err := c.clientCtx.BroadcastTx(txBytes)
    if err != nil {
        return fmt.Errorf("failed to broadcast transaction: %v", err)
    }

    fmt.Printf("Transaction broadcasted successfully. Hash: %s\n", res.TxHash)
    fmt.Printf("Data sent in memo: %s\n", data)

    return nil
}

func (c *CosmosClient) getAccountNumberAndSequence() (uint64, uint64, error) {
    accNum, seq, err := c.clientCtx.AccountRetriever.GetAccountNumberSequence(c.clientCtx, c.senderAddr)
    if err != nil {
        return 0, 0, err
    }
    return accNum, seq, nil
}

func (c *CosmosClient) Send(data string, addr string) ([]byte, error) {
    cmd := exec.Command(BashScriptPath+"/run-cosmos-tx.sh", "-m",  data, "-r", addr)
    out, err := cmd.Output()
    fmt.Printf("Sending Data: %s\n", out)
    return out, err
}
