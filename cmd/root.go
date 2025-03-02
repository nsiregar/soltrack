/*
Copyright © 2025 Ngalim Siregar ngalim.siregar@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/spf13/cobra"
)

var walletAddress string

var rootCmd = &cobra.Command{
	Use:   "soltrack",
	Short: "Tracks Solana wallet transactions",
	Long:  "Tracks Solana wallet transactions using WebSocket API",
	Run: func(cmd *cobra.Command, args []string) {
		runTracker()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&walletAddress, "wallet", "w", "", "Solana wallet address to track (required)")
	if err := rootCmd.MarkFlagRequired("wallet"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTracker() {
	pubKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		log.Fatalf("Invalid wallet address: %v", err)
	}

	fmt.Println("Wallet validated! - Monitoring Transactions...")

	ctx := context.Background()
	// rpcURL := rpc.DevNet_RPC
	// wsURL := rpc.DevNet_WS
	rpcURL := os.Getenv("RPC_URL")
	wsURL := os.Getenv("WS_URL")

	if rpcURL == "" || wsURL == "" {
		log.Fatal("RPC_URL and WS_URL environment variables must be set")
	}

	client, err := ws.Connect(ctx, wsURL)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer client.Close()
	rpcClient := rpc.New(rpcURL)

	sub, err := client.LogsSubscribeMentions(pubKey, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("Failed to subscribe to account: %v", err)
	}
	defer sub.Unsubscribe()

	for {
		resp, err := sub.Recv(ctx)
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			continue
		}

		processTransactions(ctx, rpcClient, resp)
	}
}

func processTransactions(ctx context.Context, rpcClient *rpc.Client, resp *ws.LogResult) {
	transaction, err := rpcClient.GetTransaction(
		ctx,
		resp.Value.Signature,
		&rpc.GetTransactionOpts{Encoding: solana.EncodingBase64},
	)
	if err != nil {
		return
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBorshDecoder(transaction.Transaction.GetBinary()))
	if err != nil {
		return
	}

	spew.Dump(tx)

	// for _, instruction := range tx.Message.Instructions {
	// 	// progKey, err := tx.ResolveProgramIDIndex(instruction.ProgramIDIndex)
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }

	// 	// accounts, err := instruction.ResolveInstructionAccounts(&tx.Message)
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }
	// 	var transferInstruction TransferInstruction
	// 	reader := bytes.NewReader(instruction.Data)
	// 	err := binary.Read(reader, binary.LittleEndian, &transferInstruction.Amount)
	// 	if err != nil {
	// 		log.Fatalf("Failed to decode instruction data: %v", err)
	// 	}

	// 	// Print the decoded data
	// 	fmt.Printf("Transfer Amount: %d\n", transferInstruction.Amount)
	// }
}

// type TransferInstruction struct {
// 	Amount uint64
// }
