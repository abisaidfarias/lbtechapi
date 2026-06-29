package functions

import (
	"fmt"
	"html"
	"html/template"
)

const (
	EmailButtonColorPrimary   = "#8DB319"
	EmailButtonColorSecondary = "#0b1736"
)

func outlookButtonVMLWidth(label string) int {
	width := len([]rune(label))*9 + 48
	if width < 150 {
		return 150
	}
	if width > 340 {
		return 340
	}
	return width
}

// RenderOutlookEmailButton returns table-cell markup for a bulletproof email button
// (Outlook VML + modern clients anchor).
func RenderOutlookEmailButton(href, label, bgColor string) template.HTML {
	safeHref := html.EscapeString(href)
	safeLabel := html.EscapeString(label)
	safeBg := html.EscapeString(bgColor)
	width := outlookButtonVMLWidth(label)

	markup := fmt.Sprintf(`<td align="left" style="padding:0 12px 12px 0;vertical-align:top;">
<table role="presentation" border="0" cellspacing="0" cellpadding="0" style="border-collapse:separate;">
<tr>
<td align="center" bgcolor="%[3]s" style="background-color:%[3]s;border-radius:6px;mso-padding-alt:14px 24px;">
<!--[if mso]>
<v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word" href="%[1]s" style="height:44px;v-text-anchor:middle;width:%[4]dpx;" arcsize="14%%" stroke="f" fillcolor="%[3]s">
<w:anchorlock/>
<center style="color:#ffffff;font-family:Arial,sans-serif;font-size:14px;font-weight:bold;mso-line-height-rule:exactly;">%[2]s</center>
</v:roundrect>
<![endif]-->
<!--[if !mso]><!-->
<a href="%[1]s" target="_blank" rel="noopener" style="background-color:%[3]s;border-radius:6px;color:#ffffff;display:inline-block;font-family:Arial,Helvetica,sans-serif;font-size:14px;font-weight:bold;line-height:16px;padding:14px 24px;text-align:center;text-decoration:none;-webkit-text-size-adjust:none;mso-hide:all;">%[2]s</a>
<!--<![endif]-->
</td>
</tr>
</table>
</td>`, safeHref, safeLabel, safeBg, width)

	return template.HTML(markup)
}

func parseNotificationEmailTemplate(templatePath string) (*template.Template, error) {
	return template.New("notificationEmail").Funcs(template.FuncMap{
		"emailButton": RenderOutlookEmailButton,
	}).ParseFiles(templatePath)
}
