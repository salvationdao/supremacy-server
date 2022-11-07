package discord

import (
	"fmt"
	"server"
	"server/db"
	"server/gamelog"

	"github.com/bwmarrin/discordgo"
	"github.com/ninja-software/terror/v2"
)

type DiscordSession struct {
	s                  *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
	appID              string
	guildID            string
	IsActive           bool
}

var Session *DiscordSession

func NewDiscordBot(token, appID string, isBotBinary bool) (*DiscordSession, error) {
	session := &DiscordSession{
		IsActive: true,
	}
	Session = session

	guildID := db.GetStrWithDefault(db.KeyDiscordGuildID, "927761469775441930")

	if server.IsDevelopmentEnv() {
		guildID = "1034448717006258186"
	} else if server.IsStagingEnv() {
		guildID = "685421530477232138"
	}
	session.guildID = guildID

	bot, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, terror.Error(err, "Failed to initialize discord bot")
	}

	session.s = bot
	session.appID = appID

	if isBotBinary {
		commands := []*discordgo.ApplicationCommand{{}}
		session.s.AddHandler(func(DiscordSession *discordgo.Session, r *discordgo.Ready) {
			gamelog.L.Info().Msg("Discord session ready")
		})

		err = session.s.Open()
		if err != nil {
			gamelog.L.Err(err).Msg("Discord session failed to open")
			return nil, err
		}

		for _, command := range commands {
			session.s.ApplicationCommandCreate(appID, Session.guildID, command)
			session.registeredCommands = append(session.registeredCommands, command)
		}
	}

	return Session, nil
}

func (s *DiscordSession) CloseSession() {
	if !s.IsActive || s.guildID == "" {
		return
	}

	registeredCommands, err := s.s.ApplicationCommands(s.appID, s.guildID)
	if err != nil {
		gamelog.L.Err(err).Msg("failed to get all apps command")
	}

	if len(registeredCommands) > 0 {
		for _, cmd := range registeredCommands {
			err := s.s.ApplicationCommandDelete(s.appID, Session.guildID, cmd.ID)
			if err != nil {
				gamelog.L.Err(err).Interface("command", cmd).Msg("failed to delete app command")
			}
		}

	}

	s.s.Close()
}

func (s *DiscordSession) SendDiscordMessage(message string) error {
	if !s.IsActive {
		return nil
	}

	channelID := db.GetStrWithDefault(db.KeyDiscordChannelID, "946873011368251412")
	if server.IsDevelopmentEnv() {
		channelID = "1034448717006258189"
	} else if server.IsStagingEnv() {
		channelID = "685850676534050860"
	}

	_, err := s.s.ChannelMessageSend(channelID, message)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get send discord message for announcing")
		return err
	}

	return nil
}
