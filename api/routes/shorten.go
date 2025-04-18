package routes

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Vibhuair20/shortern-url-fiber-redis/database"
	"github.com/Vibhuair20/shortern-url-fiber-redis/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skip2/go-qrcode"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short_url"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL               string        `json:"url"`
	CustomShortExpiry string        `json:"short_url"`
	SimpleQr          string        `json:"qr_code_url"`
	Expiry            time.Duration `json:"expiry"`
	XRateRemaining    int           `json:"rate_limit"`
	XRateLimitReset   time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(c *fiber.Ctx) error {

	body := new(request)

	// body request from the user to send the request
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	//implement the rate limiter
	//check the ip of user
	// if the ip has been saved in our database
	// if it has already used the servoce then it decrements from the rate limiter

	// call it 10 times in the time period of 10 minutes

	// implement rate limiting
	r2 := database.CreateClient(1)
	defer r2.Close()
	val, err := r2.Get(database.Ctx, c.IP()).Result()

	//with redis there is eithe set and get
	// set the api quota
	// in the last 30 minutes the ip has not used the service
	if err != nil && err != redis.Nil {
		fmt.Println("Error getting value from redis", err)
	}

	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		//if we find the ip has used the service in the last 30 minutes
		val, err = r2.Get(database.Ctx, c.IP()).Result()
		if err != nil {
			fmt.Println("error getting rate limit from redis: ", err)
		}
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":           "rate limit excedded",
				"rate_limit_rest": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	// check i fthe imput send by the user is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}
	//check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "nhi bhai not possible"})
	}
	//enforce https, ssl
	body.URL = helpers.EnforceHTTP(body.URL)

	// 1. send the custom url link
	// 2. has some other user did already use this custom url link
	// 3. generate the custom url link with uuid

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	//generate the qr of shortned url
	shortURL := os.Getenv("DOMAIN") + "/" + id
	// fetch the qr from the cache first
	qrImage, err := r.Get(database.Ctx, "qr:"+id).Result()
	if err == redis.Nil {
		png, err := qrcode.Encode(shortURL, qrcode.Medium, 256)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "i am sorry there is some error can't generate the qr code",
			})
		}
		//convert the png to base 64
		var buf bytes.Buffer
		encoder := base64.NewEncoder(base64.StdEncoding, &buf)
		encoder.Write(png)
		encoder.Close()

		qrImage = "data:image/png;base64," + buf.String()

		// caching the qr image
		err = r.Set(database.Ctx, "qr:"+id, qrImage, body.Expiry*3600*time.Second).Err()
		if err != nil {
			fmt.Println("error encoding qr:", err)
		}
	}

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "url custom short is already in use",
		})
	}
	// checkin gthe expiry

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unale to connect to the  server",
		})
	}

	// making the request
	resp := response{
		URL:               body.URL,
		CustomShortExpiry: "",
		Expiry:            body.Expiry,
		SimpleQr:          qrImage,
		XRateRemaining:    10,
		XRateLimitReset:   30,
	}

	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShortExpiry = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
