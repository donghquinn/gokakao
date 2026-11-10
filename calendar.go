package kakao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CalendarService calls the Talk Calendar (톡캘린더) API: user calendars and
// events, channel public events, subscription calendars, holidays and
// tasks. User calls need the talk_calendar consent item and permission.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
type CalendarService struct {
	c *Client
}

// PrimaryCalendar is the ID of the user's default calendar.
const PrimaryCalendar = "primary"

// CalendarColor is a calendar or event color.
type CalendarColor string

// Calendar colors.
const (
	ColorBlue      CalendarColor = "BLUE"
	ColorRoyalBlue CalendarColor = "ROYAL_BLUE"
	ColorNavyBlue  CalendarColor = "NAVY_BLUE"
	ColorRed       CalendarColor = "RED"
	ColorPink      CalendarColor = "PINK"
	ColorOrange    CalendarColor = "ORANGE"
	ColorGreen     CalendarColor = "GREEN"
	ColorLime      CalendarColor = "LIME"
	ColorOlive     CalendarColor = "OLIVE"
	ColorMint      CalendarColor = "MINT"
	ColorMagenta   CalendarColor = "MAGENTA"
	ColorViolet    CalendarColor = "VIOLET"
	ColorLavender  CalendarColor = "LAVENDER"
	ColorBrown     CalendarColor = "BROWN"
	ColorGray      CalendarColor = "GRAY"
)

// EventTime is when an event takes place.
type EventTime struct {
	// StartAt and EndAt are RFC 3339 UTC times on 5-minute boundaries.
	// For all-day events they must be midnight, e.g.
	// "2022-10-27T00:00:00Z". Use [FormatEventTime] to format them.
	StartAt string `json:"start_at,omitempty"`
	EndAt   string `json:"end_at,omitempty"`
	// TimeZone is an IANA name; Kakao's default is Asia/Seoul.
	TimeZone string `json:"time_zone,omitempty"`
	AllDay   bool   `json:"all_day,omitempty"`
	Lunar    bool   `json:"lunar,omitempty"`
}

// FormatEventTime formats t as an RFC 3339 UTC time for [EventTime].
func FormatEventTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// EventLocation is where an event takes place.
type EventLocation struct {
	Name       string  `json:"name,omitempty"` // up to 100 chars
	LocationID int64   `json:"location_id,omitempty"`
	Address    string  `json:"address,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// CalendarLink is a link in calendar banners and reminder buttons.
type CalendarLink struct {
	WebURL       string `json:"web_url,omitempty"`
	MobileWebURL string `json:"mobile_web_url,omitempty"`
}

// --- User calendars ---

// CalendarFilter selects which calendars [CalendarService.Calendars]
// lists.
type CalendarFilter string

// Calendar filters.
const (
	CalendarFilterAll       CalendarFilter = "ALL"       // default
	CalendarFilterUser      CalendarFilter = "USER"      // default and sub-calendars
	CalendarFilterSubscribe CalendarFilter = "SUBSCRIBE" // subscribed calendars
)

// UserCalendars are the user's calendars.
type UserCalendars struct {
	Calendars          []UserCalendar `json:"calendars"`
	SubscribeCalendars []UserCalendar `json:"subscribe_calendars"`
}

// UserCalendar is one of the user's calendars. Name and Color are only
// returned for calendars the app created.
type UserCalendar struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Color          CalendarColor `json:"color"`
	Reminder       *int          `json:"reminder"`
	ReminderAllDay *int          `json:"reminder_all_day"`
	// Description, ProfileImageURL and ThumbnailURL are set for
	// subscribed calendars.
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	ThumbnailURL    string `json:"thumbnail_url"`
}

// Calendars lists the user's calendars. An empty filter lists all.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Calendars(ctx context.Context, accessToken string, filter CalendarFilter) (*UserCalendars, error) {
	q := url.Values{}
	setString(q, "filter", string(filter))
	var res UserCalendars
	if err := s.user(ctx, accessToken, http.MethodGet, "/v2/api/calendar/calendars", q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CalendarSettings are the settings of a sub-calendar. When updating, nil
// and empty fields keep their current values.
type CalendarSettings struct {
	Name  string        // up to 50 chars; required to create
	Color CalendarColor // default BLUE
	// Reminder and ReminderAllDay are the default reminders in minutes,
	// in 5-minute steps, for timed and all-day events.
	Reminder       *int
	ReminderAllDay *int
}

func (cs *CalendarSettings) form() url.Values {
	f := url.Values{}
	setString(f, "name", cs.Name)
	setString(f, "color", string(cs.Color))
	if cs.Reminder != nil {
		f.Set("reminder", strconv.Itoa(*cs.Reminder))
	}
	if cs.ReminderAllDay != nil {
		f.Set("reminder_all_day", strconv.Itoa(*cs.ReminderAllDay))
	}
	return f
}

// CreateCalendar creates a sub-calendar and returns its ID.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) CreateCalendar(ctx context.Context, accessToken string, settings CalendarSettings) (string, error) {
	if settings.Name == "" {
		return "", errors.New("kakao: calendar name is required")
	}
	var res struct {
		CalendarID string `json:"calendar_id"`
	}
	err := s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/create/calendar", settings.form(), &res)
	return res.CalendarID, err
}

// UpdateCalendar changes a sub-calendar the app created.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) UpdateCalendar(ctx context.Context, accessToken, calendarID string, settings CalendarSettings) error {
	if calendarID == "" {
		return errors.New("kakao: empty calendar ID")
	}
	form := settings.form()
	if len(form) == 0 {
		return errors.New("kakao: no calendar settings to update")
	}
	form.Set("calendar_id", calendarID)
	return s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/update/calendar", form, nil)
}

// DeleteCalendar deletes a sub-calendar the app created.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) DeleteCalendar(ctx context.Context, accessToken, calendarID string) error {
	if calendarID == "" {
		return errors.New("kakao: empty calendar ID")
	}
	return s.user(ctx, accessToken, http.MethodDelete, "/v2/api/calendar/delete/calendar", url.Values{"calendar_id": {calendarID}}, nil)
}

// --- Events ---

// NewEvent describes an event to create.
type NewEvent struct {
	Title string    `json:"title"` // up to 50 chars
	Time  EventTime `json:"time"`
	// RRule makes the event recurring, e.g. "FREQ=DAILY;UNTIL=20221031T000000Z".
	RRule       string         `json:"rrule,omitempty"`
	Description string         `json:"description,omitempty"` // up to 5,000 chars
	Location    *EventLocation `json:"location,omitempty"`
	// Reminders are minutes before the start, at most 2, in 5-minute
	// steps. Empty uses the calendar's defaults.
	Reminders []int         `json:"reminders,omitempty"`
	Color     CalendarColor `json:"color,omitempty"`
}

// CreateEvent creates an event in calendarID (the primary calendar when
// empty) and returns its ID.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) CreateEvent(ctx context.Context, accessToken, calendarID string, event NewEvent) (string, error) {
	form, err := jsonForm("event", event, nil)
	if err != nil {
		return "", err
	}
	setString(form, "calendar_id", calendarID)
	var res struct {
		EventID string `json:"event_id"`
	}
	err = s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/create/event", form, &res)
	return res.EventID, err
}

// EventType is the kind of an event.
type EventType string

// Event types.
const (
	EventTypeUser      EventType = "USER"      // the user's own event
	EventTypeTeam      EventType = "TEAM"      // shared event
	EventTypePublic    EventType = "PUBLIC"    // channel public event
	EventTypeSubscribe EventType = "SUBSCRIBE" // subscription calendar event
)

// Event is a calendar event. Lists return a subset of the fields; events
// the app did not create only carry Time.
type Event struct {
	// ID identifies the event. Occurrences of a recurring event have IDs
	// of the form "<id>_<start>", e.g. "6358e3987ec8e318d0b813bc_20220517T120000Z".
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Type         EventType      `json:"type"`
	CalendarID   string         `json:"calendar_id"`
	Time         EventTime      `json:"time"`
	IsHost       bool           `json:"is_host"`
	IsRecurEvent bool           `json:"is_recur_event"`
	RRule        string         `json:"rrule"`
	DTStart      string         `json:"dt_start"`
	Description  string         `json:"description"`
	Location     *EventLocation `json:"location"`
	Reminders    []int          `json:"reminders"`
	Color        CalendarColor  `json:"color"`
	Memo         string         `json:"memo"`
	Banner       *EventBanner   `json:"banner"`
	// NotificationMessage is set on public events.
	NotificationMessage *NotificationMessage `json:"notification_message"`
}

// EventBanner is the banner of a public or subscription event.
type EventBanner struct {
	PCImageURL     string        `json:"pc_image_url"`
	MobileImageURL string        `json:"mobile_image_url"`
	BgColor        CalendarColor `json:"bg_color"`
	Link           *CalendarLink `json:"link"`
}

// EventPreset is a predefined range for [CalendarService.Events].
type EventPreset string

// Event presets.
const (
	PresetToday     EventPreset = "TODAY"
	PresetThisWeek  EventPreset = "THIS_WEEK"  // starting Sunday
	PresetThisMonth EventPreset = "THIS_MONTH" // starting on the 1st
)

// EventsQuery selects events for [CalendarService.Events]. Give either a
// Preset or a From/To range of at most 31 days.
type EventsQuery struct {
	// CalendarID limits the query to one calendar; empty queries all.
	CalendarID string
	Preset     EventPreset
	// TimeZone is the IANA zone presets are computed in (default
	// Asia/Seoul).
	TimeZone string
	From, To time.Time
	// Limit is the page size, up to 1000 (default 100).
	Limit int
	// NextPageToken continues a previous query; other fields are then
	// ignored. [CalendarService.AllEvents] handles it.
	NextPageToken string
}

// EventsPage is a page of events.
type EventsPage struct {
	Events  []Event `json:"events"`
	HasNext bool    `json:"has_next"`
	// AfterURL fetches the next page; see [EventsPage.NextPageToken].
	AfterURL string `json:"after_url"`
}

// NextPageToken returns the token for the next page, or "" if there is
// none.
func (p *EventsPage) NextPageToken() string {
	return queryParam(p.AfterURL, "next_page_token")
}

// Events lists the user's events.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Events(ctx context.Context, accessToken string, query EventsQuery) (*EventsPage, error) {
	q := url.Values{}
	setString(q, "calendar_id", query.CalendarID)
	if query.NextPageToken != "" {
		q.Set("next_page_token", query.NextPageToken)
	} else {
		setString(q, "preset", string(query.Preset))
		setString(q, "time_zone", query.TimeZone)
		setTime(q, "from", query.From)
		setTime(q, "to", query.To)
		setInt(q, "limit", query.Limit)
		if query.Preset == "" && (query.From.IsZero() || query.To.IsZero()) {
			return nil, errors.New("kakao: events query needs a preset or a from/to range")
		}
	}
	var res EventsPage
	if err := s.user(ctx, accessToken, http.MethodGet, "/v2/api/calendar/events", q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// AllEvents iterates over all events matching query, fetching pages as
// needed. Iteration stops at the first error, which is yielded.
func (s *CalendarService) AllEvents(ctx context.Context, accessToken string, query EventsQuery) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {
		for {
			page, err := s.Events(ctx, accessToken, query)
			if err != nil {
				yield(Event{}, err)
				return
			}
			for _, e := range page.Events {
				if !yield(e, nil) {
					return
				}
			}
			next := page.NextPageToken()
			if !page.HasNext || next == "" {
				return
			}
			query.NextPageToken = next
		}
	}
}

// Event returns an event of the user. For a recurring event, pass the ID
// of one occurrence.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Event(ctx context.Context, accessToken, eventID string) (*Event, error) {
	if eventID == "" {
		return nil, errors.New("kakao: empty event ID")
	}
	var res struct {
		Event Event `json:"event"`
	}
	if err := s.user(ctx, accessToken, http.MethodGet, "/v2/api/calendar/event", url.Values{"event_id": {eventID}}, &res); err != nil {
		return nil, err
	}
	return &res.Event, nil
}

// RecurUpdateType chooses which occurrences of a recurring event change.
// It is required when updating or deleting a recurring event.
type RecurUpdateType string

// Recurring event update scopes.
const (
	RecurAll              RecurUpdateType = "ALL"
	RecurThis             RecurUpdateType = "THIS"
	RecurThisAndFollowing RecurUpdateType = "THIS_AND_FOLLOWING"
)

// EventChanges are the changes to an event. Nil and empty fields keep
// their current values. To remove a value, list its JSON name in Clear,
// e.g. "description", "location" or "rrule".
type EventChanges struct {
	Title       *string        `json:"title,omitempty"`
	Time        *EventTime     `json:"time,omitempty"`
	RRule       *string        `json:"rrule,omitempty"`
	Description *string        `json:"description,omitempty"`
	Location    *EventLocation `json:"location,omitempty"`
	Reminders   []int          `json:"reminders,omitempty"`
	Color       CalendarColor  `json:"color,omitempty"`
	// Clear lists fields to remove.
	Clear []string `json:"-"`
}

// EventUpdate is an update to one of the user's events.
type EventUpdate struct {
	// CalendarID moves the event to another calendar the app may write to.
	CalendarID string
	// RecurUpdateType is required for recurring events.
	RecurUpdateType RecurUpdateType
	Changes         *EventChanges
}

// UpdateEvent changes one of the user's own events. It returns the event
// ID, which is empty for recurring events.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) UpdateEvent(ctx context.Context, accessToken, eventID string, update EventUpdate) (string, error) {
	if eventID == "" {
		return "", errors.New("kakao: empty event ID")
	}
	if update.CalendarID == "" && update.Changes == nil {
		return "", errors.New("kakao: nothing to update")
	}
	form := url.Values{}
	if update.Changes != nil {
		var err error
		if form, err = jsonForm("event", update.Changes, update.Changes.Clear); err != nil {
			return "", err
		}
	}
	form.Set("event_id", eventID)
	setString(form, "calendar_id", update.CalendarID)
	setString(form, "recur_update_type", string(update.RecurUpdateType))
	var res struct {
		EventID string `json:"event_id"`
	}
	err := s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/update/event/host", form, &res)
	return res.EventID, err
}

// DeleteEvent deletes one of the user's events. recur is required for
// recurring events and empty otherwise.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) DeleteEvent(ctx context.Context, accessToken, eventID string, recur RecurUpdateType) error {
	if eventID == "" {
		return errors.New("kakao: empty event ID")
	}
	q := url.Values{"event_id": {eventID}}
	setString(q, "recur_update_type", string(recur))
	return s.user(ctx, accessToken, http.MethodDelete, "/v2/api/calendar/delete/event", q, nil)
}

// GuestEventChanges are the changes a guest may make to an event they were
// invited to. Empty fields keep their current values.
type GuestEventChanges struct {
	Reminders []int         `json:"reminders,omitempty"`
	Color     CalendarColor `json:"color,omitempty"`
	Memo      *string       `json:"memo,omitempty"` // up to 5,000 chars
	// Clear lists fields to remove.
	Clear []string `json:"-"`
}

// UpdateGuestEvent changes the user's copy of an event they were invited
// to. calendarID moves it to another calendar; empty keeps it in place.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) UpdateGuestEvent(ctx context.Context, accessToken, eventID, calendarID string, changes *GuestEventChanges) error {
	if eventID == "" {
		return errors.New("kakao: empty event ID")
	}
	if calendarID == "" && changes == nil {
		return errors.New("kakao: nothing to update")
	}
	form := url.Values{}
	if changes != nil {
		var err error
		if form, err = jsonForm("event", changes, changes.Clear); err != nil {
			return err
		}
	}
	form.Set("event_id", eventID)
	setString(form, "calendar_id", calendarID)
	return s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/update/event/guest", form, nil)
}

// --- Public events ---

// NotificationMessage customizes the reminders of a public event.
type NotificationMessage struct {
	// DisplayCustomButton shows the custom button (default true).
	DisplayCustomButton *bool          `json:"display_custom_button,omitempty"`
	BeforeEventReminder *EventReminder `json:"before_event_reminder,omitempty"`
	AfterEventReminder  *EventReminder `json:"after_event_reminder,omitempty"`
}

// EventReminder customizes a reminder message.
type EventReminder struct {
	// Button replaces the default "공유하기" button.
	Button *ReminderButton `json:"button,omitempty"`
}

// ReminderButton is a reminder message button.
type ReminderButton struct {
	// Title is one of "시청하기", "예매하기", "예약하기", "참여하기" or "상세보기".
	Title string       `json:"title"`
	Link  CalendarLink `json:"link"`
}

// NewPublicEvent describes a channel public event to create.
type NewPublicEvent struct {
	Title               string               `json:"title"`
	Time                EventTime            `json:"time"`
	Description         string               `json:"description,omitempty"`
	Location            *EventLocation       `json:"location,omitempty"`
	Reminders           []int                `json:"reminders,omitempty"`
	Color               CalendarColor        `json:"color,omitempty"`
	NotificationMessage *NotificationMessage `json:"notification_message,omitempty"`
}

// CreatePublicEvent creates a public event for a channel linked to the app
// and returns its ID. channelPublicID may be empty when the app has only
// one channel. It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) CreatePublicEvent(ctx context.Context, channelPublicID string, event NewPublicEvent) (string, error) {
	form, err := jsonForm("event", event, nil)
	if err != nil {
		return "", err
	}
	setString(form, "channel_public_id", channelPublicID)
	var res struct {
		EventID string `json:"event_id"`
	}
	err = s.admin(ctx, http.MethodPost, "/v2/api/calendar/public/create/event", form, &res)
	return res.EventID, err
}

// PublicEventsPage is a page of public events.
type PublicEventsPage struct {
	Events   []Event `json:"events"`
	HasNext  bool    `json:"has_next"`
	AfterURL string  `json:"after_url"`
}

// PublicEvents lists a channel's public events between from and to (at
// most 31 days apart). limit is up to 30 (default 10). It uses the admin
// key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) PublicEvents(ctx context.Context, channelPublicID string, from, to time.Time, offset, limit int) (*PublicEventsPage, error) {
	if from.IsZero() || to.IsZero() {
		return nil, errors.New("kakao: from and to are required")
	}
	q := url.Values{}
	setString(q, "channel_public_id", channelPublicID)
	setTime(q, "from", from)
	setTime(q, "to", to)
	setInt(q, "offset", offset)
	setInt(q, "limit", limit)
	var res PublicEventsPage
	if err := s.admin(ctx, http.MethodGet, "/v2/api/calendar/public/events", q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// PublicEvent returns a channel public event. It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) PublicEvent(ctx context.Context, channelPublicID, eventID string) (*Event, error) {
	if eventID == "" {
		return nil, errors.New("kakao: empty event ID")
	}
	q := url.Values{"event_id": {eventID}}
	setString(q, "channel_public_id", channelPublicID)
	var res struct {
		Event Event `json:"event"`
	}
	if err := s.admin(ctx, http.MethodGet, "/v2/api/calendar/public/event", q, &res); err != nil {
		return nil, err
	}
	return &res.Event, nil
}

// PublicEventChanges are the changes to a public event. Nil and empty
// fields keep their current values; list JSON names in Clear to remove
// values.
type PublicEventChanges struct {
	Title               *string              `json:"title,omitempty"`
	Time                *EventTime           `json:"time,omitempty"`
	Description         *string              `json:"description,omitempty"`
	Location            *EventLocation       `json:"location,omitempty"`
	Reminders           []int                `json:"reminders,omitempty"`
	Color               CalendarColor        `json:"color,omitempty"`
	NotificationMessage *NotificationMessage `json:"notification_message,omitempty"`
	// Clear lists fields to remove.
	Clear []string `json:"-"`
}

// UpdatePublicEvent changes a channel public event. It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) UpdatePublicEvent(ctx context.Context, channelPublicID, eventID string, changes PublicEventChanges) error {
	if eventID == "" {
		return errors.New("kakao: empty event ID")
	}
	form, err := jsonForm("event", changes, changes.Clear)
	if err != nil {
		return err
	}
	form.Set("event_id", eventID)
	setString(form, "channel_public_id", channelPublicID)
	return s.admin(ctx, http.MethodPost, "/v2/api/calendar/public/update/event", form, nil)
}

// DeletePublicEvent deletes a channel public event. It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) DeletePublicEvent(ctx context.Context, channelPublicID, eventID string) error {
	if eventID == "" {
		return errors.New("kakao: empty event ID")
	}
	q := url.Values{"event_id": {eventID}}
	setString(q, "channel_public_id", channelPublicID)
	return s.admin(ctx, http.MethodDelete, "/v2/api/calendar/public/delete/event", q, nil)
}

// FollowPublicEvent adds a public event to one of the user's calendars
// (the primary calendar when calendarID is empty).
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) FollowPublicEvent(ctx context.Context, accessToken, eventID, calendarID string) error {
	if eventID == "" {
		return errors.New("kakao: empty event ID")
	}
	form := url.Values{"event_id": {eventID}}
	setString(form, "calendar_id", calendarID)
	return s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/public/follow", form, nil)
}

// --- Subscription calendars ---

// SubscribableCategory is a category of subscribable calendars.
type SubscribableCategory struct {
	CategoryName  string                    `json:"category_name"`
	Subcategories []SubscribableSubcategory `json:"subcategories"`
}

// SubscribableSubcategory is a subcategory of subscribable calendars.
type SubscribableSubcategory struct {
	SubcategoryName string                 `json:"subcategory_name"`
	Calendars       []SubscribableCalendar `json:"calendars"`
}

// SubscribableCalendar is a calendar users can subscribe to.
type SubscribableCalendar struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ProfileImageURL string `json:"profile_image_url"`
}

// SubscribableCalendars lists the calendars users can subscribe to,
// optionally limited to a category and subcategory. It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) SubscribableCalendars(ctx context.Context, category, subcategory string) ([]SubscribableCategory, error) {
	if subcategory != "" && category == "" {
		return nil, errors.New("kakao: subcategory requires a category")
	}
	q := url.Values{}
	setString(q, "category_name", category)
	setString(q, "subcategory_name", subcategory)
	var res struct {
		Categories []SubscribableCategory `json:"categories"`
	}
	if err := s.admin(ctx, http.MethodGet, "/v2/api/calendar/subscribable/calendars", q, &res); err != nil {
		return nil, err
	}
	return res.Categories, nil
}

// Subscribe subscribes the user to a subscription calendar.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Subscribe(ctx context.Context, accessToken, calendarID string) error {
	if calendarID == "" {
		return errors.New("kakao: empty calendar ID")
	}
	return s.user(ctx, accessToken, http.MethodPost, "/v2/api/calendar/subscribe", url.Values{"calendar_id": {calendarID}}, nil)
}

// Unsubscribe unsubscribes the user from a subscription calendar.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Unsubscribe(ctx context.Context, accessToken, calendarID string) error {
	if calendarID == "" {
		return errors.New("kakao: empty calendar ID")
	}
	return s.user(ctx, accessToken, http.MethodDelete, "/v2/api/calendar/unsubscribe", url.Values{"calendar_id": {calendarID}}, nil)
}

// --- Holidays ---

// Holiday is a public holiday or notable day.
type Holiday struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Time  EventTime `json:"time"`
	// Holiday reports whether the day is a public holiday.
	Holiday bool `json:"holiday"`
}

// Holidays lists public holidays and notable days between from and to (at
// most 31 days apart). It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Holidays(ctx context.Context, from, to time.Time) ([]Holiday, error) {
	if from.IsZero() || to.IsZero() {
		return nil, errors.New("kakao: from and to are required")
	}
	q := url.Values{}
	setTime(q, "from", from)
	setTime(q, "to", to)
	var res struct {
		Events []Holiday `json:"events"`
	}
	if err := s.admin(ctx, http.MethodGet, "/v2/api/calendar/holidays", q, &res); err != nil {
		return nil, err
	}
	return res.Events, nil
}

// --- Tasks ---

// DueInfo is a task's due date.
type DueInfo struct {
	// DueDate is "yyyyMMdd", not in the past, up to 20501231.
	DueDate string `json:"due_date"`
	// TimeZone is an IANA name (default Asia/Seoul).
	TimeZone string `json:"time_zone,omitempty"`
	// AlarmTime is "HHmm" in 5-minute steps; empty means no alarm.
	AlarmTime string `json:"alarm_time,omitempty"`
	// Recur makes the task repeat from DueDate.
	Recur *TaskRecur `json:"recur,omitempty"`
}

// TaskRecur makes a task repeat.
type TaskRecur struct {
	// RRule is an RFC 5545 rule, e.g. "FREQ=DAILY;COUNT=3".
	RRule string `json:"rrule"`
	// RecordOn enables the challenge record (도전 기록).
	RecordOn bool `json:"record_on,omitempty"`
}

// NewTask describes a task to create.
type NewTask struct {
	Content string   `json:"content"` // up to 1,000 chars
	DueInfo *DueInfo `json:"due_info,omitempty"`
}

// CreateTask creates a task and returns its ID.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) CreateTask(ctx context.Context, accessToken string, task NewTask) (string, error) {
	if task.Content == "" {
		return "", errors.New("kakao: task content is required")
	}
	form, err := jsonForm("task", task, nil)
	if err != nil {
		return "", err
	}
	var res struct {
		TaskID string `json:"task_id"`
	}
	err = s.user(ctx, accessToken, http.MethodPost, "/v1/api/calendar/create/task", form, &res)
	return res.TaskID, err
}

// Task is a task.
type Task struct {
	TaskID  string `json:"task_id"`
	Content string `json:"content"`
	// Status is "SCHEDULED" or "DELAYED" (past due).
	Status     string   `json:"status"`
	Bookmark   bool     `json:"bookmark"`
	Authorized bool     `json:"authorized"` // may be changed or deleted
	DueInfo    *DueInfo `json:"due_info"`
	IsEnded    bool     `json:"is_ended"`
}

// TaskStatus filters tasks by completion.
type TaskStatus string

// Task statuses.
const (
	TaskTodo      TaskStatus = "TODO" // default
	TaskCompleted TaskStatus = "COMPLETED"
	TaskAll       TaskStatus = "ALL"
)

// TasksQuery selects tasks for [CalendarService.Tasks].
type TasksQuery struct {
	// From and To are "yyyyMMdd", at most 31 days apart. Required.
	From, To string
	Status   TaskStatus
	// Filters are "authorized" and/or "bookmark"; all must match.
	Filters  []string
	Offset   int
	Limit    int // up to 1000 (default 100)
	TimeZone string
}

// TasksPage is a page of tasks.
type TasksPage struct {
	Tasks    []Task `json:"tasks"`
	Count    int    `json:"count"`
	HasNext  bool   `json:"has_next"`
	AfterURL string `json:"after_url"`
}

// Tasks lists the user's tasks due in a date range.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Tasks(ctx context.Context, accessToken string, query TasksQuery) (*TasksPage, error) {
	if query.From == "" || query.To == "" {
		return nil, errors.New("kakao: from and to are required")
	}
	q := url.Values{"from": {query.From}, "to": {query.To}}
	setString(q, "task_status", string(query.Status))
	setString(q, "task_filter", strings.Join(query.Filters, ","))
	setInt(q, "offset", query.Offset)
	setInt(q, "limit", query.Limit)
	setString(q, "time_zone", query.TimeZone)
	var res TasksPage
	if err := s.user(ctx, accessToken, http.MethodGet, "/v1/api/calendar/tasks", q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Task returns one of the user's tasks.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) Task(ctx context.Context, accessToken, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, errors.New("kakao: empty task ID")
	}
	var res TasksPage
	if err := s.user(ctx, accessToken, http.MethodGet, "/v1/api/calendar/tasks", url.Values{"task_id": {taskID}}, &res); err != nil {
		return nil, err
	}
	if len(res.Tasks) == 0 {
		return nil, fmt.Errorf("kakao: task %q not found", taskID)
	}
	return &res.Tasks[0], nil
}

// TaskRecords is the challenge record of a recurring task.
type TaskRecords struct {
	Records []TaskRecord `json:"records"`
	StartAt string       `json:"start_at"` // "yyyyMM"
	EndAt   string       `json:"end_at"`   // "yyyyMM"
}

// TaskRecord is the result of one day of a recurring task.
type TaskRecord struct {
	Date     string `json:"date"` // "yyyyMMdd"
	Complete bool   `json:"complete"`
}

// TaskRecords returns the challenge record of a recurring task with
// RecordOn set, for the months from to to ("yyyyMM").
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) TaskRecords(ctx context.Context, accessToken, taskID, from, to string) (*TaskRecords, error) {
	if taskID == "" || from == "" || to == "" {
		return nil, errors.New("kakao: task ID, from and to are required")
	}
	var res TaskRecords
	q := url.Values{"task_id": {taskID}, "from": {from}, "to": {to}}
	if err := s.user(ctx, accessToken, http.MethodGet, "/v1/api/calendar/task/records", q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// TaskChanges are the changes to a task. Nil fields keep their current
// values. List JSON names in Clear to remove values, e.g. "due_info" or
// "due_info.recur".
type TaskChanges struct {
	Content  *string  `json:"content,omitempty"`
	Bookmark *bool    `json:"bookmark,omitempty"`
	DueInfo  *DueInfo `json:"due_info,omitempty"`
	// Clear lists fields to remove.
	Clear []string `json:"-"`
}

// UpdateTask changes one of the user's tasks.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) UpdateTask(ctx context.Context, accessToken, taskID string, changes TaskChanges) error {
	if taskID == "" {
		return errors.New("kakao: empty task ID")
	}
	form, err := jsonForm("task", changes, changes.Clear)
	if err != nil {
		return err
	}
	form.Set("task_id", taskID)
	return s.user(ctx, accessToken, http.MethodPost, "/v1/api/calendar/update/task", form, nil)
}

// CompleteTask marks one of the user's tasks complete or not complete.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) CompleteTask(ctx context.Context, accessToken, taskID string, complete bool) error {
	if taskID == "" {
		return errors.New("kakao: empty task ID")
	}
	form := url.Values{"task_id": {taskID}, "complete": {strconv.FormatBool(complete)}}
	return s.user(ctx, accessToken, http.MethodPost, "/v1/api/calendar/complete/task", form, nil)
}

// DeleteTask deletes one of the user's tasks.
//
// Docs: https://developers.kakao.com/docs/ko/talkcalendar/rest-api
func (s *CalendarService) DeleteTask(ctx context.Context, accessToken, taskID string) error {
	if taskID == "" {
		return errors.New("kakao: empty task ID")
	}
	return s.user(ctx, accessToken, http.MethodDelete, "/v1/api/calendar/delete/task", url.Values{"task_id": {taskID}}, nil)
}

// --- helpers ---

// user sends a request with the user's access token. params go in the
// query for GET and DELETE and in the form body otherwise.
func (s *CalendarService) user(ctx context.Context, accessToken, method, path string, params url.Values, out any) error {
	return s.c.doAs(ctx, AccessToken(accessToken), paramsRequest(method, path, params), out)
}

// admin sends a request with the admin key.
func (s *CalendarService) admin(ctx context.Context, method, path string, params url.Values, out any) error {
	r := paramsRequest(method, path, params)
	r.auth = authAdminKey
	return s.c.do(ctx, r, out)
}

func paramsRequest(method, path string, params url.Values) *request {
	r := &request{method: method, host: HostAPI, path: path}
	if method == http.MethodGet || method == http.MethodDelete {
		r.query = params
	} else {
		if params == nil {
			params = url.Values{}
		}
		r.form = params
	}
	return r
}

// jsonForm returns a form with key set to the JSON encoding of v, with the
// dotted paths in nullPaths set to null.
func jsonForm(key string, v any, nullPaths []string) (url.Values, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("kakao: encode %s: %w", key, err)
	}
	if len(nullPaths) > 0 {
		var m map[string]any
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("kakao: encode %s: %w", key, err)
		}
		for _, p := range nullPaths {
			setNull(m, strings.Split(p, "."))
		}
		if b, err = json.Marshal(m); err != nil {
			return nil, fmt.Errorf("kakao: encode %s: %w", key, err)
		}
	}
	return url.Values{key: {string(b)}}, nil
}

func setNull(m map[string]any, path []string) {
	for len(path) > 1 {
		next, ok := m[path[0]].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[path[0]] = next
		}
		m, path = next, path[1:]
	}
	m[path[0]] = nil
}

// setTime sets key to t as an RFC 3339 UTC time unless t is zero.
func setTime(q url.Values, key string, t time.Time) {
	if !t.IsZero() {
		q.Set(key, FormatEventTime(t))
	}
}

// queryParam returns the value of key in the query of rawURL.
func queryParam(rawURL, key string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}
