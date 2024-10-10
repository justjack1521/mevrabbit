package mevrabbit

import (
	"errors"
	"fmt"
)

var (
	errUserIDNotFound                 = errors.New("user id not found")
	errExtractUserIDFromMessageHeader = func(err error) error {
		return fmt.Errorf("failed to extract user id from message header: %w", err)
	}
	errPlayerIDNotFound                 = errors.New("player id not found")
	errExtractPlayerIDFromMessageHeader = func(err error) error {
		return fmt.Errorf("failed to extract player id from message header: %w", err)
	}
)

type ExchangeKind string

const (
	Direct  ExchangeKind = "direct"
	Fanout  ExchangeKind = "fanout"
	Headers ExchangeKind = "headers"
	Topic   ExchangeKind = "topic"
)

type Exchange string

const (
	Client  Exchange = "client"
	User    Exchange = "user"
	Social  Exchange = "social"
	Ranking Exchange = "ranking"
	Game    Exchange = "game"
)

type RoutingKey string

const (
	ClientNotification RoutingKey = "client.notification"
	ClientHeartbeat    RoutingKey = "client.heartbeat"
	ClientConnected    RoutingKey = "client.connected"
	ClientDisconnected RoutingKey = "client.disconnected"
	PlayerCreation     RoutingKey = "player.creation"
	PlayerDeletion     RoutingKey = "player.deletion"
	PlayerComment      RoutingKey = "player.comment"
	PlayerCompanion    RoutingKey = "player.companion"
	PlayerPosition     RoutingKey = "player.position"
	PlayerPresence     RoutingKey = "player.presence"
	PlayerRental       RoutingKey = "player.rental"
	PlayerLevel        RoutingKey = "player.level"
	BattleComplete     RoutingKey = "battle.complete"
)

const (
	UserCreated RoutingKey = "user.created"
	UserDeleted RoutingKey = "user.deleted"
)

type Queue string

const (
	UserUpdate      Queue = "user.update"
	ClientUpdate    Queue = "client.update"
	CommentUpdate   Queue = "comment.update"
	CompanionUpdate Queue = "companion.update"
	PlayerCreated   Queue = "player.created"
	PlayerDeleted   Queue = "player.deleted"
	PositionUpdate  Queue = "position.update"
	PresenceUpdate  Queue = "presence.update"
	LevelUpdate     Queue = "level.update"
	RentalUpdate    Queue = "rental.update"
	SocialUpdate    Queue = "social.update"
	RankingUpdate   Queue = "ranking.update"
	LoadoutUpdate   Queue = "loadout.update"
)

const (
	userIDHeaderKey   = "user_id"
	playerIDHeaderKey = "player_id"
)
