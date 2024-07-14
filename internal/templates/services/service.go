package templates

import (
	"context"
"errors"
"fmt"
"strings"
	"getnoti.com/internal/notifications/domain"
	templates"getnoti.com/internal/templates/domain"
	"getnoti.com/internal/templates/repos"
)

type TemplateService struct {
    repo repos.TemplateRepository
}

func NewTemplateService(repo repos.TemplateRepository) *TemplateService {
    return &TemplateService{
        repo: repo,
    }
}

func (s *TemplateService) GetContent(ctx context.Context, templateID string, variables []domain.TemplateVariable) (string, error) {
    // Get the template
    template, err := s.repo.GetTemplateByID(ctx, templateID)
    if err != nil {
        return "", err
    }

    if template == nil {
        return "", errors.New("template not found")
    }

    // Replace variables in the template
    content,err := s.replaceVariables(template, variables)

    return content, err
}

func (s *TemplateService) replaceVariables(template *templates.Template, variables []domain.TemplateVariable) (string, error) {
    content := template.Content
    variableMap := make(map[string]string)
    for _, v := range variables {
        variableMap[v.Key] = v.Value
    }

    missingVariables := []string{}
    for _, key := range template.Variables {
        if value, exists := variableMap[key]; exists {
            content = strings.ReplaceAll(content, "{{"+key+"}}", value)
        } else {
            missingVariables = append(missingVariables, key)
        }
    }

    if len(missingVariables) > 0 {
        return "", fmt.Errorf("missing variables: %s", strings.Join(missingVariables, ", "))
    }

    return content, nil
}



// example
// <mjml><mj-body><mj-section><mj-column><mj-image src=/assets/img/logo-small.png width=100px></mj-image><mj-divider border-color=#F45E43></mj-divider><mj-text color=#F45E43 font-family=helvetica font-size=20px>{{text}}</mj-text></mj-column></mj-section></mj-body></mjml>

// Html
// <!doctype html><html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office"><head><title></title><!--[if !mso]><!--><meta http-equiv="X-UA-Compatible" content="IE=edge"><!--<![endif]--><meta http-equiv="Content-Type" content="text/html; charset=UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><style type="text/css">#outlook a { padding:0; }
//           body { margin:0;padding:0;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%; }
//           table, td { border-collapse:collapse;mso-table-lspace:0pt;mso-table-rspace:0pt; }
//           img { border:0;height:auto;line-height:100%; outline:none;text-decoration:none;-ms-interpolation-mode:bicubic; }
//           p { display:block;margin:13px 0; }</style><!--[if mso]>
//         <noscript>
//         <xml>
//         <o:OfficeDocumentSettings>
//           <o:AllowPNG/>
//           <o:PixelsPerInch>96</o:PixelsPerInch>
//         </o:OfficeDocumentSettings>
//         </xml>
//         </noscript>
//         <![endif]--><!--[if lte mso 11]>
//         <style type="text/css">
//           .mj-outlook-group-fix { width:100% !important; }
//         </style>
//         <![endif]--><style type="text/css">@media only screen and (min-width:480px) {
//         .mj-column-per-100 { width:100% !important; max-width: 100%; }
//       }</style><style media="screen and (min-width:480px)">.moz-text-html .mj-column-per-100 { width:100% !important; max-width: 100%; }</style><style type="text/css">@media only screen and (max-width:480px) {
//       table.mj-full-width-mobile { width: 100% !important; }
//       td.mj-full-width-mobile { width: auto !important; }
//     }</style></head><body style="word-spacing:normal;"><div><!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" class="" style="width:600px;" width="600" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]--><div style="margin:0px auto;max-width:600px;"><table align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="width:100%;"><tbody><tr><td style="direction:ltr;font-size:0px;padding:20px 0;text-align:center;"><!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="" style="vertical-align:top;width:600px;" ><![endif]--><div class="mj-column-per-100 mj-outlook-group-fix" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"><table border="0" cellpadding="0" cellspacing="0" role="presentation" style="vertical-align:top;" width="100%"><tbody><tr><td align="center" style="font-size:0px;padding:10px 25px;word-break:break-word;"><table border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:collapse;border-spacing:0px;"><tbody><tr><td style="width:100px;"><img height="auto" src="/assets/img/logo-small.png" style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;" width="100"></td></tr></tbody></table></td></tr><tr><td align="center" style="font-size:0px;padding:10px 25px;word-break:break-word;"><p style="border-top:solid 4px #F45E43;font-size:1px;margin:0px auto;width:100%;"></p><!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" style="border-top:solid 4px #F45E43;font-size:1px;margin:0px auto;width:550px;" role="presentation" width="550px" ><tr><td style="height:0;line-height:0;"> &nbsp;
// </td></tr></table><![endif]--></td></tr><tr><td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;"><div style="font-family:helvetica;font-size:20px;line-height:1;text-align:left;color:#F45E43;">{{text}}</div></td></tr></tbody></table></div><!--[if mso | IE]></td></tr></table><![endif]--></td></tr></tbody></table></div><!--[if mso | IE]></td></tr></table><![endif]--></div></body></html>