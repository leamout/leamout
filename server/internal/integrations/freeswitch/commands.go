package freeswitch

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Command sends a synchronous API command to FreeSWITCH and returns the response.
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

// BGAPI sends a background API command to FreeSWITCH and returns the job UUID.
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

// Originate initiates a new call to the specified endpoint and destination.
func (c *Client) Originate(ctx context.Context, req OriginateRequest) (Call, error) {
	if strings.TrimSpace(req.Endpoint) == "" || strings.TrimSpace(req.Destination) == "" {
		return Call{}, fmt.Errorf("FreeSWITCH originate endpoint and destination are required")
	}

	// Build command with proper spacing
	command := "originate"

	// Add channel variables if present
	if prefix := variables(req.Variables); prefix != "" {
		command += " " + prefix
	}

	// Add caller ID if present
	if req.CallerID != "" {
		callerIDVar := "origination_caller_id_number=" + req.CallerID
		if strings.HasSuffix(command, "}") {
			// Append to existing variables
			command = strings.TrimSuffix(command, "}") + "," + callerIDVar + "}"
		} else {
			// Create new variables block
			command += " {" + callerIDVar + "}"
		}
	}

	// Add endpoint and destination
	command += " " + req.Endpoint + " " + req.Destination

	reply, err := c.Command(ctx, command)
	if err != nil {
		return Call{}, err
	}

	body := strings.TrimSpace(reply.Body)
	if !strings.HasPrefix(body, "+OK") {
		return Call{}, fmt.Errorf("FreeSWITCH originate failed: %s", body)
	}

	return Call{
		UUID:        strings.TrimSpace(strings.TrimPrefix(body, "+OK")),
		Destination: req.Destination,
	}, nil
}

// Hangup terminates an active call by its UUID.
func (c *Client) Hangup(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_kill "+callID)
}

// Hold places a call on hold.
func (c *Client) Hold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_hold "+callID)
}

// Unhold removes a call from hold.
func (c *Client) Unhold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_unhold "+callID)
}

// Break stops all audio playback on a call.
func (c *Client) Break(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.ok(ctx, "uuid_break "+callID+" all")
}

// Transfer transfers a call to a new destination.
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

// Record starts or stops recording a call to the specified path.
func (c *Client) Record(ctx context.Context, req RecordRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("FreeSWITCH record call ID and path are required")
	}

	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "start"
	}
	if action != "start" && action != "stop" {
		return fmt.Errorf("FreeSWITCH record action must be start or stop, got %q", action)
	}
	return c.ok(ctx, "uuid_record "+req.CallID+" "+action+" "+req.Path)
}

// SetVariable sets a channel variable on an active call.
func (c *Client) SetVariable(ctx context.Context, callID, name, value string) error {
	if strings.TrimSpace(callID) == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("FreeSWITCH call ID and variable name are required")
	}
	return c.ok(ctx, "uuid_setvar "+callID+" "+name+" "+value)
}

// GetVariable retrieves a channel variable from an active call.
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

// Channels returns a list of all active channels (calls) in FreeSWITCH.
func (c *Client) Channels(ctx context.Context) ([]Channel, error) {
	reply, err := c.Command(ctx, "show channels as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Channel {
		return Channel{
			UUID:  stringField(row, "uuid"),
			Name:  stringField(row, "name"),
			State: stringField(row, "state"),
		}
	})
}

// Calls returns a list of all active calls with detailed information.
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

// Endpoints returns a list of all registered endpoints.
func (c *Client) Endpoints(ctx context.Context) ([]Endpoint, error) {
	reply, err := c.Command(ctx, "show endpoints as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Endpoint {
		return Endpoint{
			Name: stringField(row, "name"),
			Type: stringField(row, "type"),
			Data: stringField(row, "data"),
		}
	})
}

// SofiaStatus returns the status of a SIP profile or all profiles if empty.
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

// Conference executes a conference command.
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

// ListConferenceMembers returns all members in a conference.
func (c *Client) ListConferenceMembers(ctx context.Context, conference string) (ConferenceMembers, error) {
	conference, err := requiredArgument("conference name", conference)
	if err != nil {
		return ConferenceMembers{}, err
	}

	reply, err := c.Conference(ctx, ConferenceRequest{Name: conference, Command: "list"})
	if err != nil {
		return ConferenceMembers{}, err
	}

	return ConferenceMembers{
		Conference: conference,
		Members:    parseConferenceMembers(reply.Body),
	}, nil
}

// MuteMember mutes a conference member.
func (c *Client) MuteMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "mute", memberID)
}

// UnmuteMember unmutes a conference member.
func (c *Client) UnmuteMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "unmute", memberID)
}

// KickMember removes a member from a conference.
func (c *Client) KickMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "kick", memberID)
}

// DeafMember makes a conference member unable to hear.
func (c *Client) DeafMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "deaf", memberID)
}

// UndeafMember restores hearing to a deaf conference member.
func (c *Client) UndeafMember(ctx context.Context, conference, memberID string) error {
	return c.conferenceMemberCommand(ctx, conference, "undeaf", memberID)
}

// LockConference prevents new members from joining.
func (c *Client) LockConference(ctx context.Context, conference string) error {
	return c.conferenceCommand(ctx, conference, "lock")
}

// UnlockConference allows new members to join.
func (c *Client) UnlockConference(ctx context.Context, conference string) error {
	return c.conferenceCommand(ctx, conference, "unlock")
}

// conferenceMemberCommand executes a command on a specific conference member.
func (c *Client) conferenceMemberCommand(ctx context.Context, conference, command, memberID string) error {
	memberID, err := requiredArgument("conference member ID", memberID)
	if err != nil {
		return err
	}
	return c.conferenceCommand(ctx, conference, command, memberID)
}

// conferenceCommand executes a command on a conference.
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

// ok executes a command and checks for errors in the response.
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

// requiredArgument validates that a required string argument is not empty.
func requiredArgument(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("FreeSWITCH %s is required", name)
	}
	return value, nil
}

// variables converts a map of channel variables to FreeSWITCH format.
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

// decodeRows parses a JSON response into a slice of typed objects.
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

// responseRows extracts the rows array from a FreeSWITCH JSON response.
func responseRows(body string) ([]map[string]any, error) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(body)), &raw); err != nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response: %w", err)
	}

	// Try direct array first
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err == nil {
		return rows, nil
	}

	// Try envelope with rows field
	var envelope struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response envelope: %w", err)
	}
	if envelope.Rows == nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response: missing rows field")
	}
	return envelope.Rows, nil
}

// stringField safely extracts a string field from a map.
func stringField(row map[string]any, key string) string {
	value, ok := row[key]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}

// parseConferenceMembers parses the conference member list output.
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

// isMemberID checks if a string is a valid numeric member ID.
func isMemberID(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return value != ""
}
