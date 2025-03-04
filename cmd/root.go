/*
Copyright © 2025 Ngalim Siregar ngalim.siregar@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

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

		processTransactions(ctx, rpcClient, resp, pubKey)
	}
}

func processTransactions(ctx context.Context, rpcClient *rpc.Client, resp *ws.LogResult, pubKey solana.PublicKey) {
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

	instructions := tx.Message.Instructions
	lastInstruction := instructions[len(instructions)-1]
	programId, err := tx.ResolveProgramIDIndex(lastInstruction.ProgramIDIndex)
	if err != nil {
		log.Fatalf("error get program ids: %v", err)
	}

	accounts, err := lastInstruction.ResolveInstructionAccounts(&tx.Message)
	if err != nil {
		log.Fatalf("failed resolve instruction accounts: %v", err)
	}

	sender := accounts[0].PublicKey
	recipient := accounts[1].PublicKey

	if programId == solana.SystemProgramID {
		totalAmount := amountChanges(transaction.Meta.PreBalances, transaction.Meta.PostBalances)
		fmt.Printf("%s has sent %.6f SOL to %s\n", sender, totalAmount, recipient)
	}

	if programId == solana.TokenProgramID {
		totalAmount := splAmountChanges(transaction.Meta.PreTokenBalances, transaction.Meta.PostTokenBalances)
		tokenName := transaction.Meta.PreTokenBalances[0].Owner
		fmt.Printf("%s has sent %.6f %s to %s\n", sender, totalAmount, tokenName, recipient)
	}

}

func amountChanges(preBalances []uint64, postBalances []uint64) float64 {
	var maxValue float64

	maxValue = 0
	for i := 0; i < len(preBalances); i++ {
		result := float64(postBalances[i]) - float64(preBalances[i])
		if result > maxValue {
			maxValue = result
		}
	}

	return float64(maxValue / float64(solana.LAMPORTS_PER_SOL))
}

func splAmountChanges(preBalances []rpc.TokenBalance, postBalances []rpc.TokenBalance) float64 {
	var maxValue float64

	maxValue = 0
	for i := 0; i < len(preBalances); i++ {
		result := float64(postBalances[i].UiTokenAmount.Decimals) - float64(preBalances[i].UiTokenAmount.Decimals)
		if result > maxValue {
			maxValue = result
		}
	}

	return maxValue
}
