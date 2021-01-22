package report

import (
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
)

type UnbanRequestState int

const (
	UnbanRequestStatePending UnbanRequestState = iota
	UnbanRequestStateDeclined
	UnbanRequestStateAccepted
)

type UnbanRequest struct {
	ID               snowflake.ID      `json:"id"`
	UserID           string            `json:"user_id"`
	GuildID          string            `json:"guild_id"`
	UserTag          string            `json:"user_tag"`
	Message          string            `json:"message"`
	Status           UnbanRequestState `json:"status"`
	ProcessedBy      string            `json:"processed_by"`
	Processed        time.Time         `json:"processed"`
	ProcessedMessage string            `json:"processed_message"`
}

func (r *UnbanRequest) Validate() error {
	if r.GuildID == "" {
		return errors.New("invalid guild ID")
	}
	if r.Message == "" {
		return errors.New("message must be provided")
	}

	return nil
}
