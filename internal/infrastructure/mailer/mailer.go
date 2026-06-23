package mailer

import (
	"fmt"
	"log"
	"net/smtp"
)

type Mailer interface {
	SendEmail(to, subject, body string) error
	SendWelcomeCredentials(to, name, role, username, password string) error
	SendResetPasswordCode(to, code string) error
	SendCoachApplicationAlert(superadminEmail, coachName, coachEmail string) error
	SendExpiryWarning(to, coachName string, daysLeft int) error
	SendDeactivationNotice(to, coachName string) error
}

type SMTPMailer struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPMailer(host string, port int, user, pass, from string) Mailer {
	return &SMTPMailer{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

func (m *SMTPMailer) SendEmail(to, subject, body string) error {
	if m.host == "" || m.host == "localhost" {
		log.Printf("\n--- [CONSOLE EMAIL SENT] ---\nTo: %s\nSubject: %s\nBody: %s\n----------------------------\n", to, subject, body)
		return nil
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", m.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}

	err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("SMTP failed to send mail: %v\n", err)
		return err
	}
	return nil
}

func (m *SMTPMailer) SendWelcomeCredentials(to, name, role, username, password string) error {
	subject := "Evinizin Koçu - Hesabınız Oluşturuldu"
	body := fmt.Sprintf(`
		<h2>Merhaba %s,</h2>
		<p>Evinizin Koçu platformunda <strong>%s</strong> hesabınız başarıyla oluşturulmuştur.</p>
		<p>Giriş bilgileriniz aşağıdadır:</p>
		<ul>
			<li><strong>Kullanıcı Adı / E-posta:</strong> %s</li>
			<li><strong>Geçici Şifre:</strong> %s</li>
		</ul>
		<p>Güvenliğiniz için ilk girişinizde şifrenizi değiştirmeniz gerekmektedir.</p>
		<p>İyi çalışmalar dileriz.<br/>Evinizin Koçu Ekibi</p>
	`, name, role, username, password)
	return m.SendEmail(to, subject, body)
}

func (m *SMTPMailer) SendResetPasswordCode(to, code string) error {
	subject := "Evinizin Koçu - Şifre Sıfırlama Kodu"
	body := fmt.Sprintf(`
		<h2>Şifre Sıfırlama Talebi</h2>
		<p>Evinizin Koçu hesabınızın şifresini sıfırlamak için bir kod talep ettiniz.</p>
		<p>Şifre sıfırlama doğrulama kodunuz:</p>
		<h1 style="color:#0284c7; letter-spacing: 5px;">%s</h1>
		<p>Bu kod <strong>10 dakika</strong> süreyle geçerlidir. Eğer bu talebi siz yapmadıysanız lütfen bu e-postayı dikkate almayınız.</p>
		<p>Evinizin Koçu Ekibi</p>
	`, code)
	return m.SendEmail(to, subject, body)
}

func (m *SMTPMailer) SendCoachApplicationAlert(superadminEmail, coachName, coachEmail string) error {
	subject := "Yeni Koç Başvurusu Alındı"
	body := fmt.Sprintf(`
		<h2>Yeni Koç Başvurusu</h2>
		<p>Sistemde yeni bir koç başvurusu bulunmaktadır:</p>
		<ul>
			<li><strong>İsim Soyisim:</strong> %s</li>
			<li><strong>E-posta:</strong> %s</li>
		</ul>
		<p>Lütfen süper yönetici panelinden başvuruyu inceleyip onaylayınız veya reddediniz.</p>
	`, coachName, coachEmail)
	return m.SendEmail(superadminEmail, subject, body)
}

func (m *SMTPMailer) SendExpiryWarning(to, coachName string, daysLeft int) error {
	subject := "Evinizin Koçu - Yetki Süresi Uyarısı"
	body := fmt.Sprintf(`
		<h2>Sayın %s,</h2>
		<p>Evinizin Koçu platformundaki yetki sürenizin bitmesine <strong>%d gün</strong> kalmıştır.</p>
		<p>Hesabınızın pasif duruma düşmemesi için lütfen yöneticinizle iletişime geçiniz.</p>
		<p>Evinizin Koçu Ekibi</p>
	`, coachName, daysLeft)
	return m.SendEmail(to, subject, body)
}

func (m *SMTPMailer) SendDeactivationNotice(to, coachName string) error {
	subject := "Evinizin Koçu - Hesabınız Pasifleştirildi"
	body := fmt.Sprintf(`
		<h2>Sayın %s,</h2>
		<p>Evinizin Koçu platformundaki yetki süreniz sona erdiği için hesabınız otomatik olarak pasif duruma getirilmiştir.</p>
		<p>Tekrar aktif hale getirmek için lütfen yöneticinizle iletişime geçiniz.</p>
		<p>Evinizin Koçu Ekibi</p>
	`, coachName)
	return m.SendEmail(to, subject, body)
}
