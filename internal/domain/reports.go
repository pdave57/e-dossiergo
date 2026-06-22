package domain

// DashboardStats represents aggregated statistics for a dashboard view.
type DashboardStats struct {
	TotalSchools      int `json:"total_schools"`
	TotalStudents     int `json:"total_students"`
	TeachingPersonnel int `json:"teaching_personnel"`
	ResultCompletion  int `json:"result_completion"`
}

// PublicTeachingPersonnel represents the total number of teachers across the entire system.
type PublicTeachingPersonnel struct {
	Total int `json:"total"`
}

// GenderReport represents gender distribution statistics.
type GenderReport struct {
	StateID     string `json:"state_id"`
	Male        int    `json:"male"`
	Female      int    `json:"female"`
	Other       int    `json:"other"`
	Total       int    `json:"total"`
	MalePct     float64 `json:"male_pct"`
	FemalePct   float64 `json:"female_pct"`
	OtherPct    float64 `json:"other_pct"`
}