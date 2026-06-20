package worker

import "fmt"

func buildVerificationTemplate(username, verificationURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verify Your Email</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f7fa;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">

  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f7fa;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.06);">

          <tr>
            <td style="background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);padding:40px 48px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:28px;font-weight:700;letter-spacing:-0.5px;">
                Mengonten
              </h1>
              <p style="margin:8px 0 0;color:rgba(255,255,255,0.85);font-size:14px;">
                Video Clipping Platform
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:48px;">
              <h2 style="margin:0 0 8px;color:#1a1a2e;font-size:24px;font-weight:600;">
                Hi %s,
              </h2>
              <p style="margin:0 0 24px;color:#64748b;font-size:16px;line-height:1.6;">
                Thanks for signing up! Please verify your email address to activate your account and start using Mengonten.
              </p>

              <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center" style="padding:8px 0 32px;">
                    <a href="%s"
                       style="display:inline-block;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);color:#ffffff;text-decoration:none;font-size:16px;font-weight:600;padding:14px 48px;border-radius:10px;letter-spacing:0.3px;">
                      Verify Email Address
                    </a>
                  </td>
                </tr>
              </table>

              <p style="margin:0 0 8px;color:#94a3b8;font-size:13px;line-height:1.6;">
                Or copy and paste this link into your browser:
              </p>
              <p style="margin:0;color:#667eea;font-size:13px;word-break:break-all;">
                %s
              </p>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                This link expires in 24 hours.<br>
                If you didn't create an account, you can safely ignore this email.
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:16px 48px;text-align:center;">
              <p style="margin:0;color:#cbd5e1;font-size:11px;">
                &copy; 2026 Mengonten. All rights reserved.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>

</body>
</html>`, username, verificationURL, verificationURL)
}

func buildWelcomeTemplate(username string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Welcome to Mengonten</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f7fa;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">

  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f7fa;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.06);">

          <tr>
            <td style="background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);padding:40px 48px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:28px;font-weight:700;">
                Welcome to Mengonten!
              </h1>
            </td>
          </tr>

          <tr>
            <td style="padding:48px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:22px;font-weight:600;">
                You're all set, %s!
              </h2>
              <p style="margin:0 0 24px;color:#64748b;font-size:16px;line-height:1.6;">
                Your email has been verified successfully. You can now enjoy all features of Mengonten, including AI-powered video clipping.
              </p>

              <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f0fdf4;border-radius:12px;border:1px solid #bbf7d0;">
                <tr>
                  <td style="padding:20px 24px;">
                    <p style="margin:0;color:#166534;font-size:14px;line-height:1.6;">
                      <strong>What's next?</strong><br>
                      Submit a YouTube video URL and let our AI analyze the content to create the best clips automatically.
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="padding:16px 48px;text-align:center;">
              <p style="margin:0;color:#cbd5e1;font-size:11px;">
                &copy; 2026 Mengonten. All rights reserved.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>

</body>
</html>`, username)
}
