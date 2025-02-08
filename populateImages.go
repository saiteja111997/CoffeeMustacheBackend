package main

import (
	helper "coffeeMustacheBackend/pkg/helper"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type MenuItem struct {
	ID       uint
	Name     string
	ImageURL sql.NullString
}

type PexelsResponse struct {
	Page         int     `json:"page"`
	PerPage      int     `json:"per_page"`
	Photos       []Photo `json:"photos"`
	TotalResults int     `json:"total_results"`
	NextPage     string  `json:"next_page"`
}

type Photo struct {
	ID              int      `json:"id"`
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	URL             string   `json:"url"`
	Photographer    string   `json:"photographer"`
	PhotographerURL string   `json:"photographer_url"`
	PhotographerID  int      `json:"photographer_id"`
	AvgColor        string   `json:"avg_color"`
	Src             PhotoSrc `json:"src"`
	Liked           bool     `json:"liked"`
	Alt             string   `json:"alt"`
}

type PhotoSrc struct {
	Original  string `json:"original"`
	Large2x   string `json:"large2x"`
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Small     string `json:"small"`
	Portrait  string `json:"portrait"`
	Landscape string `json:"landscape"`
	Tiny      string `json:"tiny"`
}

var (
	pexelsAPIKey string
	s3Bucket     string
	s3Region     string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading environment variables file")
	}
	pexelsAPIKey = os.Getenv("PEXELS_API_KEY")
	s3Bucket = os.Getenv("S3_BUCKET_NAME")
	s3Region = os.Getenv("S3_REGION")
}

func getImageFromPexels(query string) (string, error) {

	url := fmt.Sprintf("https://api.pexels.com/v1/search?query=%s&per_page=1", strings.ReplaceAll(query, " ", "+"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", pexelsAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var pexelsResp PexelsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println("Printing length of response : ", string(body), len(body))
	if err := json.Unmarshal(body, &pexelsResp); err != nil {
		return "", err
	}

	fmt.Println("Printing response from Pexels API: ", pexelsResp)

	if len(pexelsResp.Photos) > 0 {
		return pexelsResp.Photos[0].Src.Tiny, nil
	}
	return "", fmt.Errorf("no image found")
}

func updateMenuItemImage(db *gorm.DB, id uint, imageURL string) error {
	return db.Model(&MenuItem{}).Where("id = ?", id).Update("image_url", imageURL).Error
}

func main() {

	// Database connection
	DB_USERNAME := os.Getenv("DB_USERNAME")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOSTNAME := os.Getenv("DB_HOSTNAME")
	DB_PORT := os.Getenv("DB_PORT")
	DATABASE := os.Getenv("DATABASE")

	//ctx := context.Background()
	if err := waitForHost(DB_HOSTNAME, DB_PORT); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Connection established")

	db, err := helper.Open(helper.Config{
		Username: DB_USERNAME,
		Password: DB_PASSWORD,
		Hostname: DB_HOSTNAME,
		Port:     DB_PORT,
		Database: DATABASE,
	})

	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	// Fetch all menu items
	var menuItems []MenuItem
	if err := db.Find(&menuItems).Error; err != nil {
		log.Fatalf("Error fetching menu items: %v", err)
	}

	fmt.Printf("Processing %d menu items for cross-sell suggestions...\n", len(menuItems))

	for _, item := range menuItems {

		time.Sleep(2 * time.Second)
		fmt.Println("Processing:", item.Name, item.ID)

		imageURL, err := getImageFromPexels(item.Name)
		if err != nil {
			fmt.Println("Error fetching image:", err)
			continue
		}

		err = updateMenuItemImage(db, item.ID, imageURL)
		if err != nil {
			fmt.Println("Error updating DB:", err)
		}
	}
}

func waitForHost(host, port string) error {
	timeOut := time.Second

	if host == "" {
		return errors.Errorf("unable to connect to %v:%v", host, port)
	}

	for i := 0; i < 60; i++ {
		fmt.Printf("waiting for %v:%v ...\n", host, port)
		conn, err := net.DialTimeout("tcp", host+":"+port, timeOut)
		if err == nil {
			fmt.Println("done!")
			conn.Close()
			return nil
		}

		time.Sleep(time.Second)
	}

	return errors.Errorf("timeout attempting to connect to %v:%v", host, port)
}
