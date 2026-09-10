package providercdrs

import (
	"fmt"
	"strconv"

	"github.com/leamout/leamout/internal/backoffice/components"
)

func listProps(items []ProviderCDR) components.ResourceListProps {
	rows := make([][]string, 0, len(items))
	links := make([]components.TableCellLink, 0, len(items))
	for i, v := range items {
		state := "unreconciled"
		if v.Reconciled {
			state = "reconciled"
		}
		rows = append(
			rows,
			[]string{
				v.ProviderRecordID,
				v.Provider,
				v.Direction,
				v.OrganizationName,
				fmt.Sprint(v.DurationSeconds),
				money(v.Currency, v.CostMicros),
				state,
				v.StartedAt,
			},
		)
		links = append(
			links,
			components.TableCellLink{
				Row:    i,
				Column: 0,
				Href:   "/provider-cdrs/" + v.ID,
				Class:  "link link-primary font-mono text-xs",
			},
		)
	}
	return components.ResourceListProps{
		Title:       "Provider CDRs",
		Eyebrow:     "Carriers",
		Description: "Reconcile upstream call-detail records with Leamout calls and attributed wholesale charges.",
		Active:      "providercdrs",
		Headers: []string{
			"Provider record",
			"Provider",
			"Direction",
			"Organization",
			"Seconds",
			"Cost",
			"Reconciliation",
			"Started",
		},
		Rows:  rows,
		Links: links,
	}
}

func detailProps(v Detail) components.ResourceDetailProps {
	status := "unreconciled"
	if v.ReconciledAt != "—" {
		status = "reconciled"
	}
	return components.ResourceDetailProps{
		Title:     v.ProviderRecordID,
		Eyebrow:   "Provider CDR",
		Active:    "providercdrs",
		BackLabel: "Provider CDRs",
		BackHref:  "/provider-cdrs/",
		ID:        v.ID,
		Status:    status,
		Stats: []components.ResourceField{
			{Label: "Duration", Value: fmt.Sprint(v.DurationSeconds) + " seconds"},
			{Label: "Provider cost", Value: money(v.Currency, v.CostMicros)},
			{Label: "Wholesale charge", Value: moneyText(v.WholesaleCurrency, v.WholesaleAmountMicros)},
		},
		Sections: []components.ResourceSection{
			{
				Title: "Provider record",
				Fields: []components.ResourceField{
					{Label: "Provider", Value: v.Provider},
					{Label: "Direction", Value: v.Direction},
					{Label: "SIP call ID", Value: v.SipCallID},
					{Label: "Carrier connection", Value: v.CarrierConnectionName},
					{Label: "Carrier connection ID", Value: v.CarrierConnectionID},
					{Label: "Started", Value: v.StartedAt},
				},
			},
			{
				Title: "Reconciliation",
				Fields: []components.ResourceField{
					{Label: "Organization", Value: v.OrganizationName},
					{Label: "Organization ID", Value: v.OrganizationID},
					{Label: "Call ID", Value: v.CallID},
					{Label: "Reconciled", Value: v.ReconciledAt},
					{Label: "Wholesale charge ID", Value: v.WholesaleChargeID},
					{Label: "Created", Value: v.CreatedAt},
				},
			},
		},
	}
}

func money(currency string, micros int64) string {
	return fmt.Sprintf("%s %.6f", currency, float64(micros)/1_000_000)
}

func moneyText(currency, micros string) string {
	amount, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return "—"
	}
	return money(currency, amount)
}
