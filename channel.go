package kakao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ChannelService calls the KakaoTalk Channel APIs: channel relationships
// and customer file management (고객 관리).
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
type ChannelService struct {
	c *Client
}

// Relation is a user's relationship with a KakaoTalk Channel.
type Relation string

// Channel relations.
const (
	RelationAdded   Relation = "ADDED"
	RelationBlocked Relation = "BLOCKED"
	RelationNone    Relation = "NONE" // never added or blocked
)

// ChannelRelation is a user's relationship with one channel.
type ChannelRelation struct {
	ChannelUUID     string   `json:"channel_uuid"`
	ChannelPublicID string   `json:"channel_public_id"`
	Relation        Relation `json:"relation"`
	// CreatedAt is when the channel was added. Only set when ADDED.
	CreatedAt Time `json:"created_at"`
	// UpdatedAt is when the relation last changed. Set when ADDED or
	// BLOCKED.
	UpdatedAt Time `json:"updated_at"`
}

// UserChannels is a user's relationships with the app's channels.
type UserChannels struct {
	UserID   int64             `json:"user_id"`
	Channels []ChannelRelation `json:"channels"`
}

// Relations returns the user's relationships with the channels linked to
// the app, or only with channelPublicIDs when given. It needs the
// plusfriends consent item.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
func (s *ChannelService) Relations(ctx context.Context, sub Subject, channelPublicIDs ...string) (*UserChannels, error) {
	q := channelQuery(channelPublicIDs)
	var res UserChannels
	if err := s.c.doAs(ctx, sub, &request{method: http.MethodGet, host: HostAPI, path: "/v2/api/talk/channels", query: q}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// MultiRelations returns the channel relationships of up to 200 users with
// the admin key. Users whose relationship could not be checked are left
// out.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
func (s *ChannelService) MultiRelations(ctx context.Context, userIDs []int64, channelPublicIDs ...string) ([]UserChannels, error) {
	if len(userIDs) == 0 || len(userIDs) > 200 {
		return nil, fmt.Errorf("kakao: need 1 to 200 user IDs, got %d", len(userIDs))
	}
	ids := make([]string, len(userIDs))
	for i, id := range userIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}
	q := channelQuery(channelPublicIDs)
	q.Set("target_ids", strings.Join(ids, ","))
	q.Set("target_id_type", "user_id")
	var res []UserChannels
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v2/api/talk/channels/multi", auth: authAdminKey, query: q}, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func channelQuery(channelPublicIDs []string) url.Values {
	q := url.Values{}
	if len(channelPublicIDs) > 0 {
		q.Set("channel_ids", strings.Join(channelPublicIDs, ","))
		q.Set("channel_id_type", "channel_public_id")
	}
	return q
}

// The customer file APIs accept the REST API key or the admin key. They
// use the REST API key when configured.
func (s *ChannelService) appKey() authKind {
	if s.c.restKey != "" {
		return authRESTKey
	}
	return authAdminKey
}

// SchemaType is the type of a customer file field.
type SchemaType string

// Schema types.
const (
	SchemaString SchemaType = "string"
	SchemaNumber SchemaType = "number"
)

// CreateUserFile creates a customer file (고객파일) for a channel and
// returns its ID. The schema maps field names to types; string fields are
// limited to the keys Kakao allows (생년월일, 국가, 지역, 성별, 연령, 구매금액,
// 포인트, 가입일, 최근 구매일, 응모일). A schema cannot be changed later and
// has at most 30 fields.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
func (s *ChannelService) CreateUserFile(ctx context.Context, channelPublicID, fileName string, schema map[string]SchemaType) (int64, error) {
	if channelPublicID == "" || fileName == "" || len(schema) == 0 {
		return 0, errors.New("kakao: channel public ID, file name and schema are required")
	}
	body := struct {
		ChannelPublicID string                `json:"channel_public_id"`
		FileName        string                `json:"file_name"`
		Schema          map[string]SchemaType `json:"schema"`
	}{channelPublicID, fileName, schema}
	var res struct {
		FileID int64 `json:"file_id"`
	}
	err := s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v1/talkchannel/create/target_user_file", auth: s.appKey(), json: body}, &res)
	return res.FileID, err
}

// UserFiles is the list of a channel's customer files.
type UserFiles struct {
	EmptySlot int        `json:"empty_slot"`
	UsingSlot int        `json:"using_slot"`
	Results   []UserFile `json:"results"`
}

// UserFile is a customer file.
type UserFile struct {
	FileID   int64  `json:"file_id"`
	FileName string `json:"file_name"`
	// Status is "using", "deleting" or "failed".
	Status string `json:"status"`
	// UpdateAt is when the file was uploaded, as Kakao formats it.
	UpdateAt string                `json:"update_at"`
	Schema   map[string]SchemaType `json:"schema"`
}

// UserFiles lists a channel's customer files.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
func (s *ChannelService) UserFiles(ctx context.Context, channelPublicID string) (*UserFiles, error) {
	if channelPublicID == "" {
		return nil, errors.New("kakao: empty channel public ID")
	}
	q := url.Values{"channel_public_id": {channelPublicID}}
	var res UserFiles
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v1/talkchannel/target_user_file", auth: s.appKey(), query: q}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// UserIDType says how customer file users are identified.
type UserIDType string

// User ID types.
const (
	UserIDApp   UserIDType = "app"   // Kakao Login user ID
	UserIDPhone UserIDType = "phone" // KakaoTalk phone number
)

// FileUser is a user to add to a customer file.
type FileUser struct {
	// ID is a user ID or phone number, depending on the UserIDType.
	ID string `json:"id"`
	// Field holds the values for the file's schema. Values must be strings
	// or numbers.
	Field map[string]any `json:"field"`
}

// AddUsersResult reports how many users were added to a customer file.
// Only users who are friends of the channel can be added.
type AddUsersResult struct {
	FileID       int64 `json:"file_id"`
	RequestCount int   `json:"request_count"`
	SuccessCount int   `json:"success_count"`
}

// AddUsers adds up to 2,000 users to a customer file.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
func (s *ChannelService) AddUsers(ctx context.Context, channelPublicID string, fileID int64, idType UserIDType, users []FileUser) (*AddUsersResult, error) {
	if len(users) == 0 || len(users) > 2000 {
		return nil, fmt.Errorf("kakao: need 1 to 2000 users, got %d", len(users))
	}
	body := struct {
		FileID          int64      `json:"file_id"`
		ChannelPublicID string     `json:"channel_public_id"`
		UserType        UserIDType `json:"user_type"`
		Users           []FileUser `json:"users"`
	}{fileID, channelPublicID, idType, users}
	var res AddUsersResult
	err := s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v1/talkchannel/update/target_users", auth: s.appKey(), json: body}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// RemoveUsers removes users from a customer file.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/rest-api
func (s *ChannelService) RemoveUsers(ctx context.Context, channelPublicID string, fileID int64, idType UserIDType, userIDs []string) error {
	if len(userIDs) == 0 {
		return errors.New("kakao: no user IDs")
	}
	body := struct {
		FileID          int64      `json:"file_id"`
		ChannelPublicID string     `json:"channel_public_id"`
		UserType        UserIDType `json:"user_type"`
		UserIDs         []string   `json:"user_ids"`
	}{fileID, channelPublicID, idType, userIDs}
	return s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v1/talkchannel/delete/target_users", auth: s.appKey(), json: body}, nil)
}
