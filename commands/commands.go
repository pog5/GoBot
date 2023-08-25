package commands

import (
	"gobot/config"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

type Command struct {
	Name        string
	Description string
	Execute     func(*events.MessageCreate, []string)
	Aliases     []string
	Permissions discord.Permissions
}

type Message struct {
	Content string
	Reply   bool
	Embeds  []discord.Embed
	Files   []*discord.File
}

func CreateMessage(e *events.MessageCreate, message Message) (*discord.Message, error) {
	builder := discord.NewMessageCreateBuilder().SetContent(message.Content).SetEmbeds(message.Embeds...).SetFiles(message.Files...)
	if message.Reply {
		builder.SetMessageReferenceByID(e.MessageID).SetAllowedMentions(&discord.AllowedMentions{RepliedUser: false})
	}
	return e.Client().Rest().CreateMessage(e.ChannelID, builder.Build())
}

func HasRole(client bot.Client, guildId snowflake.ID, memberId snowflake.ID, id int64) bool {
	member, _ := client.Rest().GetMember(guildId, memberId)
	for _, role := range member.RoleIDs {
		if role == snowflake.ID(id) {
			return true
		}
	}
	return false
}

func ParseMention(mention string) snowflake.ID {
	if strings.HasPrefix(mention, "<@") && strings.HasSuffix(mention, ">") && !strings.HasPrefix(mention, "<@&") {
		mention = strings.TrimPrefix(strings.TrimSuffix(mention, ">"), "<@")
	}
	id, err := strconv.ParseInt(mention, 10, 64)
	if err != nil {
		return 0
	}
	return snowflake.ID(id)
}

func Handle(message *events.MessageCreate) {
	if message.Message.Author.Bot {
		return
	}
	if strings.Contains(message.Message.Content, "mmm") {
		message.Client().Rest().AddReaction(message.ChannelID, message.MessageID, "✅")
	}
	args := strings.Split(message.Message.Content, " ")
	if !strings.HasPrefix(message.Message.Content, config.Config.Prefix) {
		if strings.HasPrefix(message.Message.Content, config.Config.InfoPrefix) {
			cmd := args[0][len(config.Config.InfoPrefix):]
			command := commands[cmd]
			if command.Execute == nil {
				command = commands[aliases[cmd]]
				if command.Execute == nil {
					return
				}
			}
			aliases := "None"
			if len(command.Aliases) > 0 {
				aliases = strings.Join(command.Aliases, ", ")
			}
			True := true
			embed := discord.NewEmbedBuilder().SetTitle(command.Name).SetDescription(command.Description).AddFields(
				discord.EmbedField{
					Name:   "Aliases",
					Value:  aliases,
					Inline: &True,
				},
				discord.EmbedField{
					Name:   "Permissions Required",
					Value:  command.Permissions.String(),
					Inline: &True,
				},
			).Build()
			CreateMessage(message, Message{Embeds: []discord.Embed{embed}})
		} else {
			return
		}
	} else {
		cmd := args[0][len(config.Config.Prefix):]
		args = args[1:]
		command := commands[cmd]
		if command.Execute == nil {
			command = commands[aliases[cmd]]
			if command.Execute == nil {
				return
			}
		}
		message.Message.Member.GuildID = *message.GuildID
		if !message.Client().Caches().MemberPermissions(*message.Message.Member).Has(command.Permissions) {
			return
		}
		command.Execute(message, args)
	}
}

var commands = make(map[string]Command)
var aliases = make(map[string]string)

func RegisterCommands(cmds ...Command) {
	for _, command := range cmds {
		commands[command.Name] = command
		for _, alias := range command.Aliases {
			aliases[alias] = command.Name
		}
	}
}
