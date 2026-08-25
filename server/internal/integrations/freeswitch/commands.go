package freeswitch

import (
	"context"
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
	if !strings.HasPrefix(reply.Body, "+OK") {
		return Call{}, fmt.Errorf("FreeSWITCH originate failed: %s", strings.TrimSpace(reply.Body))
	}

	return Call{
		UUID:        strings.TrimSpace(strings.TrimPrefix(reply.Body, "+OK")),
		Destination: req.Destination,
	}, nil
}

func (c *Client) Hangup(ctx context.Context, callID string) error {
	return c.ok(ctx, "uuid_kill "+required("call ID", callID))
}

func (c *Client) Hold(ctx context.Context, callID string) error {
	return c.ok(ctx, "uuid_hold "+required("call ID", callID))
}

func (c *Client) Unhold(ctx context.Context, callID string) error {
	return c.ok(ctx, "uuid_unhold "+required("call ID", callID))
}

func (c *Client) Break(ctx context.Context, callID string) error {
	return c.ok(ctx, "uuid_break "+required("call ID", callID)+" all")
}

func (c *Client) Transfer(ctx context.Context, req TransferRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Destination) == "" {
		return fmt.Errorf("FreeSWITCH transfer call ID and destination are required")
	}

	command := "uuid_transfer " + req.CallID + " " + req.Destination
	if req.Dialplan != "" {
		command += " " + req.Dialplan
	}
	if req.Context != "" {
		command += " " + req.Context
	}

	return c.ok(ctx, command)
}

func (c *Client) Record(ctx context.Context, req RecordRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("FreeSWITCH record call ID and path are required")
	}

	action := req.Action
	if action == "" {
		action = "start"
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

func (c *Client) Channels(ctx context.Context) (Reply, error) {
	return c.Command(ctx, "show channels")
}

func (c *Client) Calls(ctx context.Context) (Reply, error) {
	return c.Command(ctx, "show calls")
}

func (c *Client) Endpoints(ctx context.Context) (Reply, error) {
	return c.Command(ctx, "show endpoints")
}

func (c *Client) SofiaStatus(ctx context.Context, profile string) (SIPProfileStatus, error) {
	profile = strings.TrimSpace(profile)
	command := "sofia status"
	if profile != "" {
		command += " " + profile
	}

	reply, err := c.Command(ctx, command)
	if err != nil {
		return SIPProfileStatus{}, err
	}

	return SIPProfileStatus{
		Profile: profile,
		Raw:     reply.Body,
	}, nil
}

func (c *Client) Conference(ctx context.Context, req ConferenceRequest) (ConferenceResult, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Command) == "" {
		return ConferenceResult{}, fmt.Errorf("FreeSWITCH conference name and command are required")
	}

	command := "conference " + req.Name + " " + req.Command
	if len(req.Arguments) > 0 {
		command += " " + strings.Join(req.Arguments, " ")
	}

	reply, err := c.Command(ctx, command)
	if err != nil {
		return ConferenceResult{}, err
	}

	return ConferenceResult{
		Text: reply.Text,
		Body: reply.Body,
	}, nil
}

func (c *Client) MuteMember(ctx context.Context, conference, memberID string) error {
	_, err := c.Conference(ctx, ConferenceRequest{
		Name:      conference,
		Command:   "mute",
		Arguments: []string{memberID},
	})
	return err
}

func (c *Client) UnmuteMember(ctx context.Context, conference, memberID string) error {
	_, err := c.Conference(ctx, ConferenceRequest{
		Name:      conference,
		Command:   "unmute",
		Arguments: []string{memberID},
	})
	return err
}

func (c *Client) KickMember(ctx context.Context, conference, memberID string) error {
	_, err := c.Conference(ctx, ConferenceRequest{
		Name:      conference,
		Command:   "kick",
		Arguments: []string{memberID},
	})
	return err
}

func (c *Client) ok(ctx context.Context, command string) error {
	reply, err := c.Command(ctx, command)
	if err != nil {
		return err
	}
	if strings.HasPrefix(reply.Body, "-ERR") {
		return fmt.Errorf("FreeSWITCH command failed: %s", strings.TrimSpace(reply.Body))
	}

	return nil
}

func required(name, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return value
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
