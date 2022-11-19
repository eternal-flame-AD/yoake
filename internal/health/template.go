package health

const commTemplate = `
The following doses are due:

{{ range . }}
---
### {{ .Med.Name }} _{{ .Med.DirectionShorthand }}_

{{ $isPRN := false -}}
{{- range .Med.Flags -}}
{{- if eq . "prn" -}}
{{- $isPRN = true -}}
{{end}}{{- end -}}

{{ if not $isPRN -}}
Expected at: {{ .Dose.Expected.Time }}

{{if .Dose.EffectiveLastDose -}}
Last Taken at: {{ .Dose.EffectiveLastDose.Actual.Time }}

{{ end -}}
Offset: {{ .Dose.DoseOffset }}
{{ else -}}
avail as PRN
{{- end }}
{{ end }}
`

type CommCtx struct {
	Med  Direction
	Dose ComplianceLog
}
