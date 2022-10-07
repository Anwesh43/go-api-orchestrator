package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type RequestObject struct {
	headers       map[string]string
	url           string
	childRequests []*RequestObject
}

func makeCall(req *RequestObject, ch chan string, cb func()) {
	defer cb()
	respArr := make([]string, 0)
	resp, err := http.Get(req.url)
	if err != nil {
		ch <- "error while doing fetch call"
		cb()
	}
	defer resp.Body.Close()
	d, derr := ioutil.ReadAll(resp.Body)
	if derr != nil {
		ch <- "error while parsing"
		cb()
	}
	//fmt.Println("Response", string(d))
	respArr = append(respArr, string(d))

	currCh := make(chan string, len(req.childRequests))
	var wg1 sync.WaitGroup
	wg1.Add(len(req.childRequests))
	for _, childReq := range req.childRequests {
		//currCh := make(chan string)
		go makeCall(childReq, currCh, func() { wg1.Done() })
	}
	wg1.Wait()
	close(currCh)
	if len(req.childRequests) > 0 {
		respArr = append(respArr, "(")
	}
	for i := 0; i < len(req.childRequests); i++ {
		str1 := <-currCh
		respArr = append(respArr, str1)
	}
	if len(req.childRequests) > 0 {
		respArr = append(respArr, ")")
	}
	ch <- strings.Join(respArr, "")

}

func addToWaitGroup(root *RequestObject) int {
	size := len(root.childRequests)
	n := 0
	for _, req := range root.childRequests {
		n += addToWaitGroup(req)
	}
	return n + size
}

func main() {
	root := RequestObject{
		headers:       make(map[string]string),
		url:           "http://localhost:3000/books",
		childRequests: make([]*RequestObject, 0),
	}

	root.childRequests = append(root.childRequests, &RequestObject{
		headers:       make(map[string]string),
		url:           "http://localhost:3000/pets",
		childRequests: make([]*RequestObject, 0),
	})

	root.childRequests = append(root.childRequests, &RequestObject{
		headers:       make(map[string]string),
		url:           "http://localhost:3000/person",
		childRequests: make([]*RequestObject, 0),
	})

	rootChan := make(chan string)

	go makeCall(&root, rootChan, func() {
		fmt.Println("ending responses")
	})
	fmt.Println("started")
	respStr := <-rootChan
	fmt.Println("Response String", respStr)
	fmt.Println("ended")

}
