package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/smugcloud/mesos-fw/mesos"
)

const (
	host = "localhost:5050/api/v1/scheduler"
)

func main() {

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

	if err != nil {
		fmt.Println(err)
		return
	}
	log.Print(resp.Header["Mesos-Stream-Id"])

}
