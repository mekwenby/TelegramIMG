package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/levigross/grequests"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func download(url, imgsavepath string) error {
	// 发送 GET 请求获取图像数据
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载图像失败：%w", err)
	}
	defer resp.Body.Close()

	// 读取图像数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取图像数据失败：%w", err)
	}

	// 保存图像数据到文件
	if err := ioutil.WriteFile(imgsavepath, body, 0644); err != nil {
		return fmt.Errorf("保存图像文件失败：%w", err)
	}
	defer wg.Done()
	return nil
}

func main() {
	// 创建源码存放文件夹
	htmlDir := "html"
	if _, err := os.Stat(htmlDir); os.IsNotExist(err) {
		// 文件夹不存在，创建文件夹
		err := os.MkdirAll(htmlDir, 0755)
		if err != nil {
			log.Printf("html文件夹创建失败 %v\n", err)
		}
	}
	// 创建图片下载文件夹
	downloadDir := "Download"
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		// 文件夹不存在，创建文件夹
		err := os.MkdirAll(downloadDir, 0755)
		if err != nil {
			log.Printf("Download文件夹创建失败 %v\n", err)
		}
	}
	commandUrl := flag.String("url", "None", "启动配置文件")
	flag.Parse()
	// commandUrl字符串保存到url
	url := *commandUrl

	if url == "None" {
		log.Println("url 为传递参数,程序退出!")
		os.Exit(1)
	}

	host := "https://telegra.ph"
	// 创建一个包含最新Chrome浏览器User-Agent的RequestOptions结构体
	ro := &grequests.RequestOptions{
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
		},
	}

	// 使用RequestOptions发起GET请求
	resp, err := grequests.Get(url, ro)
	if err != nil {
		panic(err)
	}

	// 使用goquery.NewDocumentFromReader解析HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		log.Fatal(err)
	}

	// 提取标题
	title := doc.Find("title").Text()
	fmt.Println("Title:", title)
	htmlContent := resp.String()
	htmlName := title + ".html"
	htmlFilePath := path.Join(htmlDir, htmlName)
	// 保存html源码
	err = ioutil.WriteFile(htmlFilePath, []byte(htmlContent), 0644)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("figure").Each(func(index int, item *goquery.Selection) {
		img := item.Find("img")
		if img.Length() > 0 {
			imgSrc, exists := img.Attr("src")
			if exists {
				// 图片连接
				imgUrl := host + imgSrc
				imgFileName := strconv.Itoa(index) + "_" + path.Base(imgUrl)
				savePath := path.Join(downloadDir, title)
				// 图片保存文件夹不存在，创建文件夹
				if _, err := os.Stat(savePath); os.IsNotExist(err) {
					err := os.MkdirAll(savePath, 0755)
					if err != nil {
						log.Printf("%v文件夹创建失败 %v\n", savePath, err)
					}
				}
				// 图片报错路径
				imgSavePath := path.Join(downloadDir, title, imgFileName)
				fmt.Printf("开始下载#%d: %v%s\n", index, host, imgUrl)
				wg.Add(1)
				go download(imgUrl, imgSavePath)

			}
		}
	})
	// 等待所有下载任务完成
	wg.Wait()
	log.Println("Done!")

}
