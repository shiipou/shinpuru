package inits

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/zekroTJA/shinpuru/internal/commands"
	"github.com/zekroTJA/shinpuru/internal/core"
	"github.com/zekroTJA/shinpuru/internal/listeners"
	"github.com/zekroTJA/shinpuru/internal/util"
)

func InitDiscordBotSession(session *discordgo.Session, config *core.Config, database core.Database, cmdHandler *commands.CmdHandler) {
	snowflake.Epoch = util.DefEpoche
	err := util.SetupSnowflakeNodes()
	if err != nil {
		util.Log.Fatal("Failed setting up snowflake nodes: ", err)
	}

	session.Token = "Bot " + config.Discord.Token

	listenerInviteBlock := listeners.NewListenerInviteBlock(database, cmdHandler)

	session.AddHandler(listeners.NewListenerReady(config, database).Handler)
	session.AddHandler(listeners.NewListenerCmd(config, database, cmdHandler).Handler)
	session.AddHandler(listeners.NewListenerGuildJoin(config).Handler)
	session.AddHandler(listeners.NewListenerMemberAdd(database).Handler)
	session.AddHandler(listeners.NewListenerVote(database).Handler)
	session.AddHandler(listeners.NewListenerChannelCreate(database).Handler)
	session.AddHandler(listeners.NewListenerVoiceUpdate(database).Handler)
	session.AddHandler(listeners.NewListenerGhostPing(database, cmdHandler).Handler)
	session.AddHandler(listeners.NewListenerJdoodle(database).Handler)
	session.AddHandler(listenerInviteBlock.HandlerMessageSend)
	session.AddHandler(listenerInviteBlock.HandlerMessageEdit)

	err = session.Open()
	if err != nil {
		util.Log.Fatal("Failed connecting Discord bot session:", err)
	}

	util.Log.Info("Started event loop. Stop with CTRL-C...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	util.Log.Info("Shutting down...")
	session.Close()
}
