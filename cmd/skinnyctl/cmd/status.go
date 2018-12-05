package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/danrl/skinny/proto/control"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	flagWatch bool
)

func init() {
	statusCmd.PersistentFlags().BoolVar(&flagWatch, "watch", false, "watch status report")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Fetch status of a quorum of Skinny instances",
	Run: func(cmd *cobra.Command, args []string) {
		type report struct {
			name      string
			resp      *control.StatusResponse
			comment   string
			timestamp time.Time
		}

		done := make(chan struct{})
		if flagWatch {
			sc := make(chan os.Signal)
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			go func() {
				<-sc
				close(done)
			}()
		} else {
			close(done)
		}

		wg := sync.WaitGroup{}

		reports := make(chan *report)
		for _, in := range cfgQuorum.Instances {
			wg.Add(1)
			go func(name, address string) {
				defer wg.Done()

				// kick-off drawing by sending an empty report at startup
				reports <- &report{name: name}

				// connect to instance
				conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(5*time.Second))
				if err != nil {
					return
				}
				defer conn.Close()
				client := control.NewControlClient(conn)

				// regularly poll for status updates
				for {
					ctx, cancel := context.WithTimeout(context.Background(), cfgQuorum.Timeout)
					resp, err := client.Status(ctx, &control.StatusRequest{})
					if err != nil {
						resp = nil
					}
					reports <- &report{
						name:      name,
						resp:      resp,
						timestamp: time.Now(),
					}
					cancel()

					select {
					case <-time.After(2 * time.Second):
						continue
					case <-done:
						return
					}
				}
			}(in.Name, in.Address)
		}

		// close reports channel once we are done
		go func() {
			wg.Wait()
			close(reports)
		}()

		db := make(map[string]*report)
		bw := bufio.NewWriter(os.Stdout)
		for r := range reports {
			// store report in "database"
			db[r.name] = r

			// reset cursor via ANSI sequence, ignoring errors ¯\_(ツ)_/¯
			_, _ = bw.WriteString("\033[2J\033[0;0H")

			// print nicely formatted instance status
			tw := tabwriter.NewWriter(bw, 5, 4, 3, ' ', 0)
			fmt.Fprintln(tw, "NAME\tINCREMENT\tPROMISED\tID\tHOLDER\tLAST SEEN")
			for _, in := range cfgQuorum.Instances {
				status, ok := db[in.Name]
				if !ok || status.resp == nil {
					fmt.Fprintf(tw, "%v\t\t\t\t\tconnection error\n", in.Name)
					continue
				}
				fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t%v\t%v\n",
					in.Name,
					status.resp.Increment,
					status.resp.Promised,
					status.resp.ID,
					status.resp.Holder,
					humanize.Time(status.timestamp))
			}
			tw.Flush()
			bw.Flush()
			bw.Reset(os.Stdout)
		}
	},
}
