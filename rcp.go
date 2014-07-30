package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
)

type RPCClient struct {
	server  string
	port    int
	address string
}

func NewRPCClient(server string, port int) *RPCClient {
	address := "http://" + server + ":" + fmt.Sprintf("%d", port)
	client := RPCClient{
		server:  server,
		port:    port,
		address: address,
	}
	return &client
}

func (client *RPCClient) Prepare(method string) func(params ...interface{}) (interface{}, error) {
	prep := func(args ...interface{}) (interface{}, error) {
		data, err := json.Marshal(map[string]interface{}{
			"method": method,
			"id":     0,
			"params": args,
		})

		if err != nil {
			log.Println("Marshal: %v", err)
			return nil, err
		}

		resp, err := http.Post(client.address, "application/json", strings.NewReader(string(data)))
		if err != nil {
			log.Println("Post: %v", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("ReadAll: %v", err)
			return nil, err
		}

		result := make(map[string]interface{})
		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Println("Unmarshal: %v", err)
			return nil, err
		}

		if e := result["error"]; e != nil {
			err := errors.New(e.(map[string]interface{})["message"])
			return result["result"], err
		}

		return result["result"], nil
	}
	return prep
}

func main() {
	runtime.GOMAXPROCS(16)

	nlp := NewRPCClient("localhost", 8080)
	parse := nlp.Prepare("parse")

	for i := 0; i < 10000; i++ {
		parse("Commonwealth Games: More buses after lengthy waits: EXTRA BUSES are being put on after spectators endured lengthy... http://dlvr.it/6S34yk")
		if i%100 == 0 {
			fmt.Println(i)
		}
	}

}
