package rtpengine

import (
	"context"
	"fmt"
)

func (c *Client) Offer(ctx context.Context, request OfferRequest) (OfferResponse, error) {
	if err := request.Validate(); err != nil {
		return OfferResponse{}, err
	}

	response, err := c.do(ctx, CommandOffer, request.Session.params(request.SDP, request.Flags))
	if err != nil {
		return OfferResponse{}, err
	}

	return OfferResponse{
		SDP:  response.String("sdp"),
		Data: response.Data,
	}, nil
}

func (c *Client) Answer(ctx context.Context, request AnswerRequest) (AnswerResponse, error) {
	if err := request.Validate(); err != nil {
		return AnswerResponse{}, err
	}

	response, err := c.do(ctx, CommandAnswer, request.Session.params(request.SDP, request.Flags))
	if err != nil {
		return AnswerResponse{}, err
	}

	return AnswerResponse{
		SDP:  response.String("sdp"),
		Data: response.Data,
	}, nil
}

func (c *Client) Delete(ctx context.Context, request DeleteRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}

	_, err := c.do(ctx, CommandDelete, request.Session.params("", request.Flags))
	return err
}

func (c *Client) Query(ctx context.Context, session Session) (QueryResponse, error) {
	if err := session.Validate(); err != nil {
		return QueryResponse{}, err
	}

	response, err := c.do(ctx, CommandQuery, session.params("", nil))
	if err != nil {
		return QueryResponse{}, err
	}

	return QueryResponse{Data: response.Data}, nil
}

func (r OfferRequest) Validate() error {
	if err := r.Session.Validate(); err != nil {
		return err
	}
	if r.SDP == "" {
		return fmt.Errorf("RTPEngine offer SDP is required")
	}
	return nil
}

func (r AnswerRequest) Validate() error {
	if err := r.Session.Validate(); err != nil {
		return err
	}
	if r.SDP == "" {
		return fmt.Errorf("RTPEngine answer SDP is required")
	}
	return nil
}

func (r DeleteRequest) Validate() error {
	return r.Session.Validate()
}
