package sipdomains

import (
	"fmt"

	"github.com/leamout/leamout/internal/backoffice/components"
)

func listProps(items []SIPDomain) components.ResourceListProps {
	rows := make([][]string, 0, len(items))
	links := make([]components.TableCellLink, 0, len(items)*2)
	for i, v := range items {
		rows = append(
			rows,
			[]string{v.Domain, v.OrganizationName, fmt.Sprint(v.SubscriberCount), v.Status, v.CreatedAt},
		)
		links = append(
			links,
			components.TableCellLink{
				Row:    i,
				Column: 0,
				Href:   "/sip-domains/" + v.ID,
				Class:  "link link-primary font-mono",
			},
			components.TableCellLink{
				Row:    i,
				Column: 1,
				Href:   "/organizations/" + v.OrganizationID,
				Class:  "link link-primary",
			},
		)
	}
	return components.ResourceListProps{
		Title:       "SIP Domains",
		Eyebrow:     "Telecom",
		Description: "Inspect SIP domains, subscriber counts, application bindings, and ownership.",
		Active:      "sipdomains",
		Headers:     []string{"Domain", "Organization", "Subscribers", "Status", "Created"},
		Rows:        rows,
		Links:       links,
	}
}

func detailProps(v Detail) components.ResourceDetailProps {
	return components.ResourceDetailProps{
		Title:     v.Domain,
		Eyebrow:   "SIP domain",
		Active:    "sipdomains",
		BackLabel: "SIP Domains",
		BackHref:  "/sip-domains/",
		ID:        v.ID,
		Status:    v.Status,
		Stats: []components.ResourceField{
			{Label: "Subscribers", Value: fmt.Sprint(v.SubscriberCount)},
			{Label: "Voice bindings", Value: fmt.Sprint(v.BindingCount)},
		},
		Sections: []components.ResourceSection{
			{
				Title: "Ownership",
				Fields: []components.ResourceField{
					{Label: "Organization", Value: v.OrganizationName},
					{Label: "Organization ID", Value: v.OrganizationID},
					{Label: "Domain", Value: v.Domain},
					{Label: "Status", Value: v.Status},
				},
			},
			{
				Title: "Lifecycle",
				Fields: []components.ResourceField{
					{Label: "Created", Value: v.CreatedAt},
					{Label: "Updated", Value: v.UpdatedAt},
				},
			},
		},
	}
}
