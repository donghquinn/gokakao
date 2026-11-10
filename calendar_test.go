package kakao

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

// TestCalendarService_requests checks how each call is sent.
func TestCalendarService_requests(t *testing.T) {
	t.Parallel()
	from := time.Date(2022, 10, 26, 9, 0, 0, 0, time.FixedZone("KST", 9*3600))
	to := from.Add(4 * 24 * time.Hour)

	tests := []struct {
		name      string
		pattern   string
		response  string
		call      func(ctx context.Context, c *Client) error
		wantAuth  string
		wantQuery url.Values
		wantForm  url.Values
	}{
		{
			"calendars", "GET /v2/api/calendar/calendars", `{}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.Calendars(ctx, "AT", CalendarFilterUser)
				return err
			},
			"Bearer AT",
			url.Values{"filter": {"USER"}},
			nil,
		},
		{
			"create calendar", "POST /v2/api/calendar/create/calendar", `{"calendar_id":"user_1"}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.CreateCalendar(ctx, "AT", CalendarSettings{Name: "서비스 캘린더", Color: ColorRed, Reminder: ptr(15), ReminderAllDay: ptr(30)})
				return err
			},
			"Bearer AT", nil,
			url.Values{"name": {"서비스 캘린더"}, "color": {"RED"}, "reminder": {"15"}, "reminder_all_day": {"30"}},
		},
		{
			"update calendar", "POST /v2/api/calendar/update/calendar", `{"calendar_id":"user_1"}`,
			func(ctx context.Context, c *Client) error {
				return c.Calendar.UpdateCalendar(ctx, "AT", "user_1", CalendarSettings{Color: ColorBlue})
			},
			"Bearer AT", nil,
			url.Values{"calendar_id": {"user_1"}, "color": {"BLUE"}},
		},
		{
			"delete calendar", "DELETE /v2/api/calendar/delete/calendar", `{"calendar_id":"user_1"}`,
			func(ctx context.Context, c *Client) error { return c.Calendar.DeleteCalendar(ctx, "AT", "user_1") },
			"Bearer AT",
			url.Values{"calendar_id": {"user_1"}},
			nil,
		},
		{
			"events by range", "GET /v2/api/calendar/events", `{"events":[],"has_next":false}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.Events(ctx, "AT", EventsQuery{CalendarID: "user_1", From: from, To: to, Limit: 2, TimeZone: "Asia/Seoul"})
				return err
			},
			"Bearer AT",
			url.Values{"calendar_id": {"user_1"}, "from": {"2022-10-26T00:00:00Z"}, "to": {"2022-10-30T00:00:00Z"}, "limit": {"2"}, "time_zone": {"Asia/Seoul"}},
			nil,
		},
		{
			"events by preset", "GET /v2/api/calendar/events", `{"events":[],"has_next":false}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.Events(ctx, "AT", EventsQuery{Preset: PresetThisWeek})
				return err
			},
			"Bearer AT",
			url.Values{"preset": {"THIS_WEEK"}},
			nil,
		},
		{
			"events next page", "GET /v2/api/calendar/events", `{"events":[],"has_next":false}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.Events(ctx, "AT", EventsQuery{CalendarID: "user_1", NextPageToken: "tok", Preset: PresetToday, Limit: 9})
				return err
			},
			"Bearer AT",
			url.Values{"calendar_id": {"user_1"}, "next_page_token": {"tok"}},
			nil,
		},
		{
			"event", "GET /v2/api/calendar/event", `{"event":{"id":"e1"}}`,
			func(ctx context.Context, c *Client) error { _, err := c.Calendar.Event(ctx, "AT", "e1"); return err },
			"Bearer AT",
			url.Values{"event_id": {"e1"}},
			nil,
		},
		{
			"delete event", "DELETE /v2/api/calendar/delete/event", `{"event_id":"e1"}`,
			func(ctx context.Context, c *Client) error {
				return c.Calendar.DeleteEvent(ctx, "AT", "e1", &EventDeleteOptions{RecurUpdateType: RecurThis})
			},
			"Bearer AT",
			url.Values{"event_id": {"e1"}, "recur_update_type": {"THIS"}},
			nil,
		},
		{
			"update guest event", "POST /v2/api/calendar/update/event/guest", `{}`,
			func(ctx context.Context, c *Client) error {
				return c.Calendar.UpdateGuestEvent(ctx, "AT", "e1", GuestEventChanges{Memo: ptr("메모"), Clear: []string{"reminders"}}, &GuestEventUpdateOptions{CalendarID: "primary"})
			},
			"Bearer AT", nil,
			url.Values{"event_id": {"e1"}, "calendar_id": {"primary"}, "event": {`{"memo":"메모","reminders":null}`}},
		},
		{
			"public events", "GET /v2/api/calendar/public/events", `{"events":[],"has_next":false}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.PublicEvents(ctx, PublicEventsQuery{ChannelPublicID: "_xnrxjem", From: from, To: to, Offset: 3, Limit: 10})
				return err
			},
			"KakaoAK admin",
			url.Values{"channel_public_id": {"_xnrxjem"}, "from": {"2022-10-26T00:00:00Z"}, "to": {"2022-10-30T00:00:00Z"}, "offset": {"3"}, "limit": {"10"}},
			nil,
		},
		{
			"public event", "GET /v2/api/calendar/public/event", `{"event":{"id":"p1","type":"PUBLIC"}}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.PublicEvent(ctx, "", "p1")
				return err
			},
			"KakaoAK admin",
			url.Values{"event_id": {"p1"}},
			nil,
		},
		{
			"update public event", "POST /v2/api/calendar/public/update/event", `{"event_id":"p1"}`,
			func(ctx context.Context, c *Client) error {
				return c.Calendar.UpdatePublicEvent(ctx, "_x", "p1", PublicEventChanges{Title: ptr("새 제목"), Clear: []string{"location"}})
			},
			"KakaoAK admin", nil,
			url.Values{"channel_public_id": {"_x"}, "event_id": {"p1"}, "event": {`{"location":null,"title":"새 제목"}`}},
		},
		{
			"delete public event", "DELETE /v2/api/calendar/public/delete/event", `{"event_id":"p1"}`,
			func(ctx context.Context, c *Client) error { return c.Calendar.DeletePublicEvent(ctx, "_x", "p1") },
			"KakaoAK admin",
			url.Values{"channel_public_id": {"_x"}, "event_id": {"p1"}},
			nil,
		},
		{
			"follow public event", "POST /v2/api/calendar/public/follow", `{"event_id":"p1"}`,
			func(ctx context.Context, c *Client) error {
				return c.Calendar.FollowPublicEvent(ctx, "AT", "p1", "user_1")
			},
			"Bearer AT", nil,
			url.Values{"event_id": {"p1"}, "calendar_id": {"user_1"}},
		},
		{
			"subscribable calendars", "GET /v2/api/calendar/subscribable/calendars", `{"categories":[]}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.SubscribableCalendars(ctx, "스포츠", "야구")
				return err
			},
			"KakaoAK admin",
			url.Values{"category_name": {"스포츠"}, "subcategory_name": {"야구"}},
			nil,
		},
		{
			"subscribe", "POST /v2/api/calendar/subscribe", `{"calendar_id":"subscribe_1"}`,
			func(ctx context.Context, c *Client) error { return c.Calendar.Subscribe(ctx, "AT", "subscribe_1") },
			"Bearer AT", nil,
			url.Values{"calendar_id": {"subscribe_1"}},
		},
		{
			"unsubscribe", "DELETE /v2/api/calendar/unsubscribe", `{"calendar_id":"subscribe_1"}`,
			func(ctx context.Context, c *Client) error { return c.Calendar.Unsubscribe(ctx, "AT", "subscribe_1") },
			"Bearer AT",
			url.Values{"calendar_id": {"subscribe_1"}},
			nil,
		},
		{
			"holidays", "GET /v2/api/calendar/holidays", `{"events":[]}`,
			func(ctx context.Context, c *Client) error { _, err := c.Calendar.Holidays(ctx, from, to); return err },
			"KakaoAK admin",
			url.Values{"from": {"2022-10-26T00:00:00Z"}, "to": {"2022-10-30T00:00:00Z"}},
			nil,
		},
		{
			"tasks", "GET /v1/api/calendar/tasks", `{"tasks":[],"count":0}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.Tasks(ctx, "AT", TasksQuery{From: "20231208", To: "20231215", Status: TaskFilterAll, Conditions: []TaskCondition{TaskAuthorized, TaskBookmarked}, Offset: 4, Limit: 4, TimeZone: "Asia/Seoul"})
				return err
			},
			"Bearer AT",
			url.Values{"from": {"20231208"}, "to": {"20231215"}, "task_status": {"ALL"}, "task_filter": {"authorized,bookmark"}, "offset": {"4"}, "limit": {"4"}, "time_zone": {"Asia/Seoul"}},
			nil,
		},
		{
			"task records", "GET /v1/api/calendar/task/records", `{"records":[]}`,
			func(ctx context.Context, c *Client) error {
				_, err := c.Calendar.TaskRecords(ctx, "AT", "t1", "202311", "202312")
				return err
			},
			"Bearer AT",
			url.Values{"task_id": {"t1"}, "from": {"202311"}, "to": {"202312"}},
			nil,
		},
		{
			"update task", "POST /v1/api/calendar/update/task", `{"task_id":"t1"}`,
			func(ctx context.Context, c *Client) error {
				return c.Calendar.UpdateTask(ctx, "AT", "t1", TaskChanges{
					DueInfo: &DueInfo{DueDate: "20231231", AlarmTime: "0900"}, Clear: []string{"due_info.recur"},
				})
			},
			"Bearer AT", nil,
			url.Values{"task_id": {"t1"}, "task": {`{"due_info":{"alarm_time":"0900","due_date":"20231231","recur":null}}`}},
		},
		{
			"complete task", "POST /v1/api/calendar/complete/task", `{"task_id":"t1"}`,
			func(ctx context.Context, c *Client) error { return c.Calendar.CompleteTask(ctx, "AT", "t1", true) },
			"Bearer AT", nil,
			url.Values{"task_id": {"t1"}, "complete": {"true"}},
		},
		{
			"delete task", "DELETE /v1/api/calendar/delete/task", `{"task_id":"t1"}`,
			func(ctx context.Context, c *Client) error { return c.Calendar.DeleteTask(ctx, "AT", "t1") },
			"Bearer AT",
			url.Values{"task_id": {"t1"}},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, srv := newTestClient(t, WithAdminKey("admin"))
			srv.HandleJSON(tt.pattern, 200, tt.response)
			if err := tt.call(t.Context(), c); err != nil {
				t.Fatal(err)
			}
			req := srv.LastRequest()
			if got := req.Header.Get("Authorization"); got != tt.wantAuth {
				t.Errorf("Authorization = %q, want %q", got, tt.wantAuth)
			}
			if !equalValues(req.Query, tt.wantQuery) {
				t.Errorf("query = %v\nwant    %v", req.Query, tt.wantQuery)
			}
			if !equalValues(req.Form(), tt.wantForm) {
				t.Errorf("form = %v\nwant   %v", req.Form(), tt.wantForm)
			}
		})
	}
}

func equalValues(got, want url.Values) bool {
	if len(got) != len(want) {
		return false
	}
	for k, v := range want {
		if !slices.Equal(got[k], v) {
			return false
		}
	}
	return true
}

func TestCalendarService_events(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /v2/api/calendar/create/event", 200, `{"event_id":"63630868d89d8b4150bbb712"}`)
	srv.HandleFile("GET /v2/api/calendar/event", "testdata/calendar/event.json")
	srv.HandleJSON("POST /v2/api/calendar/update/event/host", 200, `{"event_id":"e1"}`)

	id, err := c.Calendar.CreateEvent(t.Context(), "AT", "user_63759daa38e1f752188e0cc9", NewEvent{
		Title: "일정 제목",
		Time: EventTime{
			StartAt: time.Date(2022, 10, 27, 12, 0, 0, 0, time.FixedZone("KST", 9*3600)),
			EndAt:   time.Date(2022, 10, 27, 6, 0, 0, 0, time.UTC), TimeZone: "Asia/Seoul",
		},
		RRule:       "FREQ=DAILY;UNTIL=20221031T000000Z",
		Description: "일정 설명",
		Location:    &EventLocation{Name: "카카오", LocationID: 18577297, Address: "경기 성남시 분당구 판교역로 166", Latitude: 37.39570088983171, Longitude: 127.1104335101161},
		Reminders:   []int{15, 30},
		Color:       ColorRed,
	})
	if err != nil || id != "63630868d89d8b4150bbb712" {
		t.Fatalf("CreateEvent = %q, %v", id, err)
	}
	f := srv.LastRequest().Form()
	if f.Get("calendar_id") != "user_63759daa38e1f752188e0cc9" {
		t.Errorf("form = %v", f)
	}
	assertBodyJSON(t, []byte(f.Get("event")), `{"title":"일정 제목","time":{"start_at":"2022-10-27T03:00:00Z","end_at":"2022-10-27T06:00:00Z","time_zone":"Asia/Seoul"},
		"rrule":"FREQ=DAILY;UNTIL=20221031T000000Z","description":"일정 설명",
		"location":{"name":"카카오","location_id":18577297,"address":"경기 성남시 분당구 판교역로 166","latitude":37.39570088983171,"longitude":127.1104335101161},
		"reminders":[15,30],"color":"RED"}`)

	if _, err := c.Calendar.CreateEvent(t.Context(), "AT", "", NewEvent{Title: "t"}); err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Form().Has("calendar_id") {
		t.Error("empty calendar ID sent")
	}

	e, err := c.Calendar.Event(t.Context(), "AT", "6358e3987ec8e318d0b813bc_20220517T120000Z")
	if err != nil {
		t.Fatal(err)
	}
	if e.Title != "일정 제목" || e.Type != EventTypeUser || !e.IsHost || !e.IsRecurEvent || e.Time.TimeZone != "Asia/Seoul" || !e.Time.StartAt.Equal(time.Date(2022, 10, 27, 3, 0, 0, 0, time.UTC)) ||
		e.Location == nil || e.Location.LocationID != 18577297 || !slices.Equal(e.Reminders, []int{15, 30}) ||
		e.Banner == nil || e.Banner.Link == nil || e.Banner.Link.WebURL != "https://example.com" {
		t.Errorf("event = %+v", e)
	}

	err = c.Calendar.UpdateEvent(t.Context(), "AT", "e1",
		EventChanges{Title: ptr("수정"), Reminders: []int{5}, Clear: []string{"description"}},
		&EventUpdateOptions{CalendarID: "primary", RecurUpdateType: RecurAll})
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	f = srv.LastRequest().Form()
	if f.Get("event_id") != "e1" || f.Get("calendar_id") != "primary" || f.Get("recur_update_type") != "ALL" ||
		f.Get("event") != `{"description":null,"reminders":[5],"title":"수정"}` {
		t.Errorf("form = %v", f)
	}
	if err := c.Calendar.UpdateEvent(t.Context(), "AT", "e1", EventChanges{}, &EventUpdateOptions{CalendarID: "primary"}); err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Form().Has("event") {
		t.Error("event sent without changes")
	}
}

func TestCalendarService_AllEvents(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.Handle("GET /v2/api/calendar/events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("next_page_token") == "" {
			_, _ = w.Write([]byte(`{"events":[{"id":"a"},{"id":"b"}],"has_next":true,"after_url":"https://kapi.kakao.com/v2/api/calendar/events?calendar_id=c&next_page_token=eyJmIjoi%3D%3D"}`))
			return
		}
		if r.URL.Query().Get("next_page_token") != "eyJmIjoi==" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"events":[{"id":"c"}],"has_next":false}`))
	})
	var ids []string
	for e, err := range c.Calendar.AllEvents(t.Context(), "AT", EventsQuery{Preset: PresetThisMonth}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, e.ID)
	}
	if !slices.Equal(ids, []string{"a", "b", "c"}) {
		t.Errorf("ids = %v", ids)
	}
	for e, err := range c.Calendar.AllEvents(t.Context(), "AT", EventsQuery{Preset: PresetThisMonth}) {
		if err != nil || e.ID != "a" {
			t.Errorf("%+v, %v", e, err)
		}
		break
	}
	var gotErr error
	for _, err := range c.Calendar.AllEvents(t.Context(), "AT", EventsQuery{}) {
		gotErr = err
	}
	if gotErr == nil {
		t.Error("query without range: want error")
	}
}

func TestCalendarService_publicEvents(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("POST /v2/api/calendar/public/create/event", 200, `{"event_id":"p1"}`)
	srv.HandleJSON("GET /v2/api/calendar/public/event", 200, `{"event":{"id":"p1","title":"공개 일정","type":"PUBLIC","time":{"start_at":"2022-12-01T00:00:00Z","end_at":"2022-12-02T00:00:00Z","all_day":true},
		"notification_message":{"display_custom_button":true,"before_event_reminder":{"button":{"title":"예매하기","link":{"web_url":"https://example.com"}}}}}}`)

	id, err := c.Calendar.CreatePublicEvent(t.Context(), "_xnrxjem", NewPublicEvent{
		Title: "공개 일정",
		Time:  EventTime{StartAt: time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC), EndAt: time.Date(2022, 12, 2, 0, 0, 0, 0, time.UTC), AllDay: true},
		NotificationMessage: &NotificationMessage{
			DisplayCustomButton: ptr(true),
			BeforeEventReminder: &EventReminder{Button: &ReminderButton{Title: "예매하기", Link: CalendarLink{WebURL: "https://example.com"}}},
		},
	})
	if err != nil || id != "p1" {
		t.Fatalf("CreatePublicEvent = %q, %v", id, err)
	}
	f := srv.LastRequest().Form()
	if f.Get("channel_public_id") != "_xnrxjem" {
		t.Errorf("form = %v", f)
	}
	assertBodyJSON(t, []byte(f.Get("event")), `{"title":"공개 일정","time":{"start_at":"2022-12-01T00:00:00Z","end_at":"2022-12-02T00:00:00Z","all_day":true},
		"notification_message":{"display_custom_button":true,"before_event_reminder":{"button":{"title":"예매하기","link":{"web_url":"https://example.com"}}}}}`)

	e, err := c.Calendar.PublicEvent(t.Context(), "_xnrxjem", "p1")
	if err != nil {
		t.Fatal(err)
	}
	if e.Type != EventTypePublic || !e.Time.AllDay || e.NotificationMessage == nil ||
		e.NotificationMessage.BeforeEventReminder.Button.Title != "예매하기" {
		t.Errorf("event = %+v", e)
	}
}

func TestCalendarService_responses(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/api/calendar/calendars", 200, `{"calendars":[{"id":"primary"},{"id":"user_1","name":"서비스 캘린더","color":"RED","reminder":15,"reminder_all_day":30}],
		"subscribe_calendars":[{"id":"subscribe_1","name":"KBO","description":"야구","profile_image_url":"p","thumbnail_url":"t"}]}`)
	srv.HandleJSON("GET /v2/api/calendar/subscribable/calendars", 200, `{"categories":[{"category_name":"스포츠","subcategories":[{"subcategory_name":"야구","calendars":[{"id":"subscribe_1","name":"KBO"}]}]}]}`)
	srv.HandleJSON("GET /v2/api/calendar/holidays", 200, `{"events":[{"id":"h1","title":"한글날","time":{"start_at":"2022-10-09T00:00:00Z","end_at":"2022-10-10T00:00:00Z","all_day":true},"holiday":true}]}`)
	srv.HandleJSON("POST /v1/api/calendar/create/task", 200, `{"task_id":"t1"}`)
	srv.HandleJSON("GET /v1/api/calendar/tasks", 200, `{"tasks":[{"task_id":"t1","content":"오늘의 할 일","status":"SCHEDULED","bookmark":true,"authorized":true,
		"due_info":{"due_date":"20231231","alarm_time":"0900","recur":{"rrule":"FREQ=DAILY;COUNT=3","record_on":true}},"is_ended":false}],"count":1,"has_next":false}`)
	srv.HandleJSON("GET /v1/api/calendar/task/records", 200, `{"records":[{"date":"20231208","complete":true}],"start_at":"202312","end_at":"202312"}`)
	ctx := t.Context()

	cals, err := c.Calendar.Calendars(ctx, "AT", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cals.Calendars) != 2 || *cals.Calendars[1].Reminder != 15 || cals.Calendars[0].Reminder != nil || cals.SubscribeCalendars[0].Description != "야구" {
		t.Errorf("calendars = %+v", cals)
	}
	if srv.LastRequest().Query.Has("filter") {
		t.Error("empty filter sent")
	}

	cats, err := c.Calendar.SubscribableCalendars(ctx, "", "")
	if err != nil || cats[0].Subcategories[0].Calendars[0].ID != "subscribe_1" {
		t.Errorf("categories = %+v, %v", cats, err)
	}

	hs, err := c.Calendar.Holidays(ctx, time.Now(), time.Now().Add(time.Hour))
	if err != nil || len(hs) != 1 || !hs[0].Holiday || hs[0].Title != "한글날" || !hs[0].Time.AllDay {
		t.Errorf("holidays = %+v, %v", hs, err)
	}

	id, err := c.Calendar.CreateTask(ctx, "AT", NewTask{Content: "오늘의 할 일", DueInfo: &DueInfo{
		DueDate: "20231231", TimeZone: "Asia/Seoul", AlarmTime: "0900", Recur: &TaskRecur{RRule: "FREQ=DAILY;COUNT=3", RecordOn: true},
	}})
	if err != nil || id != "t1" {
		t.Fatalf("CreateTask = %q, %v", id, err)
	}
	assertBodyJSON(t, []byte(srv.LastRequest().Form().Get("task")), `{"content":"오늘의 할 일","due_info":{"due_date":"20231231","time_zone":"Asia/Seoul","alarm_time":"0900","recur":{"rrule":"FREQ=DAILY;COUNT=3","record_on":true}}}`)

	task, err := c.Calendar.Task(ctx, "AT", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 || q.Get("task_id") != "t1" {
		t.Errorf("query = %v", q)
	}
	if task.Status != TaskScheduled || !task.Bookmark || task.DueInfo.Recur.RRule != "FREQ=DAILY;COUNT=3" {
		t.Errorf("task = %+v", task)
	}

	recs, err := c.Calendar.TaskRecords(ctx, "AT", "t1", "202312", "202312")
	if err != nil || len(recs.Records) != 1 || !recs.Records[0].Complete || recs.StartAt != "202312" {
		t.Errorf("records = %+v, %v", recs, err)
	}
}

func TestCalendarService_validation(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v1/api/calendar/tasks", 200, `{"tasks":[],"count":0}`)
	ctx := t.Context()
	now := time.Now()
	calls := map[string]func() error{
		"create calendar without name": func() error { _, err := c.Calendar.CreateCalendar(ctx, "AT", CalendarSettings{}); return err },
		"update calendar without ID":   func() error { return c.Calendar.UpdateCalendar(ctx, "AT", "", CalendarSettings{Name: "x"}) },
		"update calendar without data": func() error { return c.Calendar.UpdateCalendar(ctx, "AT", "c", CalendarSettings{}) },
		"delete calendar without ID":   func() error { return c.Calendar.DeleteCalendar(ctx, "AT", "") },
		"events without range":         func() error { _, err := c.Calendar.Events(ctx, "AT", EventsQuery{From: now}); return err },
		"event without ID":             func() error { _, err := c.Calendar.Event(ctx, "AT", ""); return err },
		"update event without ID": func() error {
			return c.Calendar.UpdateEvent(ctx, "AT", "", EventChanges{}, &EventUpdateOptions{CalendarID: "x"})
		},
		"update event without data": func() error { return c.Calendar.UpdateEvent(ctx, "AT", "e", EventChanges{}, nil) },
		"delete event without ID":   func() error { return c.Calendar.DeleteEvent(ctx, "AT", "", nil) },
		"guest update without ID": func() error {
			return c.Calendar.UpdateGuestEvent(ctx, "AT", "", GuestEventChanges{}, &GuestEventUpdateOptions{CalendarID: "c"})
		},
		"guest update without data":    func() error { return c.Calendar.UpdateGuestEvent(ctx, "AT", "e", GuestEventChanges{}, nil) },
		"public events without range":  func() error { _, err := c.Calendar.PublicEvents(ctx, PublicEventsQuery{From: now}); return err },
		"public event without ID":      func() error { _, err := c.Calendar.PublicEvent(ctx, "", ""); return err },
		"update public without ID":     func() error { return c.Calendar.UpdatePublicEvent(ctx, "", "", PublicEventChanges{}) },
		"delete public without ID":     func() error { return c.Calendar.DeletePublicEvent(ctx, "", "") },
		"follow without ID":            func() error { return c.Calendar.FollowPublicEvent(ctx, "AT", "", "") },
		"subcategory without category": func() error { _, err := c.Calendar.SubscribableCalendars(ctx, "", "x"); return err },
		"subscribe without ID":         func() error { return c.Calendar.Subscribe(ctx, "AT", "") },
		"unsubscribe without ID":       func() error { return c.Calendar.Unsubscribe(ctx, "AT", "") },
		"holidays without range":       func() error { _, err := c.Calendar.Holidays(ctx, time.Time{}, now); return err },
		"task without content":         func() error { _, err := c.Calendar.CreateTask(ctx, "AT", NewTask{}); return err },
		"tasks without range":          func() error { _, err := c.Calendar.Tasks(ctx, "AT", TasksQuery{From: "20230101"}); return err },
		"task without ID":              func() error { _, err := c.Calendar.Task(ctx, "AT", ""); return err },
		"task not found":               func() error { _, err := c.Calendar.Task(ctx, "AT", "missing"); return err },
		"records without range":        func() error { _, err := c.Calendar.TaskRecords(ctx, "AT", "t", "", ""); return err },
		"update task without ID":       func() error { return c.Calendar.UpdateTask(ctx, "AT", "", TaskChanges{}) },
		"complete task without ID":     func() error { return c.Calendar.CompleteTask(ctx, "AT", "", true) },
		"delete task without ID":       func() error { return c.Calendar.DeleteTask(ctx, "AT", "") },
		"no access token":              func() error { _, err := c.Calendar.Calendars(ctx, "", ""); return err },
	}
	for name, call := range calls {
		if err := call(); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	if _, err := c.Calendar.Calendars(ctx, "", ""); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no access token: err = %v", err)
	}
}

func TestJSONForm(t *testing.T) {
	t.Parallel()
	f, err := jsonForm("x", map[string]any{"a": 1, "b": map[string]any{"c": 2}}, []string{"b.c", "d.e", "a"})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Get("x"); got != `{"a":null,"b":{"c":null},"d":{"e":null}}` {
		t.Errorf("got %s", got)
	}
	if f, _ := jsonForm("x", map[string]int64{"big": 9007199254740993}, []string{"y"}); f.Get("x") != `{"big":9007199254740993,"y":null}` {
		t.Errorf("large integers must survive: %s", f.Get("x"))
	}
	if _, err := jsonForm("x", func() {}, nil); err == nil {
		t.Error("unencodable: want error")
	}
	if _, err := jsonForm("x", []int{1}, []string{"a"}); err == nil {
		t.Error("clearing fields of a non-object: want error")
	}
	if got := queryParam("://bad", "k"); got != "" {
		t.Errorf("queryParam(bad URL) = %q", got)
	}
}

func TestEventTime_JSON(t *testing.T) {
	t.Parallel()
	b, err := json.Marshal(EventTime{})
	if err != nil || string(b) != `{}` {
		t.Errorf("zero EventTime = %s, %v", b, err)
	}
	var et EventTime
	if err := json.Unmarshal([]byte(`{"start_at":"bad"}`), &et); err == nil {
		t.Error("bad start_at: want error")
	}
}
