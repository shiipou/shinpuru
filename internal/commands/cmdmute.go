package commands

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/zekroTJA/shinpuru/internal/core/database"
	"github.com/zekroTJA/shinpuru/internal/shared"
	"github.com/zekroTJA/shinpuru/internal/util/acceptmsg"
	"github.com/zekroTJA/shinpuru/internal/util/imgstore"
	"github.com/zekroTJA/shinpuru/internal/util/mute"
	"github.com/zekroTJA/shinpuru/internal/util/report"
	"github.com/zekroTJA/shinpuru/internal/util/snowflakenodes"
	"github.com/zekroTJA/shinpuru/internal/util/static"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/internal/util"
)

type CmdMute struct {
}

func (c *CmdMute) GetInvokes() []string {
	return []string{"mute", "m", "silence", "unmute", "um", "unsilence"}
}

func (c *CmdMute) GetDescription() string {
	return "Mute members in text channels"
}

func (c *CmdMute) GetHelp() string {
	return "`mute setup (<roleResolvable>)` - creates (or uses given) mute role and sets this role in every channel as muted\n" +
		"`mute <userResolvable>` - mute/unmute a user\n" +
		"`mute list` - display muted users on this guild\n" +
		"`mute` - display currently set mute role"
}

func (c *CmdMute) GetGroup() string {
	return GroupModeration
}

func (c *CmdMute) GetDomainName() string {
	return "sp.guild.mod.mute"
}

func (c *CmdMute) GetSubPermissionRules() []SubPermission {
	return nil
}

func (c *CmdMute) Exec(args *CommandArgs) error {
	if len(args.Args) < 1 {
		return c.displayMuteRole(args)
	}

	switch args.Args[0] {
	case "setup":
		return c.setup(args)
	case "list":
		return c.list(args)
	default:
		return c.muteUnmute(args)
	}
}

func (c *CmdMute) setup(args *CommandArgs) error {
	var muteRole *discordgo.Role
	var err error

	desc := "Following, a rolen with the name `shinpuru-muted` will be created *(if not existend yet)* and set as mute role."

	if len(args.Args) > 1 {
		muteRole, err = util.FetchRole(args.Session, args.Guild.ID, args.Args[1])
		if err != nil {
			return util.SendEmbedError(args.Session, args.Channel.ID,
				"Role could not be fetched by passed identifier.").
				DeleteAfter(8 * time.Second).Error()
		}

		desc = fmt.Sprintf("Follwoing, the role %s will be set as mute role.", muteRole.Mention())
	}

	acmsg := &acceptmsg.AcceptMessage{
		Session: args.Session,
		Embed: &discordgo.MessageEmbed{
			Color: static.ColorEmbedDefault,
			Title: "Warning",
			Description: desc + " Also, all channels *(which the bot has access to)* will be permission-overwritten that " +
				"members with this role will not be able to write in these channels anymore.",
		},
		UserID:         args.User.ID,
		DeleteMsgAfter: true,
		AcceptFunc: func(msg *discordgo.Message) {
			if muteRole == nil {
				for _, r := range args.Guild.Roles {
					if r.Name == static.MutedRoleName {
						muteRole = r
					}
				}
			}

			if muteRole == nil {
				muteRole, err = args.Session.GuildRoleCreate(args.Guild.ID)
				if err != nil {
					util.SendEmbedError(args.Session, args.Channel.ID,
						"Failed creating mute role: ```\n"+err.Error()+"\n```").
						DeleteAfter(15 * time.Second).Error()
					return
				}

				muteRole, err = args.Session.GuildRoleEdit(args.Guild.ID, muteRole.ID,
					static.MutedRoleName, 0, false, 0, false)
				if err != nil {
					util.SendEmbedError(args.Session, args.Channel.ID,
						"Failed editing mute role: ```\n"+err.Error()+"\n```").
						DeleteAfter(15 * time.Second).Error()
					return
				}
			}

			err := args.CmdHandler.db.SetMuteRole(args.Guild.ID, muteRole.ID)
			if err != nil {
				util.SendEmbedError(args.Session, args.Channel.ID,
					"Failed setting mute role in database: ```\n"+err.Error()+"\n```").
					DeleteAfter(15 * time.Second).Error()
				return
			}

			err = mute.SetupChannels(args.Session, args.Guild.ID, muteRole.ID)
			if err != nil {
				util.SendEmbedError(args.Session, args.Channel.ID,
					"Failed updating channels: ```\n"+err.Error()+"\n```").
					DeleteAfter(15 * time.Second).Error()
				return
			}

			util.SendEmbed(args.Session, args.Channel.ID,
				"Set up mute role and edited channel permissions.\nMaybe you need to increase the "+
					"position of the role to override other roles permission settings.",
				"", static.ColorEmbedUpdated).
				DeleteAfter(15 * time.Second).Error()
		},
		DeclineFunc: func(msg *discordgo.Message) {
			util.SendEmbedError(args.Session, args.Channel.ID,
				"Setup canceled.").
				DeleteAfter(8 * time.Second).Error()
		},
	}

	_, err = acmsg.Send(args.Channel.ID)
	return err
}

func (c *CmdMute) muteUnmute(args *CommandArgs) error {
	victim, err := util.FetchMember(args.Session, args.Guild.ID, args.Args[0])
	if err != nil {
		return util.SendEmbedError(args.Session, args.Channel.ID,
			"Could not fetch any user by the passed resolvable.").
			DeleteAfter(8 * time.Second).Error()
	}

	if victim.User.ID == args.User.ID {
		return util.SendEmbedError(args.Session, args.Channel.ID,
			"You can not mute yourself...").
			DeleteAfter(8 * time.Second).Error()
	}

	muteRoleID, err := args.CmdHandler.db.GetMuteRoleGuild(args.Guild.ID)
	if database.IsErrDatabaseNotFound(err) {
		return util.SendEmbedError(args.Session, args.Channel.ID,
			"Mute command is not set up. Please enter the command `mute setup`.").
			DeleteAfter(8 * time.Second).Error()
	} else if err != nil {
		return err
	}

	repType := util.IndexOfStrArray("MUTE", static.ReportTypes)
	repID := snowflakenodes.NodesReport[repType].Generate()

	var roleExists bool
	for _, r := range args.Guild.Roles {
		if r.ID == muteRoleID && !roleExists {
			roleExists = true
		}
	}
	if !roleExists {
		return util.SendEmbedError(args.Session, args.Channel.ID,
			"Mute role does not exist on this guild. Please enter `mute setup`.").
			DeleteAfter(8 * time.Second).Error()
	}

	var victimIsMuted bool
	for _, rID := range victim.Roles {
		if rID == muteRoleID && !victimIsMuted {
			victimIsMuted = true
		}
	}
	if victimIsMuted {
		err := args.Session.GuildMemberRoleRemove(args.Guild.ID, victim.User.ID, muteRoleID)
		if err != nil {
			return err
		}
		emb := &discordgo.MessageEmbed{
			Title: "Case " + repID.String(),
			Color: static.ReportColors[repType],
			Fields: []*discordgo.MessageEmbedField{
				{
					Inline: true,
					Name:   "Executor",
					Value:  fmt.Sprintf("<@%s>", args.User.ID),
				},
				{
					Inline: true,
					Name:   "Victim",
					Value:  fmt.Sprintf("<@%s>", victim.User.ID),
				},
				{
					Name:  "Type",
					Value: "UNMUTE",
				},
				{
					Name:  "Description",
					Value: "MANUAL UNMUTE",
				},
			},
			Timestamp: time.Unix(repID.Time()/1000, 0).Format("2006-01-02T15:04:05.000Z"),
		}
		args.Session.ChannelMessageSendEmbed(args.Channel.ID, emb)
		if modlogChan, err := args.CmdHandler.db.GetGuildModLog(args.Guild.ID); err == nil {
			args.Session.ChannelMessageSendEmbed(modlogChan, emb)
		}
		dmChan, err := args.Session.UserChannelCreate(victim.User.ID)
		if err == nil {
			args.Session.ChannelMessageSendEmbed(dmChan.ID, emb)
		}
		return err
	}

	err = args.Session.GuildMemberRoleAdd(args.Guild.ID, victim.User.ID, muteRoleID)
	if err != nil {
		return err
	}

	repMsg := strings.Join(args.Args[1:], " ")

	var attachment string
	repMsg, attachment = imgstore.ExtractFromMessage(repMsg, args.Message.Attachments)
	if attachment != "" {
		img, err := imgstore.DownloadFromURL(attachment)
		if err == nil && img != nil {
			err = args.CmdHandler.st.PutObject(static.StorageBucketImages, img.ID.String(),
				bytes.NewReader(img.Data), int64(img.Size), img.MimeType)
			if err != nil {
				return err
			}
			attachment = img.ID.String()
		}
	}

	rep, err := shared.PushMute(
		args.Session,
		args.CmdHandler.db,
		args.CmdHandler.config.WebServer.PublicAddr,
		args.Guild.ID,
		args.User.ID,
		victim.User.ID,
		strings.Join(args.Args[1:], " "),
		attachment,
		muteRoleID)

	if err != nil {
		err = util.SendEmbedError(args.Session, args.Channel.ID,
			"Failed creating report: ```\n"+err.Error()+"\n```").
			Error()
	} else {
		_, err = args.Session.ChannelMessageSendEmbed(args.Channel.ID, rep.AsEmbed(args.CmdHandler.config.WebServer.PublicAddr))
	}

	return err
}

func (c *CmdMute) list(args *CommandArgs) error {
	muteRoleID, err := args.CmdHandler.db.GetMuteRoleGuild(args.Guild.ID)
	if err != nil {
		return err
	}

	emb := &discordgo.MessageEmbed{
		Color:       static.ColorEmbedGray,
		Description: "Fetching muted members...",
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	msg, err := args.Session.ChannelMessageSendEmbed(args.Channel.ID, emb)
	if err != nil {
		return err
	}

	muteReports, err := args.CmdHandler.db.GetReportsFiltered(args.Guild.ID, "",
		util.IndexOfStrArray("MUTE", static.ReportTypes))

	muteReportsMap := make(map[string]*report.Report)
	for _, r := range muteReports {
		muteReportsMap[r.VictimID] = r
	}

	for _, m := range args.Guild.Members {
		if util.IndexOfStrArray(muteRoleID, m.Roles) > -1 {
			if r, ok := muteReportsMap[m.User.ID]; ok {
				emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
					Name: fmt.Sprintf("CaseID: %d", r.ID),
					Value: fmt.Sprintf("<@%s> since `%s` with reason:\n%s",
						m.User.ID, r.GetTimestamp().Format(time.RFC1123), r.Msg),
				})
			}
		}
	}

	emb.Color = static.ColorEmbedDefault
	emb.Description = ""

	_, err = args.Session.ChannelMessageEditEmbed(args.Channel.ID, msg.ID, emb)
	return err
}

func (c *CmdMute) displayMuteRole(args *CommandArgs) error {
	roleID, err := args.CmdHandler.db.GetMuteRoleGuild(args.Guild.ID)
	if err != nil {
		return err
	}

	if roleID == "" {
		return util.SendEmbedError(args.Session, args.Channel.ID,
			"Mute role is currently unset.").
			DeleteAfter(8 * time.Second).Error()
	}

	return util.SendEmbed(args.Session, args.Channel.ID,
		fmt.Sprintf("Role <@&%s> is currently set as mute role.", roleID), "", 0).
		DeleteAfter(8 * time.Second).Error()
}
