package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// 定义全局变量
var (
	debug   bool
	target  string
	list    string
	jsonOut bool
	retries int
)

// 定义结果结构体
type Result struct {
	Input      string `json:"input"`
	QueryURL   string `json:"query_url"`
	License    string `json:"license"`
	VerifyTime string `json:"verify_time"`
	ComName    string `json:"com_name"`
	Typ        string `json:"typ"`
	Permit     string `json:"permit"`
	Host       string `json:"host"`
}

// 初始化函数，用于解析命令行参数
func init() {
	flag.BoolVar(&debug, "debug", false, "启用调试模式")
	flag.StringVar(&target, "t", "", "目标域名")
	flag.StringVar(&list, "l", "", "包含域名列表的文件")
	flag.BoolVar(&jsonOut, "json", false, "以 JSON 格式输出结果")
	flag.IntVar(&retries, "r", 3, "重试次数")
	flag.Parse()
}

// 主函数
func main() {
	// 检查是否提供了 -t 或 -l 参数，或者是否有标准输入
	if target == "" && list == "" && isStdinEmpty() {
		fmt.Printf("使用方法: %s [-t <target.xyz>] [-l <lists.txt>]\n", os.Args[0])
		os.Exit(1)
	}

	// 创建 chromedp 上下文
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-web-security", false),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 根据输入参数获取域名列表
	domains := getDomains()

	// 遍历域名列表，处理每个域名
	for _, domain := range domains {
		processDomainWithRetry(ctx, domain)
	}
}

// 获取域名列表
func getDomains() []string {
	var domains []string
	if target != "" {
		domains = append(domains, target)
	} else if list != "" {
		domains = readDomainsFromFile(list)
	} else {
		domains = readDomainsFromStdin()
	}
	return domains
}

// 从文件中读取域名列表
func readDomainsFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domains = append(domains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return domains
}

// 从标准输入读取域名列表
func readDomainsFromStdin() []string {
	var domains []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		domains = append(domains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return domains
}

// 处理单个域名，带有重试机制
func processDomainWithRetry(ctx context.Context, domain string) {
	for i := 0; i < retries; i++ {
		err := processDomain(ctx, domain)
		if err != nil {
			log.Println("处理域名时出错:", err)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}
}

// 处理单个域名
func processDomain(ctx context.Context, domain string) error {
	if debug {
		fmt.Println("Starting the program...")
	}

	// 访问网站
	if debug {
		fmt.Println("Navigating to the website...")
	}
	err := chromedp.Run(ctx, chromedp.Navigate(`https://icp.chinaz.com/`+domain))
	if err != nil {
		return err
	}

	// 获取并打印页面源码
	if debug {
		fmt.Println("Getting the page source...")
	}
	var source string
	err = chromedp.Run(ctx, chromedp.OuterHTML("html", &source))
	if err != nil {
		return err
	}
	if debug {
		fmt.Println(source)
	}

	// 提取 "/home/info?host=YW1lbXYuY29t" 字段
	if debug {
		fmt.Println("Extracting the href attribute...")
	}
	re := regexp.MustCompile(`href="(/home/info[^"]+)"`)
	matches := re.FindStringSubmatch(source)
	if len(matches) < 2 {
		return fmt.Errorf("No match found")
	}
	href := matches[1]

	// 解析相对 URL 为绝对 URL
	u, err := url.Parse(`https://icp.chinaz.com`)
	if err != nil {
		return err
	}
	hrefURL, err := u.Parse(href)
	if err != nil {
		return err
	}

	// 再次访问网站
	if debug {
		fmt.Println("Navigating to the second website...")
	}
	err = chromedp.Run(ctx, chromedp.Navigate(hrefURL.String()))
	if err != nil {
		return err
	}

	// 等待元素加载
	if debug {
		fmt.Println("Waiting for elements to load...")
	}
	err = chromedp.Run(ctx, chromedp.Sleep(2*time.Second))
	if err != nil {
		return err
	}

	// 提取主体信息和 ICP 信息
	if debug {
		fmt.Println("Extracting information...")
	}
	var license, verifyTime, comName, typ, permit, host string
	err = chromedp.Run(ctx,
		chromedp.Text(`td#license`, &license),
		chromedp.Text(`td#verifyTime`, &verifyTime),
		chromedp.Text(`td#comName`, &comName),
		chromedp.Text(`td#typ`, &typ),
		chromedp.Text(`td#permit`, &permit),
		chromedp.Text(`td#host`, &host),
	)
	if err != nil {
		return err
	}

	// 输出结果
	result := Result{
		Input:      domain,
		QueryURL:   `https://icp.chinaz.com/` + domain,
		License:    strings.TrimSpace(license),
		VerifyTime: strings.TrimSpace(verifyTime),
		ComName:    strings.TrimSpace(comName),
		Typ:        strings.TrimSpace(typ),
		Permit:     strings.TrimSpace(permit),
		Host:       strings.TrimSpace(host),
	}

	if jsonOut {
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonResult))
	} else {
		fmt.Println("Input:", result.Input)
		fmt.Println("Query URL:", result.QueryURL)
		fmt.Println("ICP备案/许可证号:", result.License)
		fmt.Println("审核通过日期:", result.VerifyTime)
		fmt.Println("主办单位名称:", result.ComName)
		fmt.Println("主办单位性质:", result.Typ)
		fmt.Println("ICP备案/许可证号:", result.Permit)
		fmt.Println("网站域名:", result.Host)
	}

	if debug {
		fmt.Println("Program finished.")
	}

	return nil
}

// 检查标准输入是否为空
func isStdinEmpty() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
