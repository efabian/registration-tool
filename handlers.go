package main

import (
	"bytes"
	"crypto/md5"
	"crypto/subtle"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/gomail.v2"
)

// titler is used to replace the deprecated strings.Title.
var titler = cases.Title(language.Und)

type templateTags struct {
	FName    string
	LName    string
	Local    string
	District string
	Time     string
	ZQR      string
	ZLink    string
	ZMeet    string
	ZPass    string
	ZDate    string
}

// render parses and executes a template, writing the result to w.
// On error it writes a 500 response and returns.
func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
		return // FIX M-2: was missing; prevents partial-write + double-header confusion
	}
}

// dayKey maps a lowercase day abbreviation to the uppercase key used in cfg.Zoom.
func dayKey(day string) string {
	return strings.ToUpper(day)
}

// createMessage builds the HTML confirmation email body.
// FIX M-1: now returns (string, error) so callers can handle failures.
func createMessage(fname, lname, local, district, day string) (string, error) {
	zoom, ok := cfg.Zoom[dayKey(day)]
	if !ok {
		log.Printf("no zoom config found for day %q", day)
		zoom = ZoomConfig{}
	}

	meetTime := "8:45PM"
	if district == "HK" {
		meetTime = "9:45PM"
	}

	tags := templateTags{
		FName:    fname,
		LName:    lname,
		Local:    local,
		District: district,
		Time:     meetTime,
		ZQR:      zoom.QR,
		ZLink:    zoom.Link,
		ZMeet:    zoom.Meet,
		ZPass:    zoom.Pass,
		ZDate:    zoom.Date,
	}

	emailBody := template.New("emailtemplate.html")
	emailBody, err := emailBody.ParseFiles("templates/emailtemplate.html")
	if err != nil {
		return "", fmt.Errorf("parsing email template: %w", err)
	}
	var tpl bytes.Buffer
	if err := emailBody.Execute(&tpl, tags); err != nil {
		return "", fmt.Errorf("executing email template: %w", err)
	}
	return tpl.String(), nil
}

func sendEmail(email, message string) bool {
	smtpPort, err := strconv.Atoi(cfg.SMTPPort)
	if err != nil {
		log.Printf("invalid SMTP_PORT %q: %v; defaulting to 587", cfg.SMTPPort, err)
		smtpPort = 587
	}

	mailParam := gomail.NewMessage()
	mailParam.SetHeader("From", cfg.EmailSender)
	mailParam.SetHeader("To", email)
	mailParam.SetAddressHeader("Bcc", cfg.EmailBCC, cfg.EmailBCCNick)
	mailParam.SetHeader("Subject", cfg.EmailSubject)
	mailParam.SetBody("text/html", message)

	send := gomail.NewDialer(cfg.SMTPHost, smtpPort, cfg.SMTPUser, cfg.SMTPPass)
	if err := send.DialAndSend(mailParam); err != nil {
		log.Printf("failed to send email to %q: %v", email, err)
		return false
	}
	return true
}

// getHandler serves the registration form for GET requests.
// FIX M-6: returns 405 for any other method instead of an empty 200.
func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render(w, "templates/form.html", nil)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		getHandler(w, r)
		return
	}

	// FIX M-3: parse all form values once into msg; validatedInputs will be
	// built from msg fields (not re-read from r) so that transformations
	// (e.g. overbooked override, ToLower on email) are consistently applied.
	msg := &Inputs{
		Email:        strings.ToLower(r.FormValue("email")),
		FirstName:    r.PostFormValue("fname"),
		LastName:     r.PostFormValue("lname"),
		Area:         r.PostFormValue("area"),
		Group:        r.PostFormValue("group"),
		Function:     r.PostFormValue("function"),
		Gender:       r.PostFormValue("gender"),
		Local:        r.PostFormValue("local"),
		District:     r.PostFormValue("district"),
		Status:       r.PostFormValue("status"),
		PreferredDay: r.PostFormValue("prefday"),
	}

	// FIX C-5 (partial): the overbooked flag is set on msg.PreferredDay so
	// that Validate() catches it AND validatedInputs below uses the same value.
	size := checkSize(msg.PreferredDay)
	if size > 85 {
		msg.PreferredDay = "overbooked"
	}

	if !msg.Validate() {
		render(w, "templates/form.html", msg)
		return
	}

	// Build Entry from the already-parsed and transformed msg fields.
	validatedInputs := Entry{
		Email:        msg.Email,
		FirstName:    titler.String(strings.ToLower(msg.FirstName)),
		LastName:     titler.String(strings.ToLower(msg.LastName)),
		Area:         msg.Area,
		Group:        msg.Group,
		Function:     titler.String(msg.Function),
		Gender:       msg.Gender,
		Local:        titler.String(msg.Local),
		District:     strings.ToUpper(msg.District),
		Status:       msg.Status,
		PreferredDay: msg.PreferredDay, // uses the (possibly "overbooked") value from msg
	}

	if !record(validatedInputs) {
		render(w, "templates/registrationfailure.html", nil)
		return
	}

	// FIX M-1: handle template errors from createMessage.
	createdMessage, err := createMessage(
		validatedInputs.FirstName,
		validatedInputs.LastName,
		validatedInputs.Local,
		validatedInputs.District,
		validatedInputs.PreferredDay,
	)
	if err != nil {
		log.Printf("failed to create confirmation email for %q: %v", validatedInputs.Email, err)
		render(w, "templates/sendingfailure.html", nil)
		return
	}

	if sendEmail(validatedInputs.Email, createdMessage) {
		render(w, "templates/confirmation.html", nil)
	} else {
		render(w, "templates/sendingfailure.html", nil)
	}
}

func ReportsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		render(w, "templates/reports.html", nil)
		return
	}

	userInput := r.PostFormValue("passcode")
	// FIX C-2: use constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(userInput), []byte(cfg.AdminPasscode)) != 1 {
		render(w, "templates/incorrectpassword.html", nil)
		return
	}

	records := retrieveRecords()
	render(w, "templates/records.html", records)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	// FIX m-5: set explicit Content-Type.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Alive and well :)")
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	// FIX C-3: salt is now loaded from PING_SALT env var (see config.go).
	// FIX m-4: removed trailing space from Content-Type value.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	now := strconv.FormatInt(time.Now().Unix(), 10)
	// MD5 is used here only as a non-cryptographic uptime token, not for
	// security. The salt is loaded from env so it is no longer committed to source.
	w.Write([]byte(now + fmt.Sprintf("%x", md5.Sum([]byte(now+cfg.PingSalt)))))
}
