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
