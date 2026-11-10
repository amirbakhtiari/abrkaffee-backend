package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCORSMiddleware پیکربندی استاندارد CORS را برای اتصال فرانت‌اند و کلاینت‌ها فراهم می‌کند
func NewCORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// در محیط توسعه برای اتصال کلاینت‌ها از هر پورت و آدرسی
		AllowAllOrigins: true,

		// متدهای مجاز HTTP
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},

		// هدرهای مجاز ارسالی از سمت فرانت‌اند
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept",
			"Accept-Encoding",
			"Authorization",
			"X-Requested-With",
			"Idempotency-Key",
			"idempotency-key",
		},

		// هدرهایی که فرانت مجاز به خواندن آن‌ها از ریسپانس است
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},

		// طبق استاندارد مرورگر، زمانی که AllowAllOrigins برابر با true است، AllowCredentials باید false باشد
		AllowCredentials: false,

		// مدت زمان کش شدن درخواست Preflight (OPTIONS) در مرورگر
		MaxAge: 12 * time.Hour,
	})
}
