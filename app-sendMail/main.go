package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

// SMTP 服务器设置
const (
	smtpHost = "smtp.gmail.com"
	smtpPort = "587"
)

type UploadData struct {
	Text  string `json:"text"`
	Image string `json:"image"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "https://funwilliam.github.io")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func sendMail(from, password, to, subject, body, base64Image string) error {
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// 构建邮件内容
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Type: multipart/mixed; boundary=boundary\n\n" +
		"--boundary\n" +
		"Content-Type: text/plain; charset=utf-8\n\n" +
		body + "\n"

	// 添加图片附件
	if base64Image != "" {
		msg += "--boundary\n" +
			"Content-Type: image/png\n" +
			"Content-Transfer-Encoding: base64\n" +
			"Content-Disposition: attachment; filename=\"image.png\"\n\n" +
			base64Image + "\n" +
			"--boundary--"
	}

	// 发送邮件
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	return err
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var data UploadData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 处理 Base64 图像数据
	base64Image := strings.TrimPrefix(data.Image, "data:image/png;base64,")

	// SMTP 配置
	from := "your-email@gmail.com"       // 更改为您的 Gmail 地址
	password := os.Getenv("SMTP_PASS")   // Gmail 密码或应用专用密码
	to := "recipient@example.com"        // 更改为收件人地址
	subject := "New submission received" // 邮件主题

	if err := sendMail(from, password, to, subject, data.Text, base64Image); err != nil {
		log.Printf("Failed to send email: %v\n", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Email sent successfully")
}

func main() {
	http.HandleFunc("/", uploadHandler)
	fmt.Println("Server started at :443")
	log.Fatal(http.ListenAndServe(":443", nil))
}
