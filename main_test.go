package main

import (
	"context"
	"log/slog"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

var (
	exampleCreateHash = solana.MustSignatureFromBase58("VeA7DN7fA1fjhn9dxDQCGJY4nDjPQENwRzjJ5XLW7s3omxsyVzbWWhPTFJoDekCmeMgsWhGqsJZXx2d5PQU7kwy")
	exampleBuyHash    = ""
)

func TestBasic(t *testing.T) {
	client := rpc.New("https://rpc.ankr.com/solana_devnet")

	tx, err := client.GetParsedTransaction(context.Background(), exampleCreateHash, &rpc.GetParsedTransactionOpts{})
	require.NoError(t, err)

	events, err := checkCreateEvent(programID, tx)
	require.NoError(t, err)

	slog.Info("", "events", events)
}
