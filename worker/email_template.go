package worker

import (
	"fmt"
	"strings"
)

func buildVerificationTemplate(username, verificationURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verifikasi Email - Mengonten</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f7fa;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">

  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f7fa;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.06);">

          <tr>
            <td style="padding:36px 48px;text-align:center;">
              <img src="https://cdn-mengonten.tiroe.io/assets/logo_horizontal.png" alt="Mengonten" style="max-width:200px;height:auto;display:block;margin:0 auto;">
            </td>
          </tr>

          <tr>
            <td style="padding:48px;">
              <h2 style="margin:0 0 8px;color:#1a1a2e;font-size:24px;font-weight:600;">
                Hai %s,
              </h2>
              <p style="margin:0 0 24px;color:#64748b;font-size:16px;line-height:1.6;">
                Terima kasih telah mendaftar! Silakan verifikasi alamat email kamu untuk mengaktifkan akun dan mulai menggunakan Mengonten.
              </p>

              <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center" style="padding:8px 0 32px;">
                    <a href="%s"
                       style="display:inline-block;background:#FF0000;color:#ffffff;text-decoration:none;font-size:16px;font-weight:700;padding:14px 48px;border-radius:10px;letter-spacing:0.5px;">
                      VERIFIKASI EMAIL
                    </a>
                  </td>
                </tr>
              </table>

              <p style="margin:0 0 8px;color:#94a3b8;font-size:13px;line-height:1.6;">
                Atau salin tautan berikut ke browser kamu:
              </p>
              <p style="margin:0;color:#dc2626;font-size:13px;word-break:break-all;">
                %s
              </p>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                Tautan ini berlaku selama 24 jam.<br>
                Jika kamu tidak membuat akun, abaikan email ini.
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

func buildCheckoutTemplate(username, referenceID, planName, description string, amount, totalAmount float64, uniqueCode int, checkoutURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Checkout %s - Mengonten</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f7fa;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">

  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f7fa;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.06);">

          <tr>
            <td style="padding:36px 48px;text-align:center;">
              <img src="https://cdn-mengonten.tiroe.io/assets/logo_horizontal.png" alt="Mengonten" style="max-width:200px;height:auto;display:block;margin:0 auto;">
            </td>
          </tr>

          <tr>
            <td style="padding:48px;">
              <h2 style="margin:0 0 4px;color:#1a1a2e;font-size:24px;font-weight:600;">
                Hai %s,
              </h2>
              <p style="margin:0 0 8px;color:#64748b;font-size:16px;line-height:1.6;">
                Kami telah menerima pesanan kamu. Silakan lakukan pembayaran untuk melanjutkan.
              </p>

               <p style="margin:20px 0;color:#1a1a2e;font-size:16px;font-weight:700;">
                 #%s
               </p>

               <h3 style="margin:24px 0 12px;color:#1a1a2e;font-size:16px;font-weight:600;">
                 Detail Pesanan
              </h3>
              <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:14px;width:40%%;">Paket</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;font-weight:600;">%s</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:14px;width:40%%;">Deskripsi</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;">%s</td>
                </tr>
              </table>

              <h3 style="margin:24px 0 12px;color:#1a1a2e;font-size:16px;font-weight:600;">
                Tagihan
              </h3>
              <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:14px;width:40%%;">Harga Plan</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;font-weight:600;">Rp %s</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:14px;">Kode Unik</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;font-weight:600;">%d</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#fef2f2;border-top:2px solid #FF0000;color:#dc2626;font-size:14px;font-weight:700;">Total Bayar</td>
                  <td style="padding:12px 16px;background-color:#fef2f2;border-top:2px solid #FF0000;color:#dc2626;font-size:16px;font-weight:700;">Rp %s</td>
                </tr>
              </table>

              <table width="100%%" cellpadding="0" cellspacing="0" style="margin:32px 0 0;">
                <tr>
                  <td align="center">
                    <a href="%s"
                       style="display:inline-block;background:#FF0000;color:#ffffff;text-decoration:none;font-size:16px;font-weight:700;padding:16px 48px;border-radius:10px;letter-spacing:0.5px;">
                       LANJUTKAN PEMBAYARAN
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                Jika kamu memiliki pertanyaan, hubungi tim support kami.<br>
                Mohon tidak membalas email ini.
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
</html>`, referenceID, username, referenceID, planName, description, formatPrice(amount), int(uniqueCode), formatPrice(totalAmount), checkoutURL)
}

func formatPrice(price float64) string {
	s := fmt.Sprintf("%.0f", price)
	n := len(s)
	if n <= 3 {
		return s
	}
	var parts []string
	for i := n; i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:i]}, parts...)
	}
	return strings.Join(parts, ".")
}

func buildWelcomeTemplate(username string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Selamat Datang - Mengonten</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f7fa;font-family:'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">

  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f7fa;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.06);">

          <tr>
            <td style="padding:36px 48px;text-align:center;">
              <img src="https://cdn-mengonten.tiroe.io/assets/logo_horizontal.png" alt="Mengonten" style="max-width:200px;height:auto;display:block;margin:0 auto;">
            </td>
          </tr>

          <tr>
            <td style="padding:48px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:22px;font-weight:600;">
                Selamat datang, %s!
              </h2>
              <p style="margin:0 0 24px;color:#64748b;font-size:16px;line-height:1.6;">
                Email kamu berhasil diverifikasi. Kamu sekarang dapat menikmati semua fitur Mengonten, termasuk kliping video berbasis AI.
              </p>

              <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#fef2f2;border-radius:12px;border:1px solid #fecaca;">
                <tr>
                  <td style="padding:20px 24px;">
                    <p style="margin:0;color:#dc2626;font-size:14px;line-height:1.6;">
                      <strong>Langkah selanjutnya?</strong><br>
                      Kirim URL YouTube dan biarkan AI kami menganalisis konten untuk membuat klip terbaik secara otomatis.
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                Jika kamu memiliki pertanyaan, hubungi tim support kami.
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
</html>`, username)
}
