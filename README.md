# smtpx

A simple smtp client to send email(support Chinese). Work well with qq, 163, mac mail app, google mail.

## example

```go
	/*
	   a demo to send mail
	*/
	host := "smtp.exmail.qq.com" //  smtp server host
	port := 465    // smtp server port
	email := ""    // your email
	password := "" // your password
	toEmail := ""  // to email

	sender := NewSender(host, port, "your name", email, password)

	letterInstance := NewLetter()
	letterInstance.AddReceivers(toEmail)
	letterInstance.AddCarbonCopy("") // add cc address
	letterInstance.SetSubject("Test")
	letterInstance.SetContent("This is a test mail！")

	// attachment 2
	attachment1, err := ioutil.ReadFile("")
	if err != nil {
		return
	}
	letterInstance.AddAttachment("",  attachment1)

	// attachment 1
	attachment2, err := ioutil.ReadFile("")
	if err != nil {
		return
	}
	letterInstance.AddAttachment("", attachment2)

	letterInstance.Dump()   // you can dump content

	err = sender.Send(letterInstance)

```
