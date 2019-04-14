package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"

	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Please specific url.")
		return
	}

	url := os.Args[1]

	savePath := path.Base(url)
	if len(os.Args) > 2 {
		savePath = os.Args[2]
	} else {
		r := regexp.MustCompile(`([ \w-]+\.[ \w-]+)\?`)
		matches := r.FindStringSubmatch(savePath)
		if len(matches) > 1 {
			savePath = matches[1]
		}
	}

	fmt.Printf("Download %s to %s \n", url, savePath)
	DoDownload(url, savePath, &downloader.Config{
		MaxParallel: 50,
	})
}

// DoDownload 执行下载
func DoDownload(durl string, savePath string, cfg *downloader.Config) {
	var (
		file io.WriterAt
		err  error
	)

	if savePath != "" {
		file, err = os.Create(savePath)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		file = nil
	}

	download := downloader.NewDownloader(durl, file, cfg)

	exitDownloadFunc := make(chan struct{})

	download.OnExecute(func() {
		dc := download.GetDownloadStatusChan()
		var ts string

		for {
			select {
			case v, ok := <-dc:
				if !ok { // channel 已经关闭
					return
				}

				if v.TotalSize() <= 0 {
					ts = converter.ConvertFileSize(v.Downloaded(), 2)
				} else {
					ts = converter.ConvertFileSize(v.TotalSize(), 2)
				}

				fmt.Printf("\r ↓ %s/%s %s/s in %s ............",
					converter.ConvertFileSize(v.Downloaded(), 2),
					ts,
					converter.ConvertFileSize(v.SpeedsPerSecond(), 2),
					v.TimeElapsed(),
				)
			}
		}
	})

	download.Execute()
	close(exitDownloadFunc)
}
