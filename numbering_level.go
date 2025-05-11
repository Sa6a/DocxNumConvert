package main

type NumberingLevel struct {
	FormatType   string
	TextTemplate string
	StartValue   int
	CurrentValue int
}

func NewNumberingLevel(formatType, textTemplate string, startValue int) *NumberingLevel {
	return &NumberingLevel{
		FormatType:   formatType,
		TextTemplate: textTemplate,
		StartValue:   startValue,
		CurrentValue: startValue,
	}
}

func (nl *NumberingLevel) Reset() {
	nl.CurrentValue = nl.StartValue
}

func (nl *NumberingLevel) Increment() {
	nl.CurrentValue++
}

func (nl *NumberingLevel) FormatCurrentValue() string {
	return formatNumber(nl.CurrentValue, nl.FormatType)
}
