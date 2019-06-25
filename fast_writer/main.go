package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

//HTTPClient http客户端
var HTTPClient *http.Client

func init() {
	HTTPClient = newClient(2, 100, 50)
}

//newClient 初始化http客户端
func newClient(timeout, maxIdelConns, maxConnsPerHost int) *http.Client {
	client := &http.Client{
		Timeout: time.Second * time.Duration(timeout), //设置超时时间,默认0不设置超时时间
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second, //限制建立TCP连接的时间
				KeepAlive: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second, //限制 TLS握手的时间
			ResponseHeaderTimeout: 10 * time.Second, //限制读取response header的时间
			ExpectContinueTimeout: 1 * time.Second,  //限制client在发送包含 Expect: 100-continue的header到收到继续发送body的response之间的时间等待。
			MaxIdleConns:          maxIdelConns,     //连接池对所有host的最大连接数量，默认无穷大
			MaxConnsPerHost:       maxConnsPerHost,  //连接池对每个host的最大连接数量。
			IdleConnTimeout:       10 * time.Minute, //how long an idle connection is kept in the connection pool.
		},
	}

	return client
}

func sync(log []byte, client *http.Client) int {
	url := "http://127.0.0.1:9999/api/v2/write?org=03ec56abce43b000&bucket=03ef2fd594e78000&precision=ns"
	//input := []byte(fmt.Sprintf("m v=2,v1=2,v2=2,v3=12  %v", time.Now().Unix()))

	//measurement,tag=info log="%s
	//
	lines := fmt.Sprintf(`%s,tag=info log="%s"`, time.Now().Format("200601"), log)
	input := []byte(lines)
	req, err := http.NewRequest("POST", url, bytes.NewReader(input))
	if err != nil {
		println(err)
		return 500
	}

	req.Header.Set("Authorization", " Token QytYvO7z5aG94F84tBzWLORktwK32QGmJV2YbXtukYkF-Tprr1g2Ez_ixAjOOfRvXiYdjK2boRAR-5D2hd9Syw==") //设置version
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	//req.Header.Set("Content-Encoding", "gzip")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return 500
	}

	if resp != nil {
		if resp.StatusCode != 204 {
			bytes, _ := ioutil.ReadAll(req.Body)
			fmt.Printf("[influx]=>%s.\n", bytes)
		} else {
			io.Copy(ioutil.Discard, req.Body)
		}

		defer resp.Body.Close()
	}

	return resp.StatusCode
}

func printer(total, okTotal, failTotal *int) {
	for {
		time.Sleep(time.Second * 5)
		log.Printf("[count]:%v %v %v \n", *total, *okTotal, *failTotal)
	}
}

func main() {

	//println("[server]=> i am running.")
	port := flag.Int("port", 8080, "http port")
	flag.Parse()

	client := newClient(15, 100, 100)
	count := 0
	okCount := 0
	failtCount := 0
	go printer(&count, &okCount, &failtCount)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		count++
		if r.Method == "POST" {
			bytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				println(err)
				failtCount++
				w.WriteHeader(500)
				return
			}
			if sync(bytes, client) == 204 {
				okCount++
				w.WriteHeader(204)
			} else {

				w.WriteHeader(500)
				failtCount++
				log.Printf("[error count]:%v\n", count)
			}
			return
		}
	})

	httpURL := fmt.Sprintf(":%v", *port)
	println("listen on", httpURL)
	http.ListenAndServe(httpURL, nil)
}
