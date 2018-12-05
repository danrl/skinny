package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/danrl/skinny/proto/lock"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.PersistentFlags().StringVar(&flagInstance, "instance", "", "name of instance to connect to")
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release the lock",
	Run: func(cmd *cobra.Command, args []string) {
		// select default instance if no one was specified
		if flagInstance == "" {
			flagInstance = cfgDefaultInstance
		}

		// connect to instance
		address := cfgInstances[flagInstance]
		fmt.Printf("📡 connecting to %v (%v)\n", flagInstance, address)
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			fmt.Fprintf(os.Stderr, "dial: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), cfgQuorum.Timeout)
		defer cancel()

		// try to acquire lock
		fmt.Println("🔓 releasing lock")
		client := lock.NewLockClient(conn)
		resp, err := client.Release(ctx, &lock.ReleaseRequest{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.Released {
			fmt.Println("✅ success")
		} else {
			fmt.Println("🚫 failed")
		}
	},
}
