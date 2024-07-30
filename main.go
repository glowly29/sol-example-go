package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	programID = solana.MustPublicKeyFromBase58("JBHnGuqyTxwbnpWiTjRqydkwSRYJFcQnwUQnLfesdRG9")
)

func main() {
	client := rpc.New(rpc.DevNet_RPC)

	// Fetch signatures for the program ID
	signatures, err := fetchSignaturesForAddress(client, programID)
	if err != nil {
		log.Fatalf("Failed to fetch signatures: %v", err)
	}

	// Fetch and process transactions for each signature
	for _, signature := range signatures {
		tx, err := client.GetParsedTransaction(context.Background(), signature.Signature, &rpc.GetParsedTransactionOpts{})
		if err != nil {
			log.Printf("Failed to fetch transaction %s: %v", signature.Signature.String(), err)
			continue
		}
		events, err := getTransactionEvents(programID, tx)
		if err != nil {
			slog.Error("Failed to fetch transaction events", "err", err)
		}
		_ = events
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

func getTransactionEvents(programID solana.PublicKey, txResponse *rpc.GetParsedTransactionResult) ([]*Event, error) {
	if txResponse == nil {
		return nil, nil
	}

	eventPDA, _, err := solana.FindProgramAddress([][]byte{[]byte("__event_authority")}, programID)
	if err != nil {
		return nil, err
	}

	slog.Info("found event pda", "eventPDA", eventPDA.String())

	var createInsts []*rpc.ParsedInstruction
	for _, innerInstructions := range txResponse.Meta.InnerInstructions {
		for _, instruction := range innerInstructions.Instructions {
			if len(instruction.Accounts) == 1 && instruction.Accounts[0] == eventPDA {
				createInsts = append(createInsts, instruction)
			}
		}
	}

	var events []*Event
	for _, instruction := range createInsts {
		slog.Info("found event create", "data", instruction.Data.String())
		// eventData := base64.StdEncoding.EncodeToString(ixData[8:])
		// event, err := decodeEvent(eventData)
		// if err != nil {
		// 	slog.Error("Error decoding event:", "err", err)
		// 	continue
		// }
		// if event != nil {
		// 	events = append(events, event)
		// }
	}

	return events, nil
}

func decodeEvent(data string) (*Event, error) {
	return &Event{Data: data}, nil
}

type Event struct {
	Data string
}
