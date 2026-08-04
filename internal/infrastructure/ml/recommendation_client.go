// internal/infrastructure/ml/recommendation_client.go
//
// Calls the Python ML recommender service (POST /recommend) and
// returns typed results. Wire this into your HTTP handler layer.
//
// Usage in your handler:
//
//   client := ml.NewRecommenderClient("http://localhost:9001", 30*time.Second)
//   result, err := client.Recommend(ctx, schools)

package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ── Input types (mirror Python SchoolInput) ───────────────────────────────────

type SchoolInput struct {
	ID                  int      `json:"id"`
	Name                string   `json:"name"`
	Zone                string   `json:"zone,omitempty"`
	LGA                 string   `json:"lga,omitempty"`
	LevelType           string   `json:"level_type,omitempty"`
	TotalTeachers       int      `json:"total_teachers"`
	QualifiedTeachers   int      `json:"qualified_teachers"`
	TotalStudents       int      `json:"total_students"`
	TotalClassrooms     int      `json:"total_classrooms"`
	FunctionalClassrooms int     `json:"functional_classrooms"`
	HasLibrary          bool     `json:"has_library"`
	HasLaboratory       bool     `json:"has_laboratory"`
	HasToilet           bool     `json:"has_toilet"`
	HasElectricity      bool     `json:"has_electricity"`
	HasWater            bool     `json:"has_water"`
	HasInternet         bool     `json:"has_internet"`
	SubjectsOffered     int      `json:"subjects_offered"`
	ExpectedSubjects    int      `json:"expected_subjects"`
	BooksPerStudent     float64  `json:"books_per_student"`
	AvgPassRate         *float64 `json:"avg_pass_rate,omitempty"`
}

// ── Output types (mirror Python response) ─────────────────────────────────────

type FactorDetail struct {
	Factor         string `json:"factor"`
	Value          string `json:"value"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation"`
}

type SchoolRecommendation struct {
	SchoolID   int            `json:"school_id"`
	SchoolName string         `json:"school_name"`
	Zone       string         `json:"zone"`
	LGA        string         `json:"lga"`
	LevelType  string         `json:"level_type"`
	RiskScore  float64        `json:"risk_score"`
	RiskTier   string         `json:"risk_tier"`
	Anomaly    bool           `json:"anomaly"`
	Factors    []FactorDetail `json:"factors"`
	Summary    string         `json:"summary"`
}

type RecommendResponse struct {
	TotalSchools    int                    `json:"total_schools"`
	FlaggedSchools  int                    `json:"flagged_schools"`
	CriticalCount   int                    `json:"critical_count"`
	HighCount       int                    `json:"high_count"`
	Recommendations []SchoolRecommendation `json:"recommendations"`
}

// ── Client ────────────────────────────────────────────────────────────────────

type RecommenderClient struct {
	baseURL string
	http    *http.Client
}

func NewRecommenderClient(baseURL string, timeout time.Duration) *RecommenderClient {
	return &RecommenderClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *RecommenderClient) Recommend(
	ctx context.Context,
	schools []SchoolInput,
) (*RecommendResponse, error) {
	body, err := json.Marshal(schools)
	if err != nil {
		return nil, fmt.Errorf("ml: marshal request: %w", err)
	}
	

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		c.baseURL+"/recommend",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("ml: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ml: call recommender: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml: recommender returned %d", resp.StatusCode)
	}

	var result RecommendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ml: decode response: %w", err)
	}
	return &result, nil
}

// ── Handler helper (wire into your chi/stdlib router) ─────────────────────────
//
// Example (chi):
//
//   r.Get("/admin/recommendations", handler.GetRecommendations)
//
// func (h *AdminHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
//     schools, err := h.schoolRepo.GetAllWithFacilities(r.Context())
//     if err != nil { /* handle */ }
//
//     inputs := make([]ml.SchoolInput, len(schools))
//     for i, s := range schools {
//         inputs[i] = ml.SchoolInput{
//             ID:                   int(s.ID),
//             Name:                 s.Name,
//             Zone:                 s.Zone.Name,
//             LGA:                  s.LGA.Name,
//             LevelType:            s.LevelType,
//             TotalTeachers:        s.StaffCount,
//             QualifiedTeachers:    s.QualifiedStaffCount,
//             TotalStudents:        s.EnrolmentCount,
//             TotalClassrooms:      s.TotalClassrooms,
//             FunctionalClassrooms: s.FunctionalClassrooms,
//             HasLibrary:           s.HasLibrary,
//             HasLaboratory:        s.HasLaboratory,
//             HasToilet:            s.HasToilet,
//             HasElectricity:       s.HasElectricity,
//             HasWater:             s.HasWater,
//             HasInternet:          s.HasInternet,
//             SubjectsOffered:      s.SubjectsOffered,
//             ExpectedSubjects:     s.ExpectedSubjects,
//             BooksPerStudent:      s.BooksPerStudent,
//             AvgPassRate:          s.AvgPassRate, // *float64, nil if unknown
//         }
//     }
//
//     result, err := h.mlClient.Recommend(r.Context(), inputs)
//     if err != nil { /* handle */ } 
//
//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(result)
// }