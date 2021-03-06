// Copyright © 2018 Matt Glaman <nmd.matt@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/mglaman/tempo/pkg/tempo"
	"github.com/mglaman/tempo/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var running = false

// timerCmd represents the timer command
var timerCmd = &cobra.Command{
	Use:   "timer",
	Short: "Create a worklog timer",
	Long:  `Creates a timer that can be converted into a worklog`,
	Run: func(cmd *cobra.Command, args []string) {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			for _ = range c {
				if running {
					running = false
				} else {
					os.Exit(1)
				}
			}
		}()

		start := time.Now()
		running = true

		fmt.Print("\x1b[?25l")
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		s.Color("white", "bold")
		s.Suffix = " Timer is running…"
		//for running == true {
		time.Sleep(time.Second)
		//}
		s.Stop()
		elapsed := time.Since(start)
		fmt.Print("\x1b[?25h")

		// Round up to 15 minutes if less than 15 minutes..
		if elapsed.Minutes() < 15 {
			elapsed = time.Duration(time.Minute * 15)
		}
		// Round off all timers to 15 minute intervals.
		elapsed = elapsed.Round(time.Minute * 15)
		fmt.Println()
		fmt.Println(fmt.Sprintf("Logging %s", elapsed))
		fmt.Println()
		issueKey := util.Prompt("Enter the issue key")
		description := util.Prompt("Worklog description")

		workLog := tempo.WorklogPayload{
			IssueKey:         issueKey,
			TimeSpentSeconds: elapsed.Seconds(),
			BillableSeconds:  elapsed.Seconds(),
			StartDate:        start.Local().Format("2006-01-02"),
			StartTime:        start.Local().Format("15:04:05"),
			Description:      description,
			AuthorAccountID:  viper.GetString("username"),
		}

		url := "https://api.tempo.io/core/3/worklogs"
		j, _ := json.Marshal(workLog)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(j))
		req.Header.Add("Authorization", "Bearer "+viper.GetString("token"))
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			panic(err.Error())
		}

		worklog := new(tempo.Worklog)
		_ = json.NewDecoder(resp.Body).Decode(worklog)

		fmt.Println("Time log submitted!")
	},
}

func init() {
	rootCmd.AddCommand(timerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// timerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// timerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
