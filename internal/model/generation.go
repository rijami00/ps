package model

type ComponentCode struct {
	Name        string `json:"name"`
	Code        string `json:"code,omitempty"`
	Handler     string `json:"handler,omitempty"`
	Description string `json:"description,omitempty"`
	DaisyUIURL  string `json:"daisy_ui_url,omitempty"`

	Label string `json:"-"`
}

type ComponentCodeMap map[string][]ComponentCode

type ComponentExampleCodeMap map[string][]ComponentCode
