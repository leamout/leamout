package calls

import (
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

func validateOrganizationID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("organization context required")
	}
	return nil
}

func require(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", apperror.NewBadRequest(field + " is required")
	}
	return value, nil
}

func validateCreateRequest(req *CreateCallRequest) error {
	if req.ApplicationID != nil && *req.ApplicationID == uuid.Nil {
		return apperror.NewBadRequest("application_id is invalid")
	}
	if req.TrunkID == uuid.Nil {
		return apperror.NewBadRequest("trunk_id is required")
	}

	var err error
	if req.From, err = require(req.From, "from"); err != nil {
		return err
	}
	if req.To, err = require(req.To, "to"); err != nil {
		return err
	}
	return nil
}

func validateRecordRequest(req *RecordRequest) error {
	if req.Action == "" {
		req.Action = "start"
	}
	if req.Action != "start" && req.Action != "stop" {
		return apperror.NewBadRequest("action must be start or stop")
	}
	if req.Action == "start" {
		var err error
		req.Path, err = require(req.Path, "path")
		return err
	}
	return nil
}
