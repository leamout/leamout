package database_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// The test uses a disposable database, a unique schema, and a rolled-back transaction.
func TestOrganizationListFilters(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close(context.Background()) }()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			t.Fatalf("fixture SQL: %v", err)
		}
	}
	schema := "list_filters_" + uuid.New().String()[:8]
	exec("CREATE SCHEMA " + pgx.Identifier{schema}.Sanitize())
	exec("SET LOCAL search_path TO " + pgx.Identifier{schema}.Sanitize() + ", public")
	migrations, err := filepath.Glob("../../migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	for _, path := range migrations {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	q := sqlc.New(tx)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	before := from.Add(time.Hour)
	orgs := [2]uuid.UUID{uuid.New(), uuid.New()}
	var trunks, agents, calls, numbers, recordings, events [2][3]uuid.UUID
	for tenant, org := range orgs {
		exec("INSERT INTO organizations(id, name) VALUES ($1, 'filter test')", org)
		for i := range 3 {
			stamp := from
			if i == 2 {
				stamp = before
			}
			trunks[tenant][i], agents[tenant][i], calls[tenant][i] = uuid.New(), uuid.New(), uuid.New()
			numbers[tenant][i], recordings[tenant][i], events[tenant][i] = uuid.New(), uuid.New(), uuid.New()
			direction, status, enabled := "inbound", "active", i == 1
			if i == 0 {
				direction, status = "outbound", "disabled"
			}
			exec("INSERT INTO trunks(id, organization_id, name, direction, status, inbound_enabled, created_at) VALUES($1,$2,$3,$4,$5,$6,$7)", trunks[tenant][i], org, trunks[tenant][i].String(), direction, status, enabled, stamp)
			engine, language := "realtime", "en"
			if i == 0 {
				engine = "composable"
			}
			if i == 1 {
				language = "fr"
			}
			exec("INSERT INTO voice_agents(id, organization_id, name, instructions, engine, language, created_at) VALUES($1,$2,'test','test',$3,$4,$5)", agents[tenant][i], org, engine, language, stamp)
			country := "GH"
			if i == 1 {
				country = "US"
			}
			exec("INSERT INTO phone_numbers(id, organization_id, number, country_code, trunk_id, status, voice_enabled, created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", numbers[tenant][i], org, "+12025550"+string(rune('0'+tenant))+string(rune('0'+i)), country, trunks[tenant][i], status, enabled, stamp)
			state, recordingStatus, action := "completed", "completed", "call.completed"
			if i == 0 {
				state, recordingStatus, action = "failed", "failed", "call.failed"
			}
			exec("INSERT INTO calls(id, organization_id, direction, state, trunk_id, voice_agent_id, from_uri, to_uri, created_at) VALUES($1,$2,$3,$4,$5,$6,'from','to',$7)", calls[tenant][i], org, direction, state, trunks[tenant][i], agents[tenant][i], stamp)
			exec("INSERT INTO recordings(id, organization_id, call_id, status, created_at) VALUES($1,$2,$3,$4,$5)", recordings[tenant][i], org, calls[tenant][i], recordingStatus, stamp)
			actorType := "user"
			if i == 2 {
				actorType = "organization_token"
			}
			exec("INSERT INTO audit_events(id, organization_id, actor_type, actor_id, action, target_type, target_id, occurred_at) VALUES($1,$2,$3,$4,$5,'call',$4,$6)", events[tenant][i], org, actorType, calls[tenant][i], action, stamp)
		}
		exec("INSERT INTO recordings(organization_id, call_id, status) VALUES($1,$2,'deleted')", org, calls[tenant][0])
		exec("INSERT INTO phone_numbers(organization_id, number, country_code, status) VALUES($1,$2,'GH','released')", org, "+12025559"+string(rune('0'+tenant))+"0")
		exec("INSERT INTO voice_agents(organization_id, name, instructions, engine, language, status) VALUES($1,'hidden','test','realtime','fr','disabled')", org)
	}
	check := func(t *testing.T, got, want []uuid.UUID, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("IDs = %v, want %v", got, want)
		}
	}
	ordered := func(all [3]uuid.UUID) []uuid.UUID {
		first := []uuid.UUID{all[0], all[1]}
		sort.Slice(first, func(i, j int) bool { return first[i].String() > first[j].String() })
		return append([]uuid.UUID{all[2]}, first...)
	}
	str := func(v string) *string { return &v }
	no := false
	t.Run("calls", func(t *testing.T) {
		list := func(p sqlc.ListCallsParams) ([]uuid.UUID, error) {
			rows, err := q.ListCalls(ctx, p)
			ids := []uuid.UUID{}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			return ids, err
		}
		p := sqlc.ListCallsParams{OrganizationID: orgs[0], PageLimit: 50}
		ids, err := list(p)
		check(t, ids, ordered(calls[0]), err)
		p.State, p.Direction, p.TrunkID, p.VoiceAgentID = str("completed"), str("inbound"), &trunks[0][1], &agents[0][1]
		p.CreatedFrom, p.CreatedBefore = pgconv.TimeToTimestamptz(from), pgconv.TimeToTimestamptz(before)
		ids, err = list(p)
		check(t, ids, []uuid.UUID{calls[0][1]}, err)
		p.TrunkID = &trunks[1][1]
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
		p = sqlc.ListCallsParams{OrganizationID: orgs[0], PageLimit: 1, PageOffset: 1}
		ids, err = list(p)
		check(t, ids, ordered(calls[0])[1:2], err)
	})
	t.Run("recordings", func(t *testing.T) {
		list := func(p sqlc.ListRecordingsParams) ([]uuid.UUID, error) {
			rows, err := q.ListRecordings(ctx, p)
			ids := []uuid.UUID{}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			return ids, err
		}
		p := sqlc.ListRecordingsParams{OrganizationID: orgs[0], PageLimit: 50}
		ids, err := list(p)
		check(t, ids, ordered(recordings[0]), err)
		p.Status, p.CallID = str("completed"), &calls[0][1]
		p.CreatedFrom, p.CreatedBefore = pgconv.TimeToTimestamptz(from), pgconv.TimeToTimestamptz(before)
		ids, err = list(p)
		check(t, ids, []uuid.UUID{recordings[0][1]}, err)
		p.CallID = &calls[1][1]
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
		p = sqlc.ListRecordingsParams{OrganizationID: orgs[0], PageLimit: 50, CreatedFrom: pgconv.TimeToTimestamptz(from), CreatedBefore: pgconv.TimeToTimestamptz(before)}
		ids, err = list(p)
		check(t, ids, ordered(recordings[0])[1:], err)
		p.Status = str("deleted")
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
	})
	t.Run("numbers", func(t *testing.T) {
		list := func(p sqlc.ListPhoneNumbersByOrganizationIDParams) ([]uuid.UUID, error) {
			rows, err := q.ListPhoneNumbersByOrganizationID(ctx, p)
			ids := []uuid.UUID{}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			return ids, err
		}
		p := sqlc.ListPhoneNumbersByOrganizationIDParams{OrganizationID: orgs[0]}
		ids, err := list(p)
		check(t, ids, ordered(numbers[0]), err)
		p.Status, p.CountryCode, p.VoiceEnabled, p.TrunkID = str("active"), str("GH"), &no, &trunks[0][2]
		ids, err = list(p)
		check(t, ids, []uuid.UUID{numbers[0][2]}, err)
		p.TrunkID = &trunks[1][2]
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
		p = sqlc.ListPhoneNumbersByOrganizationIDParams{OrganizationID: orgs[0], Status: str("released")}
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
	})
	t.Run("trunks", func(t *testing.T) {
		list := func(p sqlc.ListTrunksByOrganizationIDParams) ([]uuid.UUID, error) {
			rows, err := q.ListTrunksByOrganizationID(ctx, p)
			ids := []uuid.UUID{}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			return ids, err
		}
		p := sqlc.ListTrunksByOrganizationIDParams{OrganizationID: orgs[0]}
		ids, err := list(p)
		check(t, ids, ordered(trunks[0]), err)
		p.Status, p.Direction, p.InboundEnabled = str("active"), str("inbound"), &no
		ids, err = list(p)
		check(t, ids, []uuid.UUID{trunks[0][2]}, err)
		p.Status = str("disabled")
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
	})
	t.Run("agents", func(t *testing.T) {
		list := func(p sqlc.ListVoiceAgentsByOrganizationIDParams) ([]uuid.UUID, error) {
			rows, err := q.ListVoiceAgentsByOrganizationID(ctx, p)
			ids := []uuid.UUID{}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			return ids, err
		}
		p := sqlc.ListVoiceAgentsByOrganizationIDParams{OrganizationID: orgs[0]}
		ids, err := list(p)
		check(t, ids, ordered(agents[0]), err)
		p.Engine, p.Language = str("realtime"), str("fr")
		ids, err = list(p)
		check(t, ids, []uuid.UUID{agents[0][1]}, err)
		exec("UPDATE organizations SET status='disabled' WHERE id=$1", orgs[0])
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
		exec("UPDATE organizations SET status='active', deleted_at=now() WHERE id=$1", orgs[0])
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
		exec("UPDATE organizations SET deleted_at=NULL WHERE id=$1", orgs[0])
	})
	t.Run("audit", func(t *testing.T) {
		list := func(p sqlc.ListAuditEventsByOrganizationIDParams) ([]uuid.UUID, error) {
			rows, err := q.ListAuditEventsByOrganizationID(ctx, p)
			ids := []uuid.UUID{}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			return ids, err
		}
		p := sqlc.ListAuditEventsByOrganizationIDParams{OrganizationID: orgs[0], LimitCount: 50}
		ids, err := list(p)
		check(t, ids, ordered(events[0]), err)
		p.Action, p.ActorType, p.ActorID, p.TargetType, p.TargetID = str("call.completed"), str("user"), &calls[0][1], str("call"), &calls[0][1]
		p.OccurredFrom, p.OccurredBefore = pgconv.TimeToTimestamptz(from), pgconv.TimeToTimestamptz(before)
		ids, err = list(p)
		check(t, ids, []uuid.UUID{events[0][1]}, err)
		p.ActorID = &calls[1][1]
		ids, err = list(p)
		check(t, ids, []uuid.UUID{}, err)
		p = sqlc.ListAuditEventsByOrganizationIDParams{OrganizationID: orgs[0], LimitCount: 50, OccurredFrom: pgconv.TimeToTimestamptz(from), OccurredBefore: pgconv.TimeToTimestamptz(before)}
		ids, err = list(p)
		check(t, ids, ordered(events[0])[1:], err)
	})
}
