package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	programID = solana.MustPublicKeyFromBase58("JBHnGuqyTxwbnpWiTjRqydkwSRYJFcQnwUQnLfesdRG9")
)

func main() {
	client := rpc.New("https://rpc.ankr.com/solana_devnet")

	// Fetch signatures for the program ID
	signatures, err := fetchSignaturesForAddress(client, programID)
	if err != nil {
		slog.Error("failed to fetch signatures for programID", "programID", programID, "err", err)
	}

	// Fetch and process transactions for each signature
	for _, signature := range signatures {
		tx, err := client.GetParsedTransaction(context.Background(), signature.Signature, &rpc.GetParsedTransactionOpts{})
		if err != nil {
			slog.Error("Failed to fetch parsed transaction", "err", err)
			continue
		}
		events, err := checkCreateEvent(programID, tx)
		if err != nil {
			slog.Error("Failed to fetch transaction events", "err", err)
		}
		if len(events) > 0 {
			for _, event := range events {
				slog.Info("Found event", "event", event)
			}
		}
	}
}

// Fetch signatures for the given address (program ID)
func fetchSignaturesForAddress(client *rpc.Client, address solana.PublicKey) ([]*rpc.TransactionSignature, error) {
	var signatures []*rpc.TransactionSignature
	var before solana.Signature

	limit := 10

	for {
		options := &rpc.GetSignaturesForAddressOpts{
			Limit:  &limit,
			Before: before,
		}

		signatureList, err := client.GetSignaturesForAddressWithOpts(context.Background(), address, options)
		if err != nil {
			return nil, err
		}

		if len(signatureList) == 0 {
			slog.Info("not found")
			break
		}

		signatures = append(signatures, signatureList...)
		before = signatureList[len(signatureList)-1].Signature
		time.Sleep(400 * time.Millisecond) // To prevent rate limiting
	}

	return signatures, nil
}

func checkCreateEvent(programID solana.PublicKey, txResponse *rpc.GetParsedTransactionResult) ([]CreateEvent, error) {
	if txResponse == nil {
		return nil, nil
	}

	var events []CreateEvent

	for _, innerInstructions := range txResponse.Meta.InnerInstructions {
		for _, instruction := range innerInstructions.Instructions {
			if instruction.ProgramId == token.ProgramID {
				jsonData, err := instruction.Parsed.MarshalJSON()
				if err != nil {
					slog.Error("Failed to marshal instruction", "err", err)
					continue
				}
				var mintoInfo MintToInfo
				if err = json.Unmarshal(jsonData, &mintoInfo); err != nil {
					slog.Error("Failed to unmarshal instruction", "err", err)
					continue
				}
				if mintoInfo.Info.Mint == "" || mintoInfo.Info.Account == "" || mintoInfo.Info.Amount == "" {
					continue
				}
				events = append(events, CreateEvent{
					// Name:   "",
					// Symbol: "",
					// Uri:    "",
					Mint:        mintoInfo.Info.Mint,
					TotalSupply: mintoInfo.Info.Amount,
				})
			}
		}
	}

	return events, nil
}

type CreateEvent struct {
	Name        string
	Symbol      string
	Uri         string
	Mint        string
	TotalSupply string
}

type MintToInfo struct {
	Info struct {
		Account string `json:"account"`
		Amount  string `json:"amount"`
		Mint    string `json:"mint"`
	} `json:"info"`
}
