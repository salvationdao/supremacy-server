package discord

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	session.guildID = guildID

	bot, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, terror.Error(err, "Failed to initialize discord bot")
	}

	session.s = bot
	session.appID = appID

	if isBotBinary {
		commands := []*discordgo.ApplicationCommand{{}}

		for _, command := range commands {
			session.s.ApplicationCommandCreate(appID, Session.guildID, command)
			session.registeredCommands = append(session.registeredCommands, command)
		}
	}

	session.s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Message == nil || i.Member == nil || i.Member.User == nil {
			return
		}

		if i.Type != discordgo.InteractionMessageComponent {
			return
		}

		messageAnnouncement, err := boiler.DiscordLobbyAnnoucements(
			boiler.DiscordLobbyAnnoucementWhere.MessageID.EQ(i.Message.ID),
			qm.Load(boiler.DiscordLobbyAnnoucementRels.BattleLobby),
		).One(gamedb.StdConn)
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Lobby not found",
					Flags:   1 << 6,
				},
			})
			return
		}

		discordAnnouncementFollower, err := boiler.DiscordLobbyFollowers(
			boiler.DiscordLobbyFollowerWhere.DiscordMemberID.EQ(i.Member.User.ID),
			boiler.DiscordLobbyFollowerWhere.DiscordLobbyAnnoucementsID.EQ(messageAnnouncement.ID),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return
		}

		if discordAnnouncementFollower != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You are already following this lobby",
					Flags:   1 << 6,
				},
			})
			return
		}

		newFollower := &boiler.DiscordLobbyFollower{
			DiscordMemberID:            i.Member.User.ID,
			DiscordLobbyAnnoucementsID: discordAnnouncementFollower.ID,
		}

		err = newFollower.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return
		}

		content := "You are now following this lobby"

		if messageAnnouncement.R != nil && messageAnnouncement.R.BattleLobby != nil {
			content = fmt.Sprintf("You are now following lobby `%s`", messageAnnouncement.R.BattleLobby.Name)
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   1 << 6,
			},
		})

	})

	session.s.AddHandler(func(DiscordSession *discordgo.Session, r *discordgo.Ready) {
		gamelog.L.Info().Msg("Discord session ready")
	})

	session.s.AddHandler(func(DiscordSession *discordgo.Session, r *discordgo.Ready) {
		gamelog.L.Info().Msg("Discord session ready")
	})

	err = session.s.Open()
	if err != nil {
		gamelog.L.Err(err).Msg("Discord session failed to open")
		return nil, err
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

func (s *DiscordSession) SendDiscordMessage(channelID, message string) error {
	if !s.IsActive {
		return nil
	}

	_, err := s.s.ChannelMessageSend(channelID, message)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get send discord message for announcing")
		return err
	}

	return nil
}

func (s *DiscordSession) SendBattleLobbyCreateMessage(battleLobbyID string) error {
	if !s.IsActive {
		return nil
	}

	messageEmbed, messageComponent, err := db.GetDiscordEmbedMessage(battleLobbyID)
	if err != nil {
		return err
	}

	dataSend := &discordgo.MessageSend{
		Content: "",
		Embeds: []*discordgo.MessageEmbed{
			messageEmbed,
		},
		TTS:        false,
		Components: messageComponent,
	}

	battleArenaChannelID := db.GetStrWithDefault(db.KeyDiscordBattleArenaChannelID, "973800997128392785")

	message, err := s.s.ChannelMessageSendComplex(battleArenaChannelID, dataSend)
	if err != nil {
		return err
	}

	newMessage := boiler.DiscordLobbyAnnoucement{
		MessageID:     message.ID,
		BattleLobbyID: battleLobbyID,
	}

	err = newMessage.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (s *DiscordSession) SendBattleLobbyEditMessage(battleLobbyID, arenaName string) error {
	if !s.IsActive {
		return nil
	}

	messageEmbed, _, err := db.GetDiscordEmbedMessage(battleLobbyID)
	if err != nil {
		return err
	}

	annoucement, err := boiler.DiscordLobbyAnnoucements(
		boiler.DiscordLobbyAnnoucementWhere.BattleLobbyID.EQ(battleLobbyID),
		qm.Load(boiler.DiscordLobbyAnnoucementRels.BattleLobby),
	).One(gamedb.StdConn)
	if err != nil {
		return err
	}

	_, err = s.s.ChannelMessageEditEmbed("973800997128392785", annoucement.MessageID, messageEmbed)
	if err != nil {
		return err
	}

	if annoucement.R.BattleLobby.AssignedToBattleID.Valid && !annoucement.R.BattleLobby.EndedAt.Valid {
		go func(annoucement *boiler.DiscordLobbyAnnoucement) {
			allFollowers, err := boiler.DiscordLobbyFollowers(
				boiler.DiscordLobbyFollowerWhere.DiscordLobbyAnnoucementsID.EQ(annoucement.ID),
			).All(gamedb.StdConn)
			if err != nil {
				return
			}

			peopleTag := ""

			for _, follower := range allFollowers {
				peopleTag = fmt.Sprintf("%s<@%s>\n", peopleTag, follower.DiscordMemberID)
			}

			battleArenaBaseUrl := db.GetStrWithDefault(db.KeyBattleArenaWebURL, "https://play.supremacy.game")

			arenaURLName := strings.ReplaceAll(arenaName, " ", "+")
			battleURL := fmt.Sprintf("%s/?arenaName=%s", battleArenaBaseUrl, arenaURLName)

			lobbyName := "Unknown"

			if annoucement.R != nil && annoucement.R.BattleLobby != nil {
				lobbyName = annoucement.R.BattleLobby.Name
			}

			message := fmt.Sprintf("%s\n\nLobby `%s` has entered the arena. Join your syndicate and fight now at %s", peopleTag, lobbyName, battleURL)

			battleArenaChannelID := db.GetStrWithDefault(db.KeyDiscordBattleArenaChannelID, "973800997128392785")
			_, err = s.s.ChannelMessageSend(battleArenaChannelID, message)
			if err != nil {
				return
			}
		}(annoucement)
	}

	return nil
}
