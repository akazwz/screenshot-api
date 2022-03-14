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

type ScreenConfig struct {
	URL string `json:"url" form:"url" binding:"required"`
}

type ScreenshotRes struct {
	Base64 string `json:"base64"`
}

func GetScreenShotByScreenConfig(config ScreenConfig) (err error, imageBase64 string) {
	_ = []chromedp.ExecAllocatorOption{
		chromedp.ExecPath("/headless-shell/headless-shell"),
	}

	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://browser:9222/")

	ctx, cancel := chromedp.NewContext(
		allocCtx,
	)
	defer cancel()

	var buf []byte

	err = chromedp.Run(ctx, fullScreenShot(config, &buf))
	if err != nil {
		log.Println("full screenshot error")
		return
	}
	imageBase64 = base64.StdEncoding.EncodeToString(buf)
	//err = ioutil.WriteFile("fullScreenshot.png", buf, 0o644)
	return
}

func fullScreenShot(config ScreenConfig, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.EmulateViewport(1280, 800),
		chromedp.Navigate(config.URL),
		chromedp.FullScreenshot(res, 70),
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
