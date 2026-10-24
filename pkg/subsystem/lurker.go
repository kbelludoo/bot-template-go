package subsystem

import (
	"fmt"
	"strings"
)

// BuildKickHandshakePayload monta o evento WebSocket de handshake de canal da Kick (wss://websockets.kick.com/viewer/v1/connect)
func BuildKickHandshakePayload(channelID string) map[string]any {
	return map[string]any{
		"type": "channel_handshake",
		"data": map[string]any{
			"message": map[string]any{
				"channelId": channelID,
			},
		},
	}
}

// BuildKickWatchLivestreamPayload monta o batimento de presença em live da Kick para acumular pontos/drops
func BuildKickWatchLivestreamPayload(channelID, livestreamID string) map[string]any {
	return map[string]any{
		"type": "user_event",
		"data": map[string]any{
			"message": map[string]any{
				"name":         "tracking.user.watch.livestream",
				"channel_id":   channelID,
				"livestream_id": livestreamID,
			},
		},
	}
}

// BuildTwitchClaimPointsGQL monta a mutação GraphQL oficial para resgatar baús de Channel Points na Twitch
func BuildTwitchClaimPointsGQL(channelID, claimID string) map[string]any {
	return map[string]any{
		"operationName": "ClaimCommunityPoints",
		"variables": map[string]any{
			"input": map[string]any{
				"channelID": channelID,
				"claimID":   claimID,
			},
		},
		"extensions": map[string]any{
			"persistedQuery": map[string]any{
				"version":    1,
				"sha256Hash": "46aaeebe02c99afdf4fc97c7c0cba964124bf6b0af229395f1f6d1feed05b3d0",
			},
		},
	}
}

// FormatTwitchIRCJoin gera os comandos IRC para entrar em múltiplos canais na Twitch (irc.chat.twitch.tv:6697)
func FormatTwitchIRCJoin(channels []string) []string {
	cmds := make([]string, 0, len(channels))
	for _, ch := range channels {
		clean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ch)), "#")
		if clean != "" {
			cmds = append(cmds, fmt.Sprintf("JOIN #%s", clean))
		}
	}
	return cmds
}
