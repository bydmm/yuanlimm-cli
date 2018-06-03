package main

import (
	"crypto/sha512"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/jinzhu/now"
)

const host string = "https://www.yuanlimm.com"

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

func checkStatus() int {
	resp, err := http.Get(checkURL)
	if err != nil {
		HandleError(err)
		return 16
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var status map[string]interface{}
	if err := json.Unmarshal(body, &status); err == nil {
		hard := int(status["hard"].(float64))
		return hard
	}
	HandleCriticalError(err)
	return 0
}

func postWish(hard *int, address string, code string, lovePower int64) (bool, map[string]interface{}) {
	formData := url.Values{
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
	HandleCriticalError(err)
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

func rawOre(address string, code string) (string, int64) {
	lovePower := randNumber()
	unixTime := timestamp()
	return fmt.Sprintf("%s%d%d%s", address, lovePower, unixTime, code), lovePower
}

func hash(ore string) string {
	hashWish := sha512.Sum512([]byte(ore))
	bin := ""
	for _, n := range hashWish {
		bin = fmt.Sprintf("%s%08b", bin, n)
	}
	return bin
}

func matchWish(hard int, bin string) bool {
	reg := fmt.Sprintf("0{%d}$", hard)
	matched, err := regexp.MatchString(reg, bin)
	if err != nil {
		HandleError(err)
	}
	return matched
}

func dig(address string, code string, hard *int, count *int, writeChannel chan int) {
	for true {
		ore, lovePower := rawOre(address, code)
		bin := hash(ore)
		if matchWish(*hard, bin) {
			success, res := postWish(hard, address, code, lovePower)
			if success {
				if res["type"].(string) == "coin" {
					amount := res["amount"].(float64)
					fmt.Printf("获得援力：%0.2f\n", amount/100.0)
				}
				if res["type"].(string) == "stock" {
					amount := res["amount"].(float64)
					fmt.Printf("获得股票：%1.0f\n", amount)
				}
				// println(res["hard"])
			}
		}
		writeChannel <- 1
		*count++
		<-writeChannel
	}
}

func main() {
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
	}

	// init hard
	hard := checkStatus()
	fmt.Printf("当前股票代码: %s\n", *code)

	// dig
	writeChannel := make(chan int, 1)
	count := 0
	cost := 0
	for i := 0; i < *concurrency; i++ {
		go dig(*address, *code, &hard, &count, writeChannel)
	}

	for true {
		time.Sleep(1 * time.Second)
		hard = checkStatus()
		cost++
		fmt.Printf("当前难度%d，当前速度:%d次/秒，总计计算次数:%d\n", hard, count/cost, count)
	}
}
