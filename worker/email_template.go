package worker

import (
	"fmt"
	"net/url"
	"strings"
)

func buildVerificationTemplate(verificationURL string) string {
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
                Verifikasi Email
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
</html>`, verificationURL, verificationURL)
}

func buildOTPTemplate(otp string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Kode OTP Login - Mengonten</title>
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
                Kode Login
              </h2>
              <p style="margin:0 0 24px;color:#64748b;font-size:16px;line-height:1.6;">
                Gunakan kode OTP berikut untuk masuk ke akun Mengonten kamu. Kode ini berlaku selama 5 menit.
              </p>

              <table width="100%%" cellpadding="0" cellspacing="0" style="margin:24px 0;">
                <tr>
                  <td align="center" style="background-color:#fef2f2;border-radius:12px;border:1px solid #fecaca;padding:24px;">
                    <p style="margin:0;color:#dc2626;font-size:36px;font-weight:700;letter-spacing:8px;font-family:monospace;">
                      %s
                    </p>
                  </td>
                </tr>
              </table>

              <p style="margin:0;color:#94a3b8;font-size:13px;line-height:1.6;">
                Jika kamu tidak meminta login, abaikan email ini.
              </p>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                Kode OTP berlaku selama 5 menit. Jangan bagikan kode ini kepada siapa pun.
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
</html>`, otp)
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

func paymentMethodDisplay(method string) string {
	switch method {
	case "qris":
		return "QRIS"
	case "bri_va":
		return "Virtual Account BRI"
	case "bni_va":
		return "Virtual Account BNI"
	case "cimb_niaga_va":
		return "Virtual Account CIMB Niaga"
	case "sampoerna_va":
		return "Virtual Account Bank Sampoerna"
	case "bnc_va":
		return "Virtual Account BNC"
	case "maybank_va":
		return "Virtual Account Maybank"
	case "permata_va":
		return "Virtual Account Permata"
	case "atm_bersama_va":
		return "Virtual Account ATM Bersama"
	case "artha_graha_va":
		return "Virtual Account Artha Graha"
	default:
		return method
	}
}

func buildTransactionTemplate(referenceID, planName string, amount float64, paymentMethod, paymentNumber, expiredAt, userName string) string {
	paymentMethodDisplay := paymentMethodDisplay(paymentMethod)

	var paymentInstructions string
	if paymentMethod == "qris" {
		qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", url.QueryEscape(paymentNumber))
		paymentInstructions = fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#fef2f2;border-radius:12px;border:1px solid #fecaca;margin:16px 0;">
                <tr>
                  <td style="padding:20px 24px;">
                    <p style="margin:0 0 8px;color:#dc2626;font-size:14px;font-weight:700;">Cara Pembayaran QRIS</p>
                    <ol style="margin:0;padding-left:20px;color:#64748b;font-size:13px;line-height:1.8;">
                      <li>Buka aplikasi GoPay, ShopeePay, OVO, atau mobile banking yang mendukung QRIS.</li>
                      <li>Pilih menu Scan / Bayar QR.</li>
                      <li>Scan kode QR di bawah ini:</li>
                    </ol>
                    <div style="text-align:center;margin:16px 0;">
                      <img src="%s" alt="QR Code" style="display:inline-block;max-width:200px;border-radius:8px;">
                    </div>
                    <p style="margin:12px 0 0;color:#dc2626;font-size:13px;font-weight:600;text-align:center;">Total Pembayaran: Rp %s</p>
                  </td>
                </tr>
              </table>`, qrURL, formatPriceIDR(amount))
	} else {
		paymentInstructions = fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#fef2f2;border-radius:12px;border:1px solid #fecaca;margin:16px 0;">
                <tr>
                  <td style="padding:20px 24px;">
                    <p style="margin:0 0 8px;color:#dc2626;font-size:14px;font-weight:700;">Cara Pembayaran %s</p>
                    <ol style="margin:0;padding-left:20px;color:#64748b;font-size:13px;line-height:1.8;">
                      <li>Login ke mobile banking atau ATM bank terkait.</li>
                      <li>Pilih menu Transfer ke Virtual Account.</li>
                      <li>Masukkan nomor Virtual Account berikut:</li>
                    </ol>
                    <p style="margin:12px 0 0;padding:12px;background:#ffffff;border-radius:8px;font-family:monospace;font-size:16px;color:#dc2626;font-weight:700;text-align:center;letter-spacing:2px;">%s</p>
                    <p style="margin:12px 0 0;color:#dc2626;font-size:13px;font-weight:600;text-align:center;">Total Pembayaran: Rp %s</p>
                  </td>
                </tr>
              </table>`, paymentMethodDisplay, paymentNumber, formatPriceIDR(amount))
	}

	if userName == "" {
		userName = "Hai"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Pesanan Dibuat - Mengonten</title>
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
            <td style="padding:0 48px 24px;">
              <h2 style="margin:0 0 8px;color:#1a1a2e;font-size:22px;font-weight:600;">
                %s,
              </h2>
              <p style="margin:0 0 16px;color:#64748b;font-size:15px;line-height:1.6;">
                Terima kasih, pesanan langganan Mengonten Anda telah berhasil dibuat dan saat ini sedang <strong>menunggu pembayaran</strong>.
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:0 48px;">
              <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;width:40%%;">Nomor Pesanan</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;font-weight:700;">#%s</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;">Status</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;"><span style="background:#fef2f2;color:#dc2626;padding:4px 12px;border-radius:4px;font-size:12px;font-weight:600;">Menunggu Pembayaran</span></td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;">Produk</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;">%s</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;">Batas Pembayaran</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#dc2626;font-size:14px;font-weight:600;">%s</td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="padding:24px 48px 16px;">
              <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;">Subtotal</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;text-align:right;">Rp %s</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#fef2f2;border-top:2px solid #FF0000;color:#dc2626;font-size:14px;font-weight:700;">Total Pembayaran</td>
                  <td style="padding:12px 16px;background-color:#fef2f2;border-top:2px solid #FF0000;color:#dc2626;font-size:16px;font-weight:700;text-align:right;">Rp %s</td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="padding:0 48px 8px;">
              <h3 style="margin:0 0 8px;color:#1a1a2e;font-size:14px;font-weight:600;">Metode Pembayaran</h3>
              <p style="margin:0 0 4px;color:#64748b;font-size:13px;">%s</p>
              %s
            </td>
          </tr>

          <tr>
            <td style="padding:8px 48px 24px;">
              <p style="margin:0;color:#94a3b8;font-size:12px;line-height:1.6;font-style:italic;">
                Pembayaran akan diverifikasi secara otomatis dalam beberapa menit setelah transaksi berhasil.
              </p>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                Apabila memiliki pertanyaan, silakan hubungi layanan pelanggan kami.
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
</html>`, userName, referenceID, planName, expiredAt, formatPriceIDR(amount), formatPriceIDR(amount), paymentMethodDisplay, paymentInstructions)
}

func formatPriceIDR(price float64) string {
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

func buildReminderTemplate(planID, planName, expiredAt, daysLeft, subject string, price float64) string {
	var message string
	switch daysLeft {
	case "7":
		message = "Langganan Anda akan berakhir dalam <strong>7 hari</strong>. Jangan lewatkan layanan Mengonten Anda!"
	case "3":
		message = "Langganan Anda akan berakhir dalam <strong>3 hari</strong>. Segera perpanjang untuk terus menikmati fitur lengkap."
	case "1":
		message = "Langganan Anda akan <strong>berakhir besok!</strong> Perpanjang sekarang agar tidak kehilangan akses."
	default:
		message = fmt.Sprintf("Langganan Anda akan berakhir dalam %s hari. Segera perpanjang.", daysLeft)
	}

	checkoutURL := fmt.Sprintf("https://mengonten.tiroe.io/workspace/checkout?plan_id=%s&plan=%s&price=%.0f",
		planID, url.QueryEscape(planName), price)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
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
            <td style="padding:0 48px 24px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:22px;font-weight:600;">
                %s
              </h2>
              <p style="margin:0 0 16px;color:#64748b;font-size:15px;line-height:1.6;">
                %s
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:0 48px;">
              <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;width:40%%;">Paket</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#1a1a2e;font-size:14px;font-weight:600;">%s</td>
                </tr>
                <tr>
                  <td style="padding:12px 16px;background-color:#f8fafc;border-bottom:1px solid #e2e8f0;color:#64748b;font-size:13px;">Berakhir pada</td>
                  <td style="padding:12px 16px;border-bottom:1px solid #e2e8f0;color:#dc2626;font-size:14px;font-weight:600;">%s</td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="padding:24px 48px 32px;">
              <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center">
                    <a href="%s"
                       style="display:inline-block;background:#FF0000;color:#ffffff;text-decoration:none;font-size:16px;font-weight:700;padding:14px 48px;border-radius:10px;letter-spacing:0.5px;">
                       PERPANJANG SEKARANG
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="background-color:#f8fafc;padding:24px 48px;border-top:1px solid #e2e8f0;">
              <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center;line-height:1.6;">
                Setelah masa aktif berakhir, akun Anda akan beralih ke paket Free secara otomatis.<br>
                Jika ada pertanyaan, hubungi tim support kami.
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
</html>`, subject, subject, message, planName, expiredAt, checkoutURL)
}
