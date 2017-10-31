// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	proto "github.com/golang/protobuf/proto"

	"github.com/smugcloud/mesos-fw/mesos"
	"github.com/spf13/cobra"
)

var launched = false

//Response captures the messages from Mesos master
type Response struct {
	Type       string `json:"type"`
	Subscribed struct {
		Framework struct {
			FrameworkID string `json:"value"`
		} `json:"framework_id"`
	} `json:"subscribed"`
	Offers struct {
		Offers []Offers `json:"offers"`
	} `json:"offers"`
}

//Offers holds the specific details for each Offer from the Master
type Offers struct {
	Agent struct {
		AgentID string `json:"value"`
	} `json:"agent_id"`
	Framework struct {
		FrameworkID string `json:"value"`
	} `json:"framework_id"`
	Offer struct {
		OfferID string `json:"value"`
	} `json:"id"`
}

var frameworkID string
var image string
var test string

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a Docker image.",
	Run: func(cmd *cobra.Command, args []string) {

		const (
			host = "localhost:5050/api/v1/scheduler"
		)

		var vType mesos.SubRequest_Type = 1
		user := "Nick"
		name := "Go Framework"
		payload := mesos.SubRequest{
			Type: &vType,
			Subscribe: &mesos.SubRequest_Subscribe{
				FrameworkInfo: &mesos.FrameworkInfo{
					User: &user,
					Name: &name,
				},
			},
		}

		data, err := proto.Marshal(&payload)
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := http.Post("http://localhost:5050/api/v1/scheduler", "application/x-protobuf", bytes.NewBuffer(data))

		//Capture the Mesos-Stream-Id
		mesosStringID := resp.Header["Mesos-Stream-Id"]
		log.Print("Setting Mesos-Stream-ID: ", mesosStringID[0])

		if err != nil {
			fmt.Println(err)
			return
		}
		reader := bufio.NewReader(resp.Body)
		var y int
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			line = bytes.TrimSpace(line)
			x, _ := strconv.Atoi(string(line))
			if x > 0 {
				y = x
				continue
			}

			//Print message
			log.Printf("%s\n", line[0:y])

			//Print size of next message
			//log.Printf("Next Message is %s characters. ", string(line[y:]))

			var s = new(Response)

			err = json.Unmarshal(line[0:y], &s)
			if err != nil {
				fmt.Println("Error Unmarshalling: ", err)
			}

			log.Print("Response struct: ", s)

			if s.Type == "SUBSCRIBED" {
				frameworkID = s.Subscribed.Framework.FrameworkID

				fmt.Printf("Got framework ID: %s\n", frameworkID)

				//Need to make sure y gets updated so we don't run out of bounds.
				y, _ = strconv.Atoi(string(line[y:]))
				continue

			} else if s.Type == "OFFERS" {
				if launched == true {
					y, _ = strconv.Atoi(string(line[y:]))

					continue
				}
				log.Print("Attempting to launch container.")
				var typeType mesos.Call_Type = 3
				var opType mesos.Offer_Operation_Type = 1
				var containerInfoType mesos.ContainerInfo_Type = 2
				var imageType mesos.Image_Type = 2

				offerID := s.Offers.Offers[0].Offer.OfferID
				log.Print("OfferID: ", offerID)
				taskID := "some-id"
				agentID := s.Offers.Offers[0].Agent.AgentID
				log.Print("agentID: ", agentID)
				taskName := "go-docker"
				imageName := image
				useShell := false
				cpu := "cpus"
				mem := "mem"
				var memResources float64 = 64
				cpuResources := .5
				var resourceType mesos.Value_Type
				container := mesos.Call{
					FrameworkId: &mesos.FrameworkID{
						Value: &frameworkID,
					},
					Type: &typeType,
					Accept: &mesos.Call_Accept{
						OfferIds: []*mesos.OfferID{
							{Value: &offerID},
						},
						Operations: []*mesos.Offer_Operation{
							{Type: &opType,
								Launch: &mesos.Offer_Operation_Launch{
									TaskInfos: []*mesos.TaskInfo{
										{Name: &taskName,
											TaskId: &mesos.TaskID{
												Value: &taskID},
											AgentId: &mesos.AgentID{
												Value: &agentID},
											Command: &mesos.CommandInfo{
												Shell: &useShell,
											},
											Container: &mesos.ContainerInfo{
												Type: &containerInfoType,
												Mesos: &mesos.ContainerInfo_MesosInfo{
													Image: &mesos.Image{
														Type: &imageType,
														Docker: &mesos.Image_Docker{
															Name: &imageName,
														},
													},
												},
											},
											Resources: []*mesos.Resource{
												{Name: &cpu,
													Type: &resourceType,
													Scalar: &mesos.Value_Scalar{
														Value: &cpuResources,
													},
												},
												{Name: &mem,
													Type: &resourceType,
													Scalar: &mesos.Value_Scalar{
														Value: &memResources,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
				dbug, _ := json.Marshal(container)
				fmt.Printf("Marshalled container object: %s\n", string(dbug))

				LaunchContainer(container, mesosStringID[0])
				// var test = []mesos.OfferID{
				// 	{Value: &first},
				// 	{Value: &second},
				// }
				y, _ = strconv.Atoi(string(line[y:]))

				continue

			}

			//Need to set after processing
			y, _ = strconv.Atoi(string(line[y:]))

		}

	},
}

//LaunchContainer posts to the scheduler with a request to run a Docker container
func LaunchContainer(container mesos.Call, msid string) {
	//log.Print("Using MSID: ", msid)
	data, err := proto.Marshal(&container)
	if err != nil {
		log.Fatal(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:5050/api/v1/scheduler", bytes.NewReader(data))
	req.Header.Set("Mesos-Stream-Id", msid)
	req.Header.Set("Content-Type", "application/x-protobuf")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	launched = true
	log.Printf("Full response: %v\n", resp)

}

func init() {
	RootCmd.AddCommand(launchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// launchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// launchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	launchCmd.Flags().StringVar(&image, "image", "smugcloud/echo-server", "The fully qualified Docker image you would like to use.")
}
