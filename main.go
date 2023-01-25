package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/go-resty/resty/v2"
	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/uploader"
)

// 2160 1850
const WIDTH = 2150//2895 / 2
const HEIGHT = 1860//4096 / 2
const URL = "https://api.whatsapp.com/send?phone=919818111918&text="
// const MESSAGE = "Happy with the virtual care clinic services available on Zyla app"

const Limit = 5

var clientResty = resty.New()

type ImageKitFileRequest struct{
	FileName 	string 	`json:"fileName"`
	Folder 		string `json:"folder"`
}
// {
//     "fileName": "nimit",
//     "folder": "doctor_feedback"
// }
type ImageKitFileResponse struct{
	Url 		string 	`json:"url"`
	Name 		string `json:"name"`
	FilePath 	string `json:"filePath"`
}

type DoctorProfile struct{
	Code 		string 	`json:"code"`
	Name 		string `json:"name"`
	Speciality 	string `json:"speciality"`
	Phoneno 	int `json:"phoneno"`
	Id 	int `json:"id"`
}
type Doc struct {
	name        string
	code        string
	designation string
}
type UserResponse struct {
	ID          int64    `json:"id"`
	PhoneNo     int64    `json:"phoneno"`
	CountryCode int      `json:"countryCode"`
	Patient     *Patient `json:"patientProfile"`
}
type Patient struct {
	ID              int64     `json:"id"`
	PhoneNo         int64     `json:"phoneno"`
	FirstName       string    `json:"firstName"`
	LastName        string    `json:"lastName"`
	FullName        string    `json:"fullName"`
	ProfileImage    string    `json:"profileImage"`
	Age             int       `json:"age"`
	Location        string    `json:"location"`
	Gender          int       `json:"gender"`
	Email           string    `json:"email"`
	DateOfBirth     int64 	  `json:"dateOfBirth"`
	Profession      string    `json:"profession"`
	Type            int       `json:"type"`
	Status          int       `json:"status"`
	DidYouKnowCount int       `json:"countDidYouKnow"`
	ReferredBy      int64     `json:"referredBy"`
	ReferralCode    string    `json:"referralCode"`
	ClientCode      string    `json:"clientCode"`
	CountryCode     int       `json:"countryCode"`
	// CreatedAt       string `json:"createdAt"`
	// UpdatedAt       s `json:"updatedAt"`
	// Code       		string 	  `json:"referralCode"` 
}
type Poster struct {
	PatientName     string 		`json:"patientName"`
	DoctorCode      string 		`json:"doctorCode"`
	Rating 			string 		`json:"rating"`
	DoctorName		string 		`json:"doctorName"`
	DoctorNumber	string 		`json:"doctorNumber"`
	DoctorId		string 		`json:"doctorId"`
	Comment			string 		`json:"comment"`
	DoctorSpec		string		`json:"doctorSpec"`
	PatientId		string		`json:"patientId"`
	Date			string		`json:"date"`
}

type WhatsappRequest struct{
	UserId 		string 	`json:"userId"`
	PhoneNumber 		string `json:"phoneNumber"`
	CountryCode 	string `json:"countryCode"`
	Event 	string `json:"event"`
	Traits 	WhatsappTraits `json:"traits"`
}
type WhatsappTraits struct{
	ImageUrl 		string 	`json:"imageUrl"`
	Status 		string `json:"status"`
	PatientName 	string `json:"patientName"`
	VccLink 	string `json:"vccLink"`
}
func drawPoster(poster Poster,client *s3.Client) {
log.Println("draw-potser",poster)
	dc := gg.NewContext(WIDTH, HEIGHT) // canvas 1000px by 1000px

	bgImage, err := gg.LoadImage("Template-1.jpg")
	if err != nil {
		fmt.Println(err)
	}
	dc.DrawImage(bgImage, 0, 0)
	textColor := color.Black
	if err := dc.LoadFontFace("pn.ttf", 88); err != nil {
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

	s := poster.DoctorName
	lineHeight := 1.5
	_, textHeight := dc.MeasureMultilineString(s, lineHeight)

	x := 520.0
	y := 1300.0

	dc.SetRGBA255(106, 94, 245,251)
	dc.DrawString(s, x, y)

	if err := dc.LoadFontFace("pn.ttf", 58); err != nil {
		fmt.Println(err)
	}
	dc.SetColor(color.Black)
	dc.DrawString(poster.DoctorSpec, x, y+textHeight+48,)

	if err := dc.LoadFontFace("pn.ttf", 98); err != nil {
		fmt.Println(err)
	}

	dc.SetColor(color.Black)
	// dc.SetRGB(106, 94, 245)
	dc.DrawString(poster.PatientName, 690, 322)

	// qrString := URL + url.QueryEscape(fmt.Sprintf(MESSAGE, d.name, d.code))
	//err = qrcode.WriteFile(qrString, qrcode.Medium, 310, "images/qr-"+d.code+".png")
	// err = qrcode.WriteFile(qrString, qrcode.Medium, 342, "images/qr-"+d.code+".png")
	// 8NPS =4 star, 9NPS= 4.5 stars, 10NPS =5 stars
	// qrImage:={}
	// qrImage, err := gg.LoadImage("stars/5 star.png")
	var starImage string //:="stars/5 star.png"
	if poster.Rating=="10"{
		starImage="stars/5 star.png"
	}else if poster.Rating=="9"{
		starImage="stars/4.5 star.png"
	}else if poster.Rating=="8"{
		starImage="stars/4 star.png"
	}
	log.Println("star image select",starImage)
	qrImage, err := gg.LoadImage(starImage)
	// qrImage.Bounds()
	// log.Println("image bonds:",qrImage.Bounds())
	// qrImage.Bounds()
	docImageResized := imaging.Fill(qrImage, 450, 75, imaging.Center, imaging.Lanczos)
	// log.Println("resize image bonds:",docImageResized.Bounds())
	dc.DrawImage(docImageResized, 690, 350)

	if err := dc.LoadFontFace("pn.ttf", 58); err != nil {
		fmt.Println(err)
	}
	dc.SetColor(color.Black)
	var MESSAGE string
	MESSAGE = "Happy with the virtual care clinic services available on Zyla app"
	if len(poster.Comment)>0{
		MESSAGE=poster.Comment
	}
	dc.DrawStringWrapped(MESSAGE,350, 600, 0, 0, 1450, 1.5, gg.AlignLeft)
	//set date
	if err := dc.LoadFontFace("pn.ttf", 44); err != nil {
		fmt.Println(err)
	}
	// dc.SetColor(color.RGBA{
	// 	R: uint8(237),
	// 	G: uint8(g),
	// 	B: uint8(225),
	// 	A: uint8(1),
	// })
	
	dc.SetRGB(119, 119, 119)
	date:=GetDate(poster.Date) //Dec 12, 2022
	dc.DrawString(date, 1150, 422)
//uncomment
	url := fmt.Sprintf("https://az-doc.s3.ap-south-1.amazonaws.com/%s.jpg", poster.DoctorCode)
	fmt.Println(url)
	// fmt.Println("Generating for ", d)
	resp, err := clientResty.R().Get(url)
	if resp.IsSuccess() {
		ioutil.WriteFile("images/"+poster.DoctorCode+".jpg", resp.Body(), 0666)
		docImage, err := gg.LoadImage("images/" + poster.DoctorCode + ".jpg")
		if err != nil {
			fmt.Errorf("Couldn't load image for %s", poster.DoctorCode)
		} else {
			docImageResized := imaging.Fill(docImage, 300, 300, imaging.Top, imaging.Lanczos)
			dc.DrawImage(docImageResized, 150, 1205)
		}
	}
//uncooment
	err = dc.SavePNG("output/" + poster.PatientId + ".png") // save it
	if err != nil {
		fmt.Println(err)
	}



	// Upload to S3
	filename := fmt.Sprintf("output/%s.png", poster.PatientId)
	// key := fmt.Sprintf("feedback/%s.png", poster.PatientId)

	// fileLink:=UploadToS3(*client, filename, key)
	fileLink:=UploadToIMageKit(poster,filename)	
	if len(fileLink)>0{
		log.Println("Url:",fileLink)
		// MEMBER
		vccLink:=GetVccLink(poster.DoctorCode)
		var status string
		status=""
		if poster.Rating=="8"||poster.Rating=="9"{
			status="satisfied"
		}else if poster.Rating=="10"{
			status="extremely satisfied"
		}
		WhatsappEvent(fileLink,poster.DoctorId,poster.DoctorNumber,status,poster.PatientName,vccLink)
	}

	// Cleanup
	err = os.Remove("images/" + poster.PatientId + ".png")
	// err = os.Remove("images/qr-" + d.code + ".png")
	if err != nil {
		fmt.Println(err)
	}

}


func GetDate(date string) (string) {
	dateList:=strings.Split(date," ")
	if len(dateList)==2{
		dateAr:=strings.Split(dateList[0],"-")
		year:= dateAr[0]

		intMonth,_:=strconv.Atoi(dateAr[1])
		month:= time.Month(intMonth).String()

		day:=dateAr[2]

		return month + ", " + day +" "+year 
	}

	return "Dec 12, 2022"
}

func UploadToIMageKit(poster Poster, name string) string {

	priv:="private_kabOQbr6HnbSii/u+K2mwbCUhKA="
	pub:= "public_UvTu9aPPEYck82weBemvQETkDqI="

	fiii := fmt.Sprintf("output/%s.png", poster.PatientId)
	log.Println("File:",fiii)
	image64:=IMagetoBase64(fiii)
	ctx := context.Background()
	
	ik := imagekit.NewFromParams(imagekit.NewParams{
		PublicKey: pub,
		PrivateKey: priv,
		UrlEndpoint: "https://ik.imagekit.io/e2tisv3povj",
	})
	resp, err := ik.Uploader.Upload(ctx, image64, uploader.UploadParam{
		FileName: "test-image",
		Folder: "doctor_feedback",
	})

	if err != nil {
		log.Println("error=>>",err)
	}
	return resp.Data.Url//"https://s3.ap-south-1.amazonaws.com/az-fedback-posters/"
}
func UploadToS3(client s3.Client, name string, key string) string {
	stat, err := os.Stat(name)
	bucket := "az-fedback-posters"//os.Getenv("S3_BUCKET")
	if err != nil {
		panic("Couldn't stat image: " + err.Error())
	}
	file, err := os.Open(name)

	if err != nil {
		panic("Couldn't open local file")
	}
	imagetype:="image/jpg"
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: stat.Size(),
		ACL: "public-read",
		ContentType: &imagetype,
	})

	file.Close()

	if err != nil {
		log.Println("Couldn't upload file: " + err.Error())
		return ""
	}
	return "https://s3.ap-south-1.amazonaws.com/az-fedback-posters/"+key
}

func main() {

	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		panic("Failed to load configuration")
	}

	s3client := s3.NewFromConfig(cfg)
	ReadCSV(s3client)

}
func ReadCSV(s3client *s3.Client){

	wg := sync.WaitGroup{}
	queue := make(chan Poster)

	recordFile, err := os.Open("feedback.csv")

    if err != nil {
        // fmt.Println(err ,"erf")
    }

    reader := csv.NewReader(recordFile)
	records, err := reader.ReadAll()
	if err!=nil{
		log.Println("Error: ",err)
	}
	//  allList :=[]TestData{}
	for worker := 0; worker < Limit; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for work := range queue {
				log.Println("work work wrk",work)
				drawPoster(work, s3client)
			}
		}(worker)
	}

	for  _, rec := range records {

	
	
		patientId:=rec[2]
		if len(patientId)>0&&rec[3]=="AZ_touchpoint_2_30_to_90_day_onboarding"&&(rec[4]=="8"||rec[4]=="9"||rec[4]=="10"){

			log.Println("records",rec)
			log.Println("date:",rec[0])
			log.Println("name:",rec[1])
			log.Println("patientId:",rec[2])
			log.Println("surveyId:",rec[3])//AZ_touchpoint_2_30_to_90_day_onboarding
			log.Println("rating:",rec[4])
			log.Println("comment:",rec[7])


			log.Println("For patientId: ",patientId)
			poster:=Poster{}
		
			posterResponse,_:= GetDetails(patientId)
	
			poster=posterResponse
			poster.Rating=rec[4]//rating
			poster.Comment=rec[7]//comment
			poster.PatientId=patientId
			poster.Date=rec[0]
			// res.DoctorName=
			// poster.DoctorName="Dr Arun Yadav"
			// poster.DoctorSpec="Consultant Physician and Diabetologist"
			// poster.PatientName="Nimit BB"
	
			log.Println("->>>",poster)
			queue <- poster
			// drawPoster(poster,s3client)
				
		}
	
	}
	close(queue)
	wg.Wait()




//main code
	// rec :=  records[25]

	// log.Println("records",rec)
	// log.Println("date:",rec[0])
	// log.Println("name:",rec[1])
	// log.Println("patientId:",rec[2])
	// log.Println("surveyId:",rec[3])//AZ_touchpoint_2_30_to_90_day_onboarding
	// log.Println("rating:",rec[4])
	// log.Println("comment:",rec[7])

	// patientId:=rec[2]
	// if len(patientId)>0{
	// 	poster:=Poster{}
	// 	posterResponse,_:= GetDetails(patientId)
	// 	poster=posterResponse
	// 	poster.Rating=rec[4]//rating
	// 	poster.Comment=rec[7]//comment
	// 	poster.PatientId=patientId
	// 	poster.Date=rec[0]

	// 	log.Println("->>>",poster)
	// 	drawPoster(poster,s3client)
	// }
//main code


}

func GetDetails(patientId string) (Poster, error) {

	patient:=GetUserByPatientID(patientId)
	poster:=Poster{}

	if patient!=nil{
		log.Println("PPPPP",patient)
		poster.DoctorCode=patient.ReferralCode
		poster.PatientName=patient.FirstName+" "+patient.LastName
	
		doctor,_:=GetDoctor(patient.ReferralCode)
		if doctor!=nil{
			log.Println("DDDDD",doctor)
			poster.DoctorName=doctor.Name
			poster.DoctorSpec=doctor.Speciality
			docId:= strconv.Itoa(int(doctor.Id))
			poster.DoctorId=docId
			docPhone:= strconv.Itoa(int(doctor.Phoneno))
			poster.DoctorNumber=docPhone
		}else{
			log.Println("Erroror doctor")
		}
	}else{
		log.Println("Erroror patient")
	}
	return poster,nil
}
func GetDoctor(code string) (*DoctorProfile,error) {
	docProfile:=DoctorProfile{}
	var url ="https://api.zyla.in/docprofile/doctor?code="+code
	res, err := clientResty.R().
	SetHeader("Content-Type", "application/json").
	SetHeader("auth_token", "cb7bdf36-702f-4791-918f-947dd4b6f07e-3fd004b0aa48974c").
	SetHeader("phone_no", "7838651405").
	SetHeader("client", "bridge_web").
	SetResult(&docProfile).
	Get(url)
	if err!=nil{
		log.Println("res",res,"err:",err)
		return &docProfile, err
	}
	log.Println("result doctor",docProfile)
	return &docProfile,nil
}

//GetUserByPatientID -
func  GetUserByPatientID(patientID string) *Patient {
	url := "https://api.zyla.in/patients/"
	mainurl := url+patientID //fmt.Sprintf(url, patientID)
	response := Patient{}

	log.Println("patient url :",mainurl)
	_, err := clientResty.R().
		SetHeader("auth_token", "cb7bdf36-702f-4791-918f-947dd4b6f07e-3fd004b0aa48974c").
		SetHeader("phone_no", "7838651405").
		SetHeader("client", "bridge_web").
		SetResult(&response).
		Get(mainurl)

	if err != nil {
		log.Println("Error:",err)
	}
	log.Println("patient response",response.ReferralCode,response.FirstName,response.LastName)

	return &response
}
// 
func  WhatsappEvent(fileLink string,doctorId string,doctorNmber string,status string,patientName string,vccLink string) bool {
	url := "https://services.prod.zyla.in/api/v2/zylawhatsapp/event"
	// mainurl := url+patientID //fmt.Sprintf(url, patientID)
	// response := Patient{}
	whatsappRequest:=WhatsappRequest{}
	log.Println("patient url :",doctorNmber)
	whatsappRequest.Event="DOCTOR_FEEDBACK"
	whatsappRequest.CountryCode="91"
	whatsappRequest.PhoneNumber=doctorNmber //"9911990012"//"8980666234"//"6377337453"////,"9456563445"//"9911990012"//doctorNmber//doctor number
	whatsappRequest.UserId=""//doctorId//doctor id
	whatsappRequest.Traits.ImageUrl=fileLink
	whatsappRequest.Traits.Status=status//??
	whatsappRequest.Traits.PatientName=patientName
	whatsappRequest.Traits.VccLink=vccLink[20:]
	_, err := clientResty.R().
		SetHeader("auth_token", "cb7bdf36-702f-4791-918f-947dd4b6f07e-3fd004b0aa48974c").
		SetHeader("phone_no", "7838651405").
		SetHeader("client", "bridge_web").
		SetBody(&whatsappRequest).
		Post(url)

	if err != nil {
		log.Println("Whatsapp Error:",err)
		return false
	}
	log.Println("whatsapp done response")

	return true
}
func GetVccLink(doctorCode string)string{

	recordFile, err := os.Open("vccLink.csv")

    if err != nil {
        // fmt.Println(err ,"erf")
    }

    reader := csv.NewReader(recordFile)
    // records, err := reader.ReadAll()
	 records, err := reader.ReadAll()
	if err!=nil{
		log.Println("Error: ",err)
	}

	result:=""
	for  _, rec := range records {
		if rec[0]==doctorCode{
			result=rec[7]
		}
	}
	return result
}
func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
func IMagetoBase64(image string)string{

		bytes, err := ioutil.ReadFile(image)
		if err != nil {
			log.Fatal(err)
		}
	
		var base64Encoding string
	
		// Determine the content type of the image file
		mimeType := http.DetectContentType(bytes)
	
		// Prepend the appropriate URI scheme header depending
		// on the MIME type
		switch mimeType {
		case "image/jpeg":
			base64Encoding += "data:image/jpeg;base64,"
		case "image/png":
			base64Encoding += "data:image/png;base64,"
		}
	
		// Append the base64 encoded output
		base64Encoding += toBase64(bytes)
	
		// Print the full base64 representation of the image
		// fmt.Println(base64Encoding)
		return base64Encoding
}