package freeswitch

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func (c *Client) Command(ctx context.Context, command string) (Reply, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Reply{}, fmt.Errorf("FreeSWITCH command is required")
	}

	frame, err := c.command(ctx, "api "+command)
	if err != nil {
		return Reply{}, err
	}
	if frame.ContentType != ContentTypeAPIResponse {
		return Reply{}, fmt.Errorf("unexpected FreeSWITCH command response: %q", frame.ContentType)
	}

	return Reply{Text: frame.ReplyText(), Body: frame.Body}, nil
}

func (c *Client) Originate(ctx context.Context, req OriginateRequest) (Call, error) {
	if strings.TrimSpace(req.Endpoint) == "" || strings.TrimSpace(req.Destination) == "" {
		return Call{}, fmt.Errorf("FreeSWITCH originate endpoint and destination are required")
	}

	prefix := variables(req.Variables)
	if req.CallerID != "" {
		prefix = mergeVariable(prefix, "origination_caller_id_number", req.CallerID)
	}

	reply, err := c.Command(ctx, "originate "+prefix+req.Endpoint+" "+req.Destination)
	if err != nil {
		return Call{}, err
	}
	body := strings.TrimSpace(reply.Body)
	if !strings.HasPrefix(body, "+OK") {
		return Call{}, fmt.Errorf("FreeSWITCH originate failed: %s", body)
	}

	return Call{UUID: strings.TrimSpace(strings.TrimPrefix(body, "+OK")), Destination: req.Destination}, nil
}

func (c *Client) Hangup(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_kill "+callID)
}

func (c *Client) Hold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_hold "+callID)
}

func (c *Client) Unhold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_unhold "+callID)
}

func (c *Client) Break(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_break "+callID+" all")
}

func (c *Client) Transfer(ctx context.Context, req TransferRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Destination) == "" {
		return fmt.Errorf("FreeSWITCH transfer call ID and destination are required")
	}

	args := []string{req.CallID, req.Destination}
	if req.Dialplan != "" {
		args = append(args, req.Dialplan)
	}
	if req.Context != "" {
		args = append(args, req.Context)
	}
	return c.ok(ctx, "uuid_transfer "+strings.Join(args, " "))
}

func (c *Client) Record(ctx context.Context, req RecordRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("FreeSWITCH record call ID and path are required")
	}

	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "start"
	}
	if action != "start" && action != "stop" {
		return fmt.Errorf("FreeSWITCH record action must be start or stop")
	}
	return c.ok(ctx, "uuid_record "+req.CallID+" "+action+" "+req.Path)
}

func (c *Client) SetVariable(ctx context.Context, callID, name, value string) error {
	if strings.TrimSpace(callID) == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("FreeSWITCH call ID and variable name are required")
	}
	return c.ok(ctx, "uuid_setvar "+callID+" "+name+" "+value)
}

func (c *Client) GetVariable(ctx context.Context, callID, name string) (string, error) {
	if strings.TrimSpace(callID) == "" || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("FreeSWITCH call ID and variable name are required")
	}

	reply, err := c.Command(ctx, "uuid_getvar "+callID+" "+name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply.Body), nil
}

func (c *Client) Channels(ctx context.Context) ([]Channel, error) {
	reply, err := c.Command(ctx, "show channels as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Channel {
		return Channel{UUID: stringField(row, "uuid"), Name: stringField(row, "name"), State: stringField(row, "state")}
	})
}

func (c *Client) Calls(ctx context.Context) ([]Call, error) {
	reply, err := c.Command(ctx, "show calls as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Call {
		return Call{
			UUID:         stringField(row, "uuid"),
			CallerName:   stringField(row, "cid_name"),
			CallerNumber: stringField(row, "cid_num"),
			Destination:  stringField(row, "dest"),
			State:        stringField(row, "state"),
		}
	})
}

func (c *Client) Endpoints(ctx context.Context) ([]Endpoint, error) {
	reply, err := c.Command(ctx, "show endpoints as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Endpoint {
		return Endpoint{Name: stringField(row, "name"), Type: stringField(row, "type"), Data: stringField(row, "data")}
	})
}

func (c *Client) SofiaStatus(ctx context.Context, profile string) (SIPProfileStatus, error) {
	profile = strings.TrimSpace(profile)
	command := "sofia status"
	if profile != "" {
		command += " profile " + profile
	}

	reply, err := c.Command(ctx, command)
	if err != nil {
		return SIPProfileStatus{}, err
	}
	return SIPProfileStatus{Profile: profile, Raw: reply.Body}, nil
}

func (c *Client) Conference(ctx context.Context, req ConferenceRequest) (ConferenceResult, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Command) == "" {
		return ConferenceResult{}, fmt.Errorf("FreeSWITCH conference name and command are required")
	}

	args := append([]string{req.Name, req.Command}, req.Arguments...)
	reply, err := c.Command(ctx, "conference "+strings.Join(args, " "))
	if err != nil {
		return ConferenceResult{}, err
	}
	return ConferenceResult(reply), nil
}

func (c *Client) ListConferenceMembers(ctx context.Context, conference string) (ConferenceMembers, error) {
	conference, err := requiredArgument("conference name", conference)
	if err != nil {
		return ConferenceMembers{}, err
	}

	reply, err := c.Conference(ctx, ConferenceRequest{Name: conference, Command: "list"})
	if err != nil {
		return ConferenceMembers{}, err
	}

	return ConferenceMembers{Conference: conference, Members: parseConferenceMembers(reply.Body)}, nil
}

func (c *Client) MuteMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "mute", memberID)
}

func (c *Client) UnmuteMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "unmute", memberID)
}

func (c *Client) KickMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "kick", memberID)
}

func (c *Client) DeafMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "deaf", memberID)
}

func (c *Client) UndeafMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "undeaf", memberID)
}

func (c *Client) LockConference(ctx context.Context, conference string) error {
	return c.conferenceCommand(ctx, conference, "lock")
}

func (c *Client) UnlockConference(ctx context.Context, conference string) error {
	return c.conferenceCommand(ctx, conference, "unlock")
}

func (c *Client) conferenceMemberCommand(ctx context.Context, conference, command, memberID string) error {
	memberID, err := requiredArgument("conference member ID", memberID)
	if err != nil {
		return err
	}
	return c.conferenceCommand(ctx, conference, command, memberID)
}

func (c *Client) conferenceCommand(ctx context.Context, conference, command string, args ...string) error {
	conference, err := requiredArgument("conference name", conference)
	if err != nil {
		return err
	}
	command, err = requiredArgument("conference command", command)
	if err != nil {
		return err
	}

	_, err = c.Conference(ctx, ConferenceRequest{Name: conference, Command: command, Arguments: args})
	return err
}

func (c *Client) ok(ctx context.Context, command string) error {
	reply, err := c.Command(ctx, command)
	if err != nil {
		return err
	}
	body := strings.TrimSpace(reply.Body)
	if strings.HasPrefix(body, "-ERR") {
		return fmt.Errorf("FreeSWITCH command failed: %s", body)
	}
	return nil
}

func requiredArgument(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("FreeSWITCH %s is required", name)
	}
	return value, nil
}

func variables(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func mergeVariable(prefix, key, value string) string {
	if prefix == "" {
		return "{" + key + "=" + value + "}"
	}
	return strings.TrimSuffix(prefix, "}") + "," + key + "=" + value + "}"
}

func decodeRows[T any](body string, mapRow func(map[string]any) T) ([]T, error) {
	rows, err := responseRows(body)
	if err != nil {
		return nil, err
	}

	result := make([]T, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapRow(row))
	}
	return result, nil
}

func responseRows(body string) ([]map[string]any, error) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(body)), &raw); err != nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response: %w", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err == nil {
		return rows, nil
	}

	var envelope struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Rows == nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response rows")
	}
	return envelope.Rows, nil
}

func stringField(row map[string]any, key string) string {
	value, ok := row[key]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}

func parseConferenceMembers(body string) []ConferenceMember {
	var members []ConferenceMember
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(strings.ToLower(line), "member id") {
			continue
		}

		fields := strings.FieldsFunc(line, func(r rune) bool { return r == '|' || r == ',' })
		if len(fields) == 0 {
			continue
		}

		member := ConferenceMember{ID: strings.TrimSpace(fields[0])}
		if member.ID == "" || !isMemberID(member.ID) {
			continue
		}
		if len(fields) > 1 {
			member.CallerID = strings.TrimSpace(fields[1])
		}
		members = append(members, member)
	}
	return members
}

func isMemberID(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return value != ""
}

func (c *Client) BGAPI(ctx context.Context, command string) (Job, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Job{}, fmt.Errorf("FreeSWITCH background command is required")
	}

	frame, err := c.command(ctx, "bgapi "+command)
	if err != nil {
		return Job{}, err
	}
	if !frame.OK() {
		return Job{}, fmt.Errorf("FreeSWITCH background command failed: %s", frame.ReplyText())
	}
	return Job{ID: frame.Header("Job-UUID")}, nil
}
