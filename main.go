package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/jinzhu/now"
)

const host string = "http://yuanlimm-server.yuanlimm.com"

var wishURL = fmt.Sprintf("%s/api/super_wishs", host)
var checkURL = fmt.Sprintf("%s/api/super_wishs", host)

// HandleError 处理普通错误
func HandleError(err error) {
	fmt.Println("occurred error:", err)
}

// HandleCriticalError 处理致命错误
func HandleCriticalError(err error) {
	fmt.Println("occurred error:", err)
	os.Exit(-1)
}

func checkStatus() (int, int64) {
	resp, err := http.Get(checkURL)
	if err != nil {
		HandleError(err)
		return 0, 16
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var status map[string]interface{}
	if err := json.Unmarshal(body, &status); err == nil {
		hard := int(status["hard"].(float64))
		unixTime := int64(status["unix_time"].(float64))
		return hard, unixTime
	}
	HandleError(err)
	return 0, 16
}

func postWish(hard *int, cheerWord string, address string, code string, lovePower int64) (bool, map[string]interface{}) {
	formData := url.Values{
		"cheer_word": {cheerWord},
		"address":    {address},
		"code":       {code},
		"love_power": {fmt.Sprintf("%d", lovePower)},
	}
	resp, err := http.PostForm(wishURL, formData)
	if err != nil {
		HandleError(err)
		return false, map[string]interface{}{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err == nil {
		fmt.Println(res)
		*hard = int(res["hard"].(float64))
		success := res["success"].(bool)
		return success, res
	}
	HandleError(err)
	return false, res
}

// 时间戳
func timestamp() int64 {
	return now.BeginningOfMinute().Unix()
}

// 随机数
func randNumber() int64 {
	return rand.Int63()
}

func rawOre(unixtime *int64, cheerWord string, address string, code string) ([]byte, int64) {
	lovePower := randNumber()
	ore := bytes.Join([][]byte{
		[]byte(cheerWord),
		[]byte(address),
		[]byte(strconv.FormatInt(lovePower, 10)),
		[]byte(strconv.FormatInt(*unixtime, 10)),
		[]byte(code),
	}, []byte{})
	return ore, lovePower
}

func hash(ore []byte) [64]byte {
	return sha512.Sum512(ore)
}

func oldHash(ore string) string {
	hashWish := sha512.Sum512([]byte(ore))
	bin := ""
	for _, n := range hashWish {
		bin = fmt.Sprintf("%s%08b", bin, n)
	}
	return bin
}

func matchWish(hard int, ore []byte) bool {
	bin := hash(ore)
	zero := (hard / 8)
	for index := 1; index <= zero; index++ {
		if bin[len(bin)-index] != 0 {
			return false
		}
	}

	residual := (hard % 8)

	if residual > 0 {
		last := bin[len(bin)-(hard/8)-1]
		head := fmt.Sprintf("%08b", last)

		if len(head) < residual {
			return false
		}

		headZero := ""
		for index := 0; index < residual; index++ {
			headZero += "0"
		}

		// if head[len(head)-residual:] == headZero {
		// 	nbin := ""
		// 	for _, n := range bin {
		// 		fmt.Printf("%08b ", n)
		// 		nbin = fmt.Sprintf("%s%08b", nbin, n)
		// 	}
		// 	fmt.Println()
		// 	fmt.Println(bin)
		// 	fmt.Println(head)
		// 	fmt.Println(last)
		// 	fmt.Println(hard)
		// }

		return head[len(head)-residual:] == headZero
	}
	return true
}

func dig(unixtime *int64, cheerWord string, address string, code string, hard *int, count *int) {
	for true {
		ore, lovePower := rawOre(unixtime, cheerWord, address, code)
		if matchWish(*hard, ore) {
			// fmt.Println(oldHash(ore))
			success, res := postWish(hard, cheerWord, address, code, lovePower)
			if success {
				if res["type"].(string) == "coin" {
					amount := res["amount"].(float64)
					fmt.Printf("获得援力：%0.2f\n", amount/100.0)
				}
				if res["type"].(string) == "stock" {
					amount := res["amount"].(float64)
					fmt.Printf("获得股票：%1.0f\n", amount)
				}
			}
		}
		*count++
	}
}

func main() {
	cheerWord := flag.String("w", "", "应援词，神秘玄学")
	address := flag.String("a", "", "钱包地址，请不要泄露")
	code := flag.String("code", "", "股票代码")
	concurrency := flag.Int("c", 0, "并发数,默认为1, 不建议超过CPU数")
	flag.Parse()

	if *address == "" {
		var inputAddress string
		fmt.Println("请输入你的钱包地址: ")
		fmt.Scanln(&inputAddress)
		if inputAddress == "" {
			fmt.Println("无法获取地址，详情咨询群774800449")
			os.Exit(-1)
		}
		address = &inputAddress
	}

	if *code == "" {
		var inputCode string
		fmt.Println("请输入股票代码: ")
		fmt.Scanln(&inputCode)
		if inputCode == "" {
			fmt.Println("无法获取股票代码，详情咨询群774800449")
			os.Exit(-1)
		}
		code = &inputCode
	}

	if *concurrency == 0 {
		var inputConcurrency int
		fmt.Println("并发数，建议与CPU核心相同，默认为1: ")
		fmt.Scanln(&inputConcurrency)
		if inputConcurrency == 0 {
			inputConcurrency = 1
		}
		concurrency = &inputConcurrency
		// 最大并发
		runtime.GOMAXPROCS(*concurrency)
	}

	if *cheerWord == "" {
		var inputCheerWord string
		fmt.Println("你想对他/她说?")
		fmt.Scanln(&inputCheerWord)
		cheerWord = &inputCheerWord
	}

	// init hard
	hard, unixTime := checkStatus()
	fmt.Printf("当前股票代码: %s\n", *code)

	// dig
	// writeChannel := make(chan int, 1)
	count := 0
	cost := 0
	for i := 0; i < *concurrency; i++ {
		go dig(&unixTime, *cheerWord, *address, *code, &hard, &count)
	}

	for true {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(1000 * time.Millisecond)
		hard, unixTime = checkStatus()
		cost++
		fmt.Printf("当前难度%d，当前速度:%d次/秒，总计计算次数:%d, Go: %d\n",
			hard, count/cost, count, runtime.NumGoroutine())
	}
}
