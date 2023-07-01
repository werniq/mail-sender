package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type EmailHeader struct {
	DeliveredTo              string `json:"Delivered-To"`
	Received                 string `json:"Received"`
	XGoogleSmtpSource        string `json:"X-Google-Smtp-Source"`
	XReceived                string `json:"X-Received"`
	ARCSeal                  string `json:"ARC-Seal"`
	ARCMessageSignature      string `json:"ARC-Message-Signature"`
	ARCAuthenticationResults string `json:"ARC-Authentication-Results"`
	ReturnPath               string `json:"Return-Path"`
	ReceivedFrom             string `json:"Received-From"`
	ReceivedSPF              string `json:"Received-SPF"`
	AuthenticationResults    string `json:"Authentication-Results"`
	DKIMSignature1           string `json:"DKIM-Signature1"`
	DKIMSignature2           string `json:"DKIM-Signature2"`
	XFeedbackID              string `json:"X-Feedback-Id"`
	XMailgunSendingIP        string `json:"X-Mailgun-Sending-Ip"`
	XMailgunSID              string `json:"X-Mailgun-Sid"`
	ReceivedFromMailgun      string `json:"Received-From-Mailgun"`
	Sender                   string `json:"Sender"`
	Date                     string `json:"Date"`
	MimeVersion              string `json:"Mime-Version"`
	Subject                  string `json:"Subject"`
	From                     string `json:"From"`
	To                       string `json:"To"`
	XMailgunTag              string `json:"X-Mailgun-Tag"`
	XMailgunTemplateName     string `json:"X-Mailgun-Template-Name"`
	MessageID                string `json:"Message-Id"`
	ContentType              string `json:"Content-Type"`
	ContentTransferEncoding  string `json:"Content-Transfer-Encoding"`
}

type EmailResponse struct {
	Header EmailHeader `json:"headers"`
	Body   string      `json:"body"`
}

type Message struct {
	From string `json:"from"`
	ID   string `json:"id"`
	Text struct {
		Body string `json:"body"`
	} `json:"text"`
	Timestamp int    `json:"timestamp"`
	Type      string `json:"type"`
}

type EmailEnvelope struct {
	UID           uint32
	Envelope      Envelope
	BodyStructure []BodyStructure
}

type Envelope struct {
	Date      time.Time
	Subject   string
	From      []EmailContact
	To        []EmailContact
	Cc        []EmailContact
	Bcc       []EmailContact
	InReplyTo string
	MessageID string
}

type EmailContact struct {
	Name  string
	Email string
}

type BodyStructure struct {
	MIMEType          string
	MIMEParams        map[string]string
	BodyType          string
	BodySubType       string
	Parameters        map[string]string
	Disposition       string
	DispositionParams map[string]string
	Language          []string
	Encoding          string
	Size              uint32
	Lines             uint32
	ExtendedData      interface{}
	Parts             []BodyStructure
}

type Response struct {
	Messages []Message `json:"messages"`
}

type WhatsAppMessageData struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             TextMessage `json:"text"`
}

type WhatsAppResponseData struct {
	MessagingProduct string            `json:"messaging_product"`
	Contacts         []Contact         `json:"contacts"`
	Messages         []ResponseMessage `json:"messages"`
}

type Contact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type ResponseMessage struct {
	Id string `json:"id"`
}

type TextMessage struct {
	Text string `json:"text"`
}

type GmailAuth struct {
	Email    string
	Password string
}

type Business struct {
	VerifiedName       string `json:"verified_name"`
	DisplayPhoneNumber string `json:"display_phone_number"`
	ID                 string `json:"id"`
	QualityRating      string `json:"quality_rating"`
}

type Application struct {
	GmailAuth     *GmailAuth
	BotApi        *tgbotapi.BotAPI
	TgBotApiToken string
}

var EmailAddress = "qniwerniq@gmail.com"

func ReadFromReader(reader io.Reader) ([]byte, error) {
	buffer := make([]byte, 1024) // Adjust the buffer size as needed
	var data []byte

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		data = append(data, buffer[:n]...)
	}

	return data, nil
}

// TgChannelId
// https://api.telegram.org/bot<YourBotToken>/getUpdates
var TgChannelId = int64(869473101)

// RetrieveUpcomingEmails retrieves the upcoming emails from the server
func (App *Application) RetrieveUpcomingEmails(server, username, password string) {
	// Connect to the server
	c, err := client.DialTLS(server, nil)
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}
	defer c.Logout()

	// Authenticate with the server
	if err := c.Login(username, password); err != nil {
		fmt.Println("Error authenticating with the server:", err)
		return
	}

	_, err = c.Select("INBOX", false)
	if err != nil {
		fmt.Println("Error selecting the mailbox:", err)
		return
	}

	// Create a channel to receive new messages
	newMsgs := make(chan *imap.Message)

	// Create a channel to receive errors
	errCh := make(chan error)

	// Create a goroutine to continuously fetch new messages
	go func() {
		// Loop indefinitely to fetch new messages
		for {
			// Search for new messages
			searchCriteria := imap.NewSearchCriteria()
			searchCriteria.WithoutFlags = []string{string(imap.StatusUnseen)} // Retrieve only unseen messages
			searchCriteria.Since = time.Now().Add(time.Minute * -5)

			uids, err := c.UidSearch(searchCriteria)
			if err != nil {
				errCh <- err
				return
			}

			// Fetch the new messages
			seqset := new(imap.SeqSet)
			seqset.AddNum(uids...)
			fetchedMsgs := make(chan *imap.Message)
			go func() {
				errCh <- c.UidFetch(seqset, []imap.FetchItem{imap.FetchItem("BODY.PEEK[]"), imap.FetchEnvelope, imap.FetchBodyStructure}, fetchedMsgs)
			}()

			// Process the fetched messages
			for msg := range fetchedMsgs {
				newMsgs <- msg
			}

			// Wait for a certain duration before checking for new messages again
			time.Sleep(time.Minute * 5)
		}
	}()

	// Handle new messages and errors
	for {
		select {
		case msg := <-newMsgs:
			if msg.Envelope.Date.Format("2006-01-02 15:04:05") < time.Now().Add(time.Minute*-5).Format("2006-01-02 15:04:05") {
				continue
			}

			var messageData []byte
			for _, literal := range msg.Body {
				data, err := io.ReadAll(literal)
				if err != nil {
					fmt.Println("Error reading message body:", err)
					return
				}
				messageData = append(messageData, data...)

				fmt.Println("Message body:", string(data))
			}

			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(messageData))
			if err != nil {
				fmt.Println("Error parsing message body:", err)
				return
			}

			plainText := doc.Find("div").Text()

			sentText := plainText + "\n\n"
			if len(sentText) > 4096 {
				sentText = sentText[:3900]
			}

			message := tgbotapi.NewMessage(TgChannelId, fmt.Sprintf(
				"Date: %s\n\nFrom: %s\nSubject: %s\n\nBody: %v",
				msg.Envelope.Date.Format("2006-01-02 15:04:05"),
				msg.Envelope.From[0].Address(),
				msg.Envelope.Subject,
				sentText,
			))

			_, err = App.BotApi.Send(message)
			if err != nil {
				App.BotApi.Send(tgbotapi.NewMessage(TgChannelId, fmt.Sprintf("Error sending message: %s", err.Error())))
				return
			}

		case err := <-errCh:
			// Handle the error
			fmt.Println("Error:", err)
		}
	}
}

func (App *Application) handleConnection(conn net.Conn) {
	defer conn.Close()

	msg, err := mail.ReadMessage(conn)
	if err != nil {
		log.Println(err)
		return
	}

	to := msg.Header.Get("to")
	subject := msg.Header.Get("Subject")
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		log.Println(err)
		return
	}

	go App.SendEmail(to, subject, string(body))
}

func (g *GmailAuth) ChangeEmailAndPassword(email, password string) {
	g.Email = email
	g.Password = password
}

func (App *Application) SendEmail(to, subject, content string) {
	smtpHost := "smtp.gmail.com"
	stmtPort := 587

	// Email content
	body := fmt.Sprintf("%s", content)

	// Create the authentication credentials
	auth := smtp.PlainAuth("", App.GmailAuth.Email, App.GmailAuth.Password, smtpHost)

	fmt.Println(auth)

	message := []byte("Email to: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n\n" +
		body + "\r\n")

	// Send the email
	err := smtp.SendMail(smtpHost+":"+fmt.Sprintf("%d", stmtPort), auth, App.GmailAuth.Email, []string{to}, message)
	if err != nil {
		App.BotApi.Send(tgbotapi.NewMessage(TgChannelId, fmt.Sprintf("Error sending email: %s", err.Error())))
		return
	}
	fmt.Println("Email Sent!")
}

func (App *Application) MessageHandler(c *gin.Context) {
	var update tgbotapi.Update
	err := c.ShouldBindJSON(&update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error decoding response body: %s", err.Error()),
		})
		return
	}

	if strings.Contains(update.Message.Text, "/changeemail") {
		args := strings.Split(strings.TrimPrefix(update.Message.Text, "/changeemail"), ",")
		if args == nil {
			_, err = App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Invalid message format /changeemail <email> <password>")))
			return
		} else {
			App.GmailAuth.Email = args[0]
			App.GmailAuth.Password = args[1]
			_, err = App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Email successfully changed"))
			if err != nil {
				App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error sending message: %s", err.Error())))
				return
			}
		}
		return
	}
	// Split the message into recipient email, subject, and body
	messageParts := strings.Split(update.Message.Text, ";")

	if messageParts[0] == "/start" {
		_, err = App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the Email Bot!"))
		if err != nil {
			App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error sending message: %s", err.Error())))
			return
		}
		return
	}

	if len(messageParts) != 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid message format: %d", update.Message.From.ID),
		})
		return
	}

	to := strings.TrimSpace(messageParts[0])
	subject := strings.TrimSpace(messageParts[1])
	body := strings.TrimSpace(messageParts[2])

	// Send the message to the recipient email
	fmt.Println("sending email...")
	App.SendEmail(to, subject, body)
	_, err = App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Email successfully sent!"))
	if err != nil {
		App.BotApi.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error sending message: %s", err.Error())))
		log.Println(err)
		return
	}
}

func (App *Application) Serve(tun ngrok.Tunnel) error {
	r := gin.Default()

	fmt.Println()

	App.ConfigureRoutes(r)

	return http.Serve(tun, r)
}

func (App *Application) ConfigureRoutes(r *gin.Engine) {
	r.POST("/webhook"+App.TgBotApiToken, App.MessageHandler)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	gmailAuth := &GmailAuth{
		Email:    os.Getenv("YOUR_EMAIL"),
		Password: os.Getenv("YOUR_APP_CODE"),
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Printf("error creating new bot: %v\n", err)
		return
	}

	tun, err := ngrok.Listen(context.Background(),
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)

	if err != nil {
		log.Printf("error creating ngrok tunnel: %v\n", err)
		return
	}

	webhookURL := tun.URL() + "/webhook" + bot.Token

	// sets up webhook for bot
	_, err = http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook?url=%s", bot.Token, webhookURL), "application/json", nil)
	if err != nil {
		log.Printf("error setting webhook: %v\n", err)
		return
	}

	bot.ListenForWebhook(webhookURL)

	app := Application{gmailAuth, bot, bot.Token}

	rand.NewSource(time.Now().UnixNano())
	//emailSubject := generateRandomString(10)
	//emailBody := generateRandomString(20)

	// Create the variable with the generated values

	go app.RetrieveUpcomingEmails("imap.gmail.com:993", app.GmailAuth.Email, app.GmailAuth.Password)

	if err := app.Serve(tun); err != nil {
		panic(err)
	}
}
