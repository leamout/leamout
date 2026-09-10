package conferences

import (
	"fmt"

	"github.com/leamout/leamout/internal/backoffice/components"
)

func listProps(items []Conference) components.ResourceListProps {
	rows := make([][]string, 0, len(items))
	links := make([]components.TableCellLink, 0, len(items)*2)
	for i, v := range items {
		rows = append(
			rows,
			[]string{v.Name, v.OrganizationName, fmt.Sprint(v.ParticipantCount), v.State, v.StartedAt, v.EndedAt},
		)
		links = append(
			links,
			components.TableCellLink{Row: i, Column: 0, Href: "/conferences/" + v.ID, Class: "link link-primary"},
			components.TableCellLink{
				Row:    i,
				Column: 1,
				Href:   "/organizations/" + v.OrganizationID,
				Class:  "link link-primary",
			},
		)
	}
	return components.ResourceListProps{
		Title:       "Conferences",
		Eyebrow:     "Telecom",
		Description: "Inspect conference sessions, applications, participant counts, and timing.",
		Active:      "conferences",
		Headers:     []string{"Conference", "Organization", "Participants", "State", "Started", "Ended"},
		Rows:        rows,
		Links:       links,
	}
}

func detailProps(v Detail) components.ResourceDetailProps {
	return components.ResourceDetailProps{
		Title:     v.Name,
		Eyebrow:   "Conference",
		Active:    "conferences",
		BackLabel: "Conferences",
		BackHref:  "/conferences/",
		ID:        v.ID,
		Status:    v.State,
		Stats: []components.ResourceField{
			{Label: "Participants", Value: fmt.Sprint(v.ParticipantCount)},
			{Label: "Currently active", Value: fmt.Sprint(v.ActiveParticipantCount)},
		},
		Sections: []components.ResourceSection{
			{
				Title: "Application",
				Fields: []components.ResourceField{
					{Label: "Voice application", Value: v.ApplicationName},
					{Label: "Application ID", Value: v.ApplicationID},
					{Label: "Organization", Value: v.OrganizationName},
					{Label: "Organization ID", Value: v.OrganizationID},
				},
			},
			{
				Title: "Timeline",
				Fields: []components.ResourceField{
					{Label: "Started", Value: v.StartedAt},
					{Label: "Ended", Value: v.EndedAt},
					{Label: "Created", Value: v.CreatedAt},
					{Label: "Updated", Value: v.UpdatedAt},
				},
			},
		},
	}
}
