package openai

type VulnerabilityDetectionResponse struct {
	RelevantDiscussion string `json:"relevant_discussion"`
	Reason             string `json:"reason"`
}

type VulnerabilityDetectionResult struct {
	PR                 int    `json:"pr"`
	RelevantDiscussion string `json:"relevant_discussion"`
	Reason             string `json:"reason"`
}
