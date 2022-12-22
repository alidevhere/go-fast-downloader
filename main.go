package main

import (
	"fmt"
	"time"

	"github.com/alidevhere/go-fast-downloader/downloader"
)

func main() {
	start := time.Now()

	downloader.Download(downloader.DownloadOptions{
		NoOfWorkers: 30,
		NoOfChunks:  500,
		URI:         `https://rr3---sn-p5qlsndz.googlevideo.com/videoplayback?expire=1671740798&ei=HWmkY6meO8bligTnh4xo&ip=205.185.214.122&id=o-ALMnPBvULu_FY-KEVoDPaTU1911NKYVjACZRrW2zphsc&itag=18&source=youtube&requiressl=yes&spc=zIddbPJiJVDq56R-reOIn3lBtS7x8Is&vprv=1&mime=video%2Fmp4&ns=7bwTq6FneH4RBc08BJoCCgQK&cnr=14&ratebypass=yes&dur=814.950&lmt=1657994571026915&fexp=24001373,24007246&c=WEB&txp=2318224&n=M3NaUOr57SQfXw&sparams=expire%2Cei%2Cip%2Cid%2Citag%2Csource%2Crequiressl%2Cspc%2Cvprv%2Cmime%2Cns%2Ccnr%2Cratebypass%2Cdur%2Clmt&sig=AOq0QJ8wRgIhAMy5L_xC1U_G4sT9esycbMoIVD2V1FUCu4qJ-eDwqIBYAiEArhq_Vv7K-Dx4LFixtCEB_FY1L2IJMLMd8xOsOcgGGfU%3D&redirect_counter=1&rm=sn-q4fel77e&req_id=319ad25dab3da3ee&cms_redirect=yes&cmsv=e&ipbypass=yes&mh=Z3&mip=119.160.59.102&mm=31&mn=sn-p5qlsndz&ms=au&mt=1671720799&mv=D&mvi=3&pl=0&lsparams=ipbypass,mh,mip,mm,mn,ms,mv,mvi,pl&lsig=AG3C_xAwRgIhAP80cXpw_YPPh91to4jJrN7muvKYRR7z6T1ZNkmmSDXRAiEAhzWK0fatdUYcE5ZGt0IubE8aTINshkQvGZ1RDOw-46Q%3D`,
	})
	fmt.Printf("Completed Download in %s", time.Since(start).String())

}
