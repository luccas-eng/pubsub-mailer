package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type from struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type to struct {
	Email string `json:"email"`
	Name  string `json:"username"`
}

type personalization struct {
	To        []to   `json:"to"`
	CartPrice string `json:"cart_price"`
}

type sendGridJSON struct {
	From             from              `json:"from"`
	Personalizations []personalization `json:"personalizations"`
	TemplateID       string            `json:"template_id"`
}

func main() {
	//Implementation at PubSub as Cloud Function func main() just to avoid linter errors
}

//NewEventListener ...
func NewEventListener(ctx context.Context, m PubSubMessage) error {

	var (
		err        error
		discount   bool
		templateID string
	)

	data := string(m.Data)

	session := strings.Split(data, ",")

	userID := session[0]
	userSessionID := strings.Split(userID, "-")
	userSessionID = userSessionID[2:len(userSessionID)]

	cart := session[1]
	cart = cart[1 : len(cart)-1]

	price, err := strconv.ParseFloat(cart, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseFloat(): %w", err)
	}

	if price > 30 {
		discount = true
		templateID = "d-8f700b6a69454b769a29f80aaf0cfb7c"
	} else {
		templateID = "d-b2e10b0be0de42848d9495b1cb384c7a"
	}

	getUserInfo := func(userID string) (mailAddress, userName string) {
		return "lucas@itguru.com.br", "Lucas Costa"
	}

	getAuthorization := func(userID string) bool {
		return true
	}

	userEmail, userName := getUserInfo(userID)
	isAuthorized := getAuthorization(userID)

	if isAuthorized {

		sendEmail := func(userEmail, userName string, isDiscount bool, cartprice float64) error {

			newFrom := from{
				Email: "contato@econsmart.com",
				Name:  "Eco'N'Smart",
			}

			newTo := to{
				Email: userEmail,
				Name:  userName,
			}
			arrayTo := []to{newTo}

			personalization := []personalization{
				{
					To:        arrayTo,
					CartPrice: strconv.FormatFloat(cartprice, 'f', -1, 64),
				},
			}

			sendgridObj := sendGridJSON{
				From:             newFrom,
				Personalizations: personalization,
				TemplateID:       templateID,
			}

			bytesRepresentation, err := json.Marshal(sendgridObj)
			if err != nil {
				return fmt.Errorf("json.Marshal(): %w", err)
			}

			// For debugging proposal
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, bytesRepresentation, "", "\t")
			if err != nil {
				return fmt.Errorf("json.Indent(): %w", err)
			}

			req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(bytesRepresentation))

			if err != nil {
				return fmt.Errorf("http.NewRequest(): %w", err)
			}

			req.Header.Add("Authorization", "Bearer "+os.Getenv("SENDGRID_API_KEY"))
			req.Header.Add("content-type", "application/json")

			res, err := http.DefaultClient.Do(req)

			if err != nil {
				return fmt.Errorf("http.DefaultClient.Do(): %w", err)
			}

			if res.StatusCode == 202 {
				fmt.Println("HTTP 202")
				return nil
			}

			return fmt.Errorf("Error sending email Status Code: %d", res.StatusCode)
		}

		if err = sendEmail(userEmail, userName, discount, price); err != nil {
			return fmt.Errorf("sendEmail(): %w", err)
		}
	}

	fmt.Println("Email Enviado com Sucesso.")
	return nil

}
