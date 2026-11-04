package newsletter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"strconv"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/mail"
)

type Confirmation struct {
	Project    models.Project
	Email, URL string
}

func (m Confirmation) Envelope() mail.Envelope {
	return mail.Envelope{From: m.Project.FromEmail, To: []string{m.Email}, Subject: "Confirme sua inscrição em " + m.Project.Name}
}
func (m Confirmation) Content() mail.Content { return mail.React("confirmation", m) }

type Newsletter struct {
	Project                        models.Project
	Email, Subject, UnsubscribeURL string
	Body                           template.HTML
}

func (m Newsletter) Envelope() mail.Envelope {
	return mail.Envelope{From: m.Project.FromEmail, To: []string{m.Email}, Subject: m.Subject}
}
func (m Newsletter) Content() mail.Content { return mail.React("campaign", m) }

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func UnsubscribeToken(key []byte, id int64) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte("unsubscribe:" + strconv.FormatInt(id, 10)))
	return hex.EncodeToString(m.Sum(nil))
}
