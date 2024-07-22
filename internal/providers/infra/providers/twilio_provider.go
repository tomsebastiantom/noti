package providers

import (
    "context"
    "fmt"
    "getnoti.com/internal/providers/dtos"
    "github.com/twilio/twilio-go"
    twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioProvider struct {
    client *twilio.RestClient
}

func NewTwilioProvider(accountSid, authToken string) *TwilioProvider {
    client := twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: accountSid,
        Password: authToken,
    })
    return &TwilioProvider{
        client: client,
    }
}

func (p *TwilioProvider) CreateClient(ctx context.Context, credentials map[string]interface{}) error {
    accountSid, ok := credentials["account_sid"].(string)
    if !ok {
        return fmt.Errorf("invalid account_sid in credentials")
    }
    authToken, ok := credentials["auth_token"].(string)
    if !ok {
        return fmt.Errorf("invalid auth_token in credentials")
    }

    p.client = twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: accountSid,
        Password: authToken,
    })
    return nil
}

func (p *TwilioProvider) SendNotification(ctx context.Context, req dtos.SendNotificationRequest) dtos.SendNotificationResponse {
    if p.client == nil {
        return dtos.SendNotificationResponse{Success: false, Message: "Client not initialized"}
    }

    switch req.Channel {
    case "SMS":
        params := &twilioApi.CreateMessageParams{}
        params.SetTo(req.Receiver)
        params.SetFrom(req.Sender)
        params.SetBody(req.Content)
        
        resp, err := p.client.Api.CreateMessage(params)
        if err != nil {
            return dtos.SendNotificationResponse{Success: false, Message: err.Error()}
        }
        return dtos.SendNotificationResponse{Success: true, Message: *resp.Sid}

    case "Call":
        params := &twilioApi.CreateCallParams{}
        params.SetTo(req.Receiver)
        params.SetFrom(req.Sender)
        params.SetUrl("http://demo.twilio.com/docs/voice.xml")
        
        resp, err := p.client.Api.CreateCall(params)
        if err != nil {
            return dtos.SendNotificationResponse{Success: false, Message: err.Error()}
        }
        return dtos.SendNotificationResponse{Success: true, Message: *resp.Sid}

    default:
        return dtos.SendNotificationResponse{Success: false, Message: "Unsupported notification channel"}
    }
}
