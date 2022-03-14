package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ScreenConfig  网页截屏配置
type ScreenConfig struct {
	URL     string `json:"url" form:"url" binding:"required"`
	Width   int64  `json:"width" form:"width"`
	Height  int64  `json:"height" form:"height"`
	Full    int64  `json:"full" form:"full"`
	Quality int64  `json:"quality" form:"quality"`
}

type ScreenshotRes struct {
	Base64 string `json:"base64"`
}

func GetScreenShotByScreenConfig(config ScreenConfig) (err error, imageBase64 string) {
	var allocCtx context.Context
	var cancel context.CancelFunc

	if gin.Mode() == gin.ReleaseMode {
		/* release */
		allocCtx, cancel = chromedp.NewRemoteAllocator(context.Background(), "ws://browser:9222/")
	} else {
		/* dev */
		allocCtx, cancel = chromedp.NewContext(context.Background())
	}

	defer cancel()

	ctx, cancel := chromedp.NewContext(
		allocCtx,
	)
	defer cancel()

	var buf []byte
	err = chromedp.Run(ctx, screenShot(config, &buf))
	if err != nil {
		log.Println("full screenshot error")
		return
	}
	imageBase64 = base64.StdEncoding.EncodeToString(buf)
	//err = ioutil.WriteFile("fullScreenshot.png", buf, 0o644)
	return
}

func screenShot(config ScreenConfig, res *[]byte) chromedp.Tasks {
	/* 默认宽度 */
	width := int64(1280)
	if config.Width > 0 {
		width = config.Width
	}
	/* 默认高度 */
	height := int64(800)
	if config.Height > 0 {
		height = config.Height
	}

	navigate := chromedp.Tasks{
		chromedp.EmulateViewport(width, height),
		chromedp.Navigate(config.URL),
	}

	if config.Full == 1 {
		return chromedp.Tasks{
			navigate,
			chromedp.FullScreenshot(res, int(config.Quality)),
		}
	}
	return chromedp.Tasks{
		navigate,
		chromedp.CaptureScreenshot(res),
	}
}

func main() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))

	r.GET("screenshot", func(c *gin.Context) {
		var config ScreenConfig
		/* 参数错误 */
		if c.ShouldBindQuery(&config) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "参数错误",
			})
			return
		}

		err, imageBase64 := GetScreenShotByScreenConfig(config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "截图失败",
			})
			return
		}
		c.JSON(http.StatusOK, ScreenshotRes{Base64: imageBase64})
		return
	})

	err := r.Run("0.0.0.0:8000")
	if err != nil {
		log.Fatalln("run error")
	}
}
