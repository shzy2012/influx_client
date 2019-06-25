package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//HTTPClient http客户端
var HTTPClient *http.Client

func init() {
	HTTPClient = newClient(5, 100, 50)
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

/*
from(bucket:\"LaiyeBucket\")
	|>range(start:1559122200, stop: 1559123054)
	|>filter(fn:(r)=>r._measurement==\"2019052917\" )
	|>limit(n:100, offset: 0)"
*/

func formatFlux(db, table, start, stop string, offset int) string {
	flux := fmt.Sprintf(`from(bucket: "longfor")
	|> range(start:-1h)
	|> filter(fn: (r) => r._measurement == "201906")
	|> keep(columns: ["tag","_value"])
	|> limit(n:600,offset:%v)`, offset)

	log.Printf("[sql]=>%s", flux)
	flux = strings.Replace(flux, "\n\t", "", -1)   //处理换行
	flux = strings.Replace(flux, "\"", "\\\"", -1) //处理转义
	return flux
}

//RespData 获取数据
type RespData struct {
	Result string `json:"result"`
	Table  int    `json:"table"`
	Value  string `json:"value"`
	Tag    string `json:"tag"`
}

func sync(start, stop string, offset int, client *http.Client) []byte {
	url := "http://39.96.21.121:9999/api/v2/query?orgID=03ec56abce43b000&pretty=true&chunked=true"
	flux := formatFlux("longfor", "20190610", start, stop, offset)
	input := []byte(`{"query":"` + flux + `","type":"flux"}`)
	fmt.Printf("[sql]=>%s", input)
	req, err := http.NewRequest("POST", url, bytes.NewReader(input))

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", " Token QytYvO7z5aG94F84tBzWLORktwK32QGmJV2YbXtukYkF-Tprr1g2Ez_ixAjOOfRvXiYdjK2boRAR-5D2hd9Syw==") //设置version

	resp, err := HTTPClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return *new([]byte)
	}

	if resp != nil {
		//bytes, _ := ioutil.ReadAll(resp.Body)
		//csvFile := csv.NewReader(resp.Body)
		reader := csv.NewReader(resp.Body)
		reader.FieldsPerRecord = -1

		csvData, err := reader.ReadAll()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var emp RespData
		var employees []RespData
		//fmt.Printf("%v\n", len(csvData))
		//fmt.Printf("%s\n", csvData)
		if len(csvData) > 2 {
			for i := 1; i < len(csvData); i++ {
				each := csvData[i]

				emp.Result = each[0]
				emp.Table, _ = strconv.Atoi(each[1])
				emp.Tag = each[2]
				emp.Value = each[3]
				employees = append(employees, emp)
			}
		} else {
			employees = []RespData{}
		}

		// Convert to JSON

		bytes, err := json.Marshal(employees)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		//fmt.Printf("%s\n", bytes)
		defer resp.Body.Close()
		return bytes
	}
	return *new([]byte)
}

func sync2(offset int, client *http.Client) []byte {
	url := "http://127.0.0.1:9999/api/v2/query?orgID=03ec56abce43b000&pretty=true&chunked=true"
	flux := formatFlux("longfor", "20190610", "", "", offset)
	input := []byte(`{"query":"` + flux + `","type":"flux"}`)
	//fmt.Printf("[sql]=>%s", input)
	req, err := http.NewRequest("POST", url, bytes.NewReader(input))

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", " Token QytYvO7z5aG94F84tBzWLORktwK32QGmJV2YbXtukYkF-Tprr1g2Ez_ixAjOOfRvXiYdjK2boRAR-5D2hd9Syw==") //设置version

	resp, err := HTTPClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return *new([]byte)
	}

	if resp != nil {
		reader := csv.NewReader(resp.Body)
		reader.FieldsPerRecord = -1

		csvData, err := reader.ReadAll()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var lines []string
		if len(csvData) > 2 {
			for i := 1; i < len(csvData); i++ {
				each := csvData[i]
				lines = append(lines, each[3])
			}
		}

		defer resp.Body.Close()
		return []byte(strings.Join(lines, ""))
	}
	return *new([]byte)
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	//(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS ,GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func main() {

	//println("[server]=> i am running.")
	port := flag.Int("port", 6001, "http port")
	flag.Parse()

	client := newClient(15, 100, 100)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		setupResponse(&w, r)
		if r.Method != "GET" {
			w.Write([]byte("not allow"))
		}

		fmt.Printf("[request]=>%s,%s", r.Host, r.URL)
		query := r.URL.Query()
		if len(query) <= 0 {
			query.Add("n", "1")
		}

		offset, err := strconv.Atoi(query["n"][0])
		if err != nil {
			offset = 1
		}
		//start := "2019-06-23T13:30:00Z" //query["start"][0]
		//stop := "2019-06-24T14:30:00Z"  //query["stop"][0]

		w.Write(sync2(offset, client))
	})

	httpURL := fmt.Sprintf(":%v", *port)
	println("listen on", httpURL)
	http.ListenAndServe(httpURL, nil)
}
