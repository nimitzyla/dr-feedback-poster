package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/go-resty/resty/v2"
	"github.com/skip2/go-qrcode"
)

const WIDTH = 2895 / 2
const HEIGHT = 4096 / 2
const URL = "https://api.whatsapp.com/send?phone=919818111918&text="
const MESSAGE = "I give consent to Zyla to call me back and explain the Diet and Nutrition counseling program that my HCP %s (%s) has recommended to enroll."

const Limit = 5

var client = resty.New()

type Doc struct {
	name        string
	code        string
	designation string
}

func drawPoster(d Doc, s3client *s3.Client) {

	dc := gg.NewContext(WIDTH, HEIGHT) // canvas 1000px by 1000px

	bgImage, err := gg.LoadImage("template.jpg")
	if err != nil {
		fmt.Println(err)
	}
	dc.DrawImage(bgImage, 0, 0)
	textColor := color.Black
	if err := dc.LoadFontFace("pn-bold.ttf", 38); err != nil {
		fmt.Println(err)
	}
	r, g, b, _ := textColor.RGBA()
	mutedColor := color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(255),
	}
	dc.SetColor(mutedColor)
	s := d.name
	lineHeight := 1.5
	_, textHeight := dc.MeasureMultilineString(s, lineHeight)
	x := 450.0
	//y := float64(dc.Height()) - textHeight - 1104
	y := 1100.0
	dc.DrawStringWrapped(s, x, y, 1, 0, 300, lineHeight, gg.AlignCenter)
	if err := dc.LoadFontFace("pn.ttf", 28); err != nil {
		fmt.Println(err)
	}
	//dc.DrawStringWrapped(d.designation, x, y+textHeight+48, 1, 0, 276, 1.4, gg.AlignCenter)
	dc.DrawStringWrapped(d.designation, x+20, y+textHeight+48, 1, 0, 320, 1.4, gg.AlignCenter)
	dc.SetColor(color.RGBA{
		R: uint8(255),
		G: uint8(255),
		B: uint8(255),
		A: uint8(255),
	})
	dc.DrawString(d.name+" and Astrazeneca", 618, 1874)
	if err := dc.LoadFontFace("pn-bold.ttf", 64); err != nil {
		fmt.Println(err)
	}
	primaryColor := color.RGBA{
		R: uint8(128),
		G: uint8(g),
		B: uint8(83),
		A: uint8(255),
	}
	dc.SetColor(primaryColor)
	dc.DrawString(d.code, 754, 1100)

	qrString := URL + url.QueryEscape(fmt.Sprintf(MESSAGE, d.name, d.code))
	//err = qrcode.WriteFile(qrString, qrcode.Medium, 310, "images/qr-"+d.code+".png")
	err = qrcode.WriteFile(qrString, qrcode.Medium, 342, "images/qr-"+d.code+".png")
	qrImage, err := gg.LoadImage("images/qr-" + d.code + ".png")
	dc.DrawImage(qrImage, 1020, 342)

	url := fmt.Sprintf("https://az-doc.s3.ap-south-1.amazonaws.com/%s.jpg", d.code)
	// fmt.Println(url)
	fmt.Println("Generating for ", d)
	resp, err := client.R().Get(url)
	if resp.IsSuccess() {
		ioutil.WriteFile("images/"+d.code+".jpg", resp.Body(), 0666)
		docImage, err := gg.LoadImage("images/" + d.code + ".jpg")
		if err != nil {
			fmt.Errorf("Couldn't load image for %s", d.code)
		} else {
			docImageResized := imaging.Fill(docImage, 224, 260, imaging.Top, imaging.Lanczos)
			dc.DrawImage(docImageResized, 196, 805)
		}
	}

	err = dc.SavePNG("output/" + d.code + ".png") // save it
	if err != nil {
		fmt.Println(err)
	}

	// Upload to S3
	filename := fmt.Sprintf("output/%s.png", d.code)
	key := fmt.Sprintf("wallmount2.0/%s.png", d.code)

	if _, err := os.Stat(filename); err == nil {
		UploadToS3(*s3client, filename, key)

	}

	// Cleanup
	err = os.Remove("images/" + d.code + ".jpg")
	err = os.Remove("images/qr-" + d.code + ".png")
	if err != nil {
		fmt.Println(err)
	}

}

var NewCodes = []string{"AZ544897"}

func UploadToS3(client s3.Client, name string, key string) {
	stat, err := os.Stat(name)
	bucket := os.Getenv("S3_BUCKET")
	if err != nil {
		panic("Couldn't stat image: " + err.Error())
	}
	file, err := os.Open(name)

	if err != nil {
		panic("Couldn't open local file")
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: stat.Size(),
	})

	file.Close()

	if err != nil {
		panic("Couldn't upload file: " + err.Error())
	}
}

func main() {
	resp, err := client.R().
		SetHeader("client", "service").
		SetHeader("access_token", "Ekyhed9jslfRxJc8J2ajbDmzcjPdWK0p").
		Get("https://api.zyla.in/doctor/all")

	wg := sync.WaitGroup{}
	queue := make(chan Doc)

	if err != nil {
		fmt.Errorf("couldn't find %s", err.Error())

	}

	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		panic("Failed to load configuration")
	}

	s3client := s3.NewFromConfig(cfg)
	// tt := time.Now().UTC().UnixMilli()

	// fmt.Println(resp)
	var dat []map[string]interface{}
	json.Unmarshal(resp.Body(), &dat)
	var docs []Doc
	for _, d := range dat {
		//docString := fmt.Sprintf("Dr-%s-%s", strings.Replace(strings.Replace(d["name"].(string), "Dr. ", "", 1), " ", "-", -1), d["code"])
		code := d["code"].(string)
		if code[0:2] == "AZ" || code[0:2] == "HH" {
			name := d["name"].(string)
			name = "Dr. " + strings.Replace(name, "Dr.", "", 1)

			//if float64(tt)-d["createdAt"].(float64) <= 172569444 {
			if d["createdAt"].(float64) >= 1661365873000 {
				//if d["code"].(string) == "AZ59000B" {
				docs = append(docs, Doc{name: name, code: d["code"].(string), designation: d["title"].(string)})
			}
		}
	}

	for worker := 0; worker < Limit; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for work := range queue {
				drawPoster(work, s3client)
			}
		}(worker)
	}

	for _, d := range docs {
		/*
			if d.code == "AZ269294" || d.code == "AZ79813C" || d.code == "AZ54751D" || d.code == "AZ9556B3" || d.code == "AZE81091" {
				queue <- d
			}
		*/
		queue <- d
	}

	close(queue)
	wg.Wait()

}
