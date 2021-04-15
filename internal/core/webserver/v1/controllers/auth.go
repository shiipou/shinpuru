package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/internal/core/database"
	"github.com/zekroTJA/shinpuru/internal/core/webserver/auth"
	"github.com/zekroTJA/shinpuru/internal/core/webserver/v1/models"
	"github.com/zekroTJA/shinpuru/internal/util"
	"github.com/zekroTJA/shinpuru/internal/util/static"
	discordoauth "github.com/zekroTJA/shinpuru/pkg/discordoauth/v2"
)

type AuthController struct {
	discordOAuth *discordoauth.DiscordOAuth
	rth          auth.RefreshTokenHandler
	ath          auth.AccessTokenHandler
}

func (c *AuthController) Setup(container di.Container, router fiber.Router) {
	c.discordOAuth = container.Get(static.DiDiscordOAuthModule).(*discordoauth.DiscordOAuth)
	c.rth = container.Get(static.DiAuthRefreshTokenHandler).(auth.RefreshTokenHandler)
	c.ath = container.Get(static.DiAuthAccessTokenHandler).(auth.AccessTokenHandler)

	router.Get("/login", c.discordOAuth.HandlerInit)
	router.Get("/oauthcallback", c.discordOAuth.HandlerCallback)
	router.Get("/accesstoken", c.GetAccessToken)
}

func (c *AuthController) GetAccessToken(ctx *fiber.Ctx) error {
	refreshToken := ctx.Cookies(static.RefreshTokenCookieName)
	if refreshToken == "" {
		return fiber.ErrUnauthorized
	}

	ident, err := c.rth.ValidateRefreshToken(refreshToken)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		util.Log.Error("WEBSERVER :: failed validating refresh token:", err)
	}
	if ident == "" {
		return fiber.ErrUnauthorized
	}

	token, expires, err := c.ath.GetAccessToken(ident)
	if err != nil {
		return err
	}

	return ctx.JSON(&models.AccessTokenResponse{
		Token:   token,
		Expires: expires,
	})
}