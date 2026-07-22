package bot

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BotChannel provides resource knowledge for
// Microsoft.BotService/botServices/channels.
//
// This is a MERGED knowledge file: AzureRM splits this single ARM resource type
// into ten typed Terraform resources, each of which calls
// commonids.NewBotServiceChannelID(...) with a channel.BotServiceChannelType*
// discriminator as the ARM name segment and expands one channel.<X>Channel body
// variant into the polymorphic properties (channel.Channel, discriminated by
// properties.channelName). All ten register the SAME ARM type, so azwise emits a
// single file here.
//
// Contributing TF resources (ARM name discriminator in parentheses):
//   - azurerm_bot_channel_email             (EmailChannel)
//   - azurerm_bot_channel_facebook          (FacebookChannel)
//   - azurerm_bot_channel_line              (LineChannel)
//   - azurerm_bot_channel_ms_teams          (MsTeamsChannel)
//   - azurerm_bot_channel_slack             (SlackChannel)
//   - azurerm_bot_channel_sms               (SmsChannel)
//   - azurerm_bot_channel_web_chat          (WebChatChannel)
//   - azurerm_bot_channel_directline        (DirectLineChannel)
//   - azurerm_bot_channel_alexa             (AlexaChannel)
//   - azurerm_bot_channel_direct_line_speech (DirectLineSpeechChannel)
//
// Only knowledge TRUE FOR EVERY channel variant is encoded here:
//   - ResourceType / ApiVersions / Timeouts (all variants use 30m/5m/30m/30m).
//   - The BotChannel envelope read-only fields (etag/id/type + channel
//     provisioningState) returned by GET but absent from any Create body.
//
// Deliberately NOT unioned onto the shared type (would corrupt validation for
// the other nine variants, since each fires only when its own body variant is
// present):
//   - RequiredFields: each variant requires different body props, e.g. Email
//     requires properties.properties.emailAddress; Facebook requires appId +
//     page/verifyToken; Line requires channelId+channelSecret registrations;
//     SMS requires phone/account/isValidated; Slack requires clientId/secret/
//     verificationToken; DirectLineSpeech requires cognitiveServiceRegion+key.
//     These stay per-variant, not universal.
//   - ForceNew: the only unconditional ForceNew argument shared by every
//     resource is bot_name — a PARENT reference (part of the resource ID), not
//     an ARM body property, so it is not encoded as a body ForceNew rule.
//     Variant-specific body ForceNew (e.g. Slack/Line channel identifiers) stays
//     per-variant.
//   - DefaultValues: none are universal; each variant hardcodes isEnabled=true
//     inside its own channelProperties sub-object.
//
// Note: properties.channelName is the polymorphic discriminator and is supplied
// as the ARM resource NAME segment (BotServiceChannelType*), not as a free body
// input, so it is not encoded as a RequiredField.
//
// Sources:
//   - terraform-provider-azurerm internal/services/bot/bot_channel_email_resource.go
//     (schema + Create; timeouts 30m/5m/30m/30m; NewBotServiceChannelID with
//     BotServiceChannelTypeEmailChannel; Kind=bot; isEnabled hardcoded true)
//   - .../bot_channel_facebook_resource.go, bot_channel_line_resource.go,
//     bot_channel_ms_teams_resource.go, bot_channel_slack_resource.go,
//     bot_channel_sms_resource.go, bot_channel_web_chat_resource.go,
//     bot_channel_directline_resource.go, bot_channel_alexa_resource.go,
//     bot_channel_direct_line_speech_resource.go (all share the timeout block and
//     NewBotServiceChannelID discriminator pattern)
//   - go-azure-sdk resource-manager/botservice/2022-09-15/channel
//     model_botchannel.go (BotChannel envelope: etag/id/type read-only;
//     Properties is polymorphic channel.Channel), model_channel.go
//     (BaseChannelImpl.provisioningState read-only), constants.go
//     (BotServiceChannelType discriminator set)
type BotChannel struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BotChannel)(nil)

// NewBotChannel returns knowledge for the botServices/channels resource.
func NewBotChannel() *BotChannel {
	return &BotChannel{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.BotService/botServices/channels",
			ApiVersions:  []string{"2022-09-15"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// BotChannel envelope + base channel props returned by GET but absent
			// from every Create body variant.
			ComputedFields: []string{
				"etag",
				"id",
				"type",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewBotChannel()) }
