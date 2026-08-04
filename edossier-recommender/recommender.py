"""
e-Dossier — School Facility Shortage Recommender
=================================================
Algorithm stack:
  1. Feature engineering  — builds a numeric risk profile per school
  2. Isolation Forest     — unsupervised anomaly detection flags schools
                            that deviate from the norm across all features
  3. Weighted Risk Score  — interpretable 0-100 score with per-factor
                            breakdown so staff know *why* a school is flagged
  4. KMeans clustering    — groups schools into risk tiers (Critical /
                            High / Medium / Low) for prioritisation
  5. Recommendation text  — human-readable action for each flagged factor

Input (JSON via POST /recommend):
  List of school objects — see SchoolInput schema below.

Output:
  List of recommendations sorted by risk_score descending.
"""

from __future__ import annotations
import math
from typing import List, Optional
from fastapi.responses import Response

import numpy as np
import pandas as pd
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from sklearn.ensemble import IsolationForest
from sklearn.cluster import KMeans
from sklearn.preprocessing import MinMaxScaler


# ──────────────────────────────────────────────────────────────────────────────
# Pydantic schemas
# ──────────────────────────────────────────────────────────────────────────────

class SchoolInput(BaseModel):
    id: int
    name: str
    zone: Optional[str] = None
    lga: Optional[str] = None
    level_type: Optional[str] = None          # NURSERY | PRIMARY | JSS | SSS | VOCATIONAL

    # Staffing
    total_teachers: int = 0
    qualified_teachers: int = 0               # hold teaching qualification
    total_students: int = 0

    # Infrastructure
    total_classrooms: int = 0
    functional_classrooms: int = 0            # not damaged / in use
    has_library: bool = False
    has_laboratory: bool = False              # science lab
    has_toilet: bool = False
    has_electricity: bool = False
    has_water: bool = False
    has_internet: bool = False

    # Academic
    subjects_offered: int = 0
    expected_subjects: int = 0                # based on level_type curriculum
    books_per_student: float = 0.0            # average across core subjects

    # Performance proxy (if available — optional)
    avg_pass_rate: Optional[float] = None     # 0-100


class FactorDetail(BaseModel):
    factor: str
    value: str
    severity: str                             # critical | high | medium
    recommendation: str


class SchoolRecommendation(BaseModel):
    school_id: int
    school_name: str
    zone: Optional[str]
    lga: Optional[str]
    level_type: Optional[str]
    risk_score: float = Field(..., description="0-100, higher = more at risk")
    risk_tier: str                            # Critical | High | Medium | Low
    anomaly: bool                             # flagged by Isolation Forest
    factors: List[FactorDetail]
    summary: str


class RecommendResponse(BaseModel):
    total_schools: int
    flagged_schools: int
    critical_count: int
    high_count: int
    recommendations: List[SchoolRecommendation]


# ──────────────────────────────────────────────────────────────────────────────
# Feature engineering
# ──────────────────────────────────────────────────────────────────────────────

# Expected subjects per level type (Nigerian curriculum baseline)
EXPECTED_SUBJECTS = {
    "NURSERY": 6,
    "PRIMARY": 9,
    "JSS": 12,
    "SSS": 10,
    "VOCATIONAL": 8,
}

# WHO/UNESCO safe ratios
SAFE_PTR = 40          # pupils per teacher
SAFE_PCR = 35          # pupils per classroom
MIN_BOOKS_PER_STUDENT = 1.0


def _safe_div(num: float, denom: float, default: float = 0.0) -> float:
    return num / denom if denom > 0 else default


def engineer_features(schools: List[SchoolInput]) -> pd.DataFrame:
    rows = []
    for s in schools:
        expected_subj = s.expected_subjects or EXPECTED_SUBJECTS.get(
            (s.level_type or "").upper(), 10
        )

        ptr = _safe_div(s.total_students, s.total_teachers, default=999)
        pcr = _safe_div(
            s.total_students, s.functional_classrooms, default=999
        )
        qual_ratio = _safe_div(s.qualified_teachers, max(s.total_teachers, 1))
        subj_coverage = _safe_div(s.subjects_offered, expected_subj)
        class_func_ratio = _safe_div(
            s.functional_classrooms, max(s.total_classrooms, 1)
        )
        sanitation_ok = int(s.has_toilet)
        utility_score = sum([
            s.has_electricity, s.has_water, s.has_internet
        ]) / 3.0
        infra_score = sum([
            s.has_library, s.has_laboratory,
            sanitation_ok,
            s.has_electricity, s.has_water,
        ]) / 5.0

        rows.append({
            "id": s.id,
            # Raw features fed to ML model
            "ptr": min(ptr, 200),               # cap outliers
            "pcr": min(pcr, 200),
            "qual_ratio": qual_ratio,
            "subj_coverage": subj_coverage,
            "class_func_ratio": class_func_ratio,
            "books_per_student": min(s.books_per_student, 5.0),
            "utility_score": utility_score,
            "infra_score": infra_score,
            "has_library": int(s.has_library),
            "has_lab": int(s.has_laboratory),
            "sanitation": sanitation_ok,
            # Stored for scoring but not used directly in clustering
            "avg_pass_rate": s.avg_pass_rate if s.avg_pass_rate is not None else -1,
            "expected_subj": expected_subj,
            "total_students": s.total_students,
        })
    return pd.DataFrame(rows)


# ──────────────────────────────────────────────────────────────────────────────
# Weighted risk score (interpretable, 0-100)
# ──────────────────────────────────────────────────────────────────────────────

WEIGHTS = {
    "staffing":      0.30,
    "infrastructure":0.25,
    "curriculum":    0.20,
    "facilities":    0.15,
    "performance":   0.10,
}


def compute_risk_score(row: pd.Series, school: SchoolInput) -> tuple[float, List[FactorDetail]]:
    """
    Returns (risk_score 0-100, list of FactorDetail for flagged issues).
    Each dimension is scored 0-1 (1 = worst), then weighted and scaled.
    """
    factors: List[FactorDetail] = []

    # ── Staffing (30%) ──────────────────────────────────────────────────
    ptr_score = min(_safe_div(row["ptr"] - SAFE_PTR, SAFE_PTR, 0), 1.0)
    ptr_score = max(ptr_score, 0)
    qual_score = 1 - row["qual_ratio"]                 # lower qual = higher risk
    staffing_dim = (ptr_score * 0.6 + qual_score * 0.4)

    if row["ptr"] > SAFE_PTR:
        sev = "critical" if row["ptr"] > SAFE_PTR * 1.5 else "high"
        factors.append(FactorDetail(
            factor="Pupil–Teacher Ratio",
            value=f"{row['ptr']:.0f}:1 (safe ≤{SAFE_PTR}:1)",
            severity=sev,
            recommendation=(
                f"Recruit at least "
                f"{math.ceil((school.total_students / SAFE_PTR) - school.total_teachers)} "
                f"additional teachers to reach the safe ratio."
            ),
        ))

    if row["qual_ratio"] < 0.6:
        factors.append(FactorDetail(
            factor="Teacher Qualification Rate",
            value=f"{row['qual_ratio']*100:.0f}% qualified (target ≥60%)",
            severity="high" if row["qual_ratio"] < 0.4 else "medium",
            recommendation=(
                "Enrol unqualified teachers in the state's CPD/NCE bridging programme."
            ),
        ))

    # ── Infrastructure (25%) ────────────────────────────────────────────
    pcr_score = min(_safe_div(row["pcr"] - SAFE_PCR, SAFE_PCR, 0), 1.0)
    pcr_score = max(pcr_score, 0)
    infra_dim = (pcr_score * 0.4 + (1 - row["infra_score"]) * 0.4 +
                 (1 - row["class_func_ratio"]) * 0.2)

    if row["pcr"] > SAFE_PCR:
        sev = "critical" if row["pcr"] > SAFE_PCR * 1.5 else "high"
        factors.append(FactorDetail(
            factor="Classroom Congestion",
            value=f"{row['pcr']:.0f} pupils/classroom (safe ≤{SAFE_PCR})",
            severity=sev,
            recommendation=(
                f"Construct or rehabilitate at least "
                f"{math.ceil(school.total_students / SAFE_PCR - school.functional_classrooms)} "
                f"classrooms to reduce congestion."
            ),
        ))

    if row["class_func_ratio"] < 0.7 and school.total_classrooms > 0:
        pct = int(row["class_func_ratio"] * 100)
        factors.append(FactorDetail(
            factor="Classroom Functionality",
            value=f"Only {pct}% of classrooms functional",
            severity="critical" if pct < 50 else "high",
            recommendation="Prioritise structural repair of damaged classrooms.",
        ))

    # ── Curriculum / academic resources (20%) ───────────────────────────
    book_score = max(1 - _safe_div(row["books_per_student"], MIN_BOOKS_PER_STUDENT), 0)
    subj_score = max(1 - row["subj_coverage"], 0)
    curriculum_dim = book_score * 0.5 + subj_score * 0.5

    if row["subj_coverage"] < 0.8:
        missing = max(0, int(row["expected_subj"] - school.subjects_offered))
        factors.append(FactorDetail(
            factor="Subject Coverage Gap",
            value=f"{school.subjects_offered}/{int(row['expected_subj'])} subjects offered",
            severity="critical" if row["subj_coverage"] < 0.5 else "high",
            recommendation=(
                f"Fill {missing} subject gap(s) by deploying specialist teachers "
                f"or merging classes with a nearby school."
            ),
        ))

    if row["books_per_student"] < MIN_BOOKS_PER_STUDENT:
        factors.append(FactorDetail(
            factor="Textbook Shortage",
            value=f"{row['books_per_student']:.1f} books/student (target ≥1.0)",
            severity="high" if row["books_per_student"] < 0.5 else "medium",
            recommendation=(
                "Request emergency textbook allocation from the Ministry's "
                "instructional materials unit."
            ),
        ))

    # ── Facilities / welfare (15%) ───────────────────────────────────────
    facilities_dim = 0.0
    if not school.has_toilet:
        facilities_dim += 0.4
        factors.append(FactorDetail(
            factor="Sanitation Deficit",
            value="No toilet facility available",
            severity="critical",
            recommendation=(
                "Construct gender-separated toilet blocks — a WASH prerequisite "
                "for school accreditation and girl retention."
            ),
        ))
    if not school.has_water:
        facilities_dim += 0.3
        factors.append(FactorDetail(
            factor="No Potable Water",
            value="No water supply on site",
            severity="critical",
            recommendation=(
                "Install borehole or rainwater harvesting system; "
                "this is a health & sanitation baseline requirement."
            ),
        ))
    if not school.has_electricity:
        facilities_dim += 0.2
        factors.append(FactorDetail(
            factor="No Electricity",
            value="No power supply",
            severity="high",
            recommendation="Connect to grid or install solar — critical for ICT and evening study.",
        ))
    if not school.has_library:
        facilities_dim += 0.05
        factors.append(FactorDetail(
            factor="No Library",
            value="Library absent",
            severity="medium",
            recommendation="Establish a reading room with donated or Ministry-supplied books.",
        ))
    if not school.has_laboratory and (school.level_type or "").upper() in ("JSS", "SSS"):
        facilities_dim += 0.05
        factors.append(FactorDetail(
            factor="No Science Laboratory",
            value="Lab absent for science-bearing level",
            severity="high",
            recommendation=(
                "Construct or equip a basic science lab — required for WAEC/NECO practical exams."
            ),
        ))
    facilities_dim = min(facilities_dim, 1.0)

    # ── Performance (10%) ───────────────────────────────────────────────
    perf_dim = 0.0
    if school.avg_pass_rate is not None and school.avg_pass_rate >= 0:
        perf_dim = max(1 - school.avg_pass_rate / 100, 0)
        if school.avg_pass_rate < 50:
            factors.append(FactorDetail(
                factor="Low Pass Rate",
                value=f"{school.avg_pass_rate:.0f}% average pass rate",
                severity="critical" if school.avg_pass_rate < 30 else "high",
                recommendation=(
                    "Conduct root-cause analysis; cross-reference with staffing and "
                    "resource deficits above. Consider academic improvement plan."
                ),
            ))

    # ── Composite weighted score ─────────────────────────────────────────
    score = (
        staffing_dim    * WEIGHTS["staffing"]      +
        infra_dim       * WEIGHTS["infrastructure"] +
        curriculum_dim  * WEIGHTS["curriculum"]    +
        facilities_dim  * WEIGHTS["facilities"]    +
        perf_dim        * WEIGHTS["performance"]
    ) * 100

    return round(float(np.clip(score, 0, 100)), 2), factors


# ──────────────────────────────────────────────────────────────────────────────
# Risk tier via KMeans clustering
# ──────────────────────────────────────────────────────────────────────────────

TIER_LABELS = ["Low", "Medium", "High", "Critical"]

def assign_tiers(scores: np.ndarray) -> List[str]:
    """
    Use KMeans (k=4) on risk scores to find natural cluster boundaries,
    then label them Low/Medium/High/Critical by cluster centroid order.
    Falls back to percentile thresholds when n < 4.
    """
    n = len(scores)
    if n < 4:
        # Percentile fallback
        tiers = []
        for s in scores:
            if s >= 70:   tiers.append("Critical")
            elif s >= 50: tiers.append("High")
            elif s >= 30: tiers.append("Medium")
            else:         tiers.append("Low")
        return tiers

    km = KMeans(n_clusters=4, n_init=10, random_state=42)
    km.fit(scores.reshape(-1, 1))
    centers = km.cluster_centers_.flatten()
    # Sort cluster labels so label 0 = lowest centroid
    order = np.argsort(centers)
    remap = {old: new for new, old in enumerate(order)}
    labels = [TIER_LABELS[remap[c]] for c in km.labels_]
    return labels


# ──────────────────────────────────────────────────────────────────────────────
# Isolation Forest anomaly detection
# ──────────────────────────────────────────────────────────────────────────────

ML_FEATURE_COLS = [
    "ptr", "pcr", "qual_ratio", "subj_coverage",
    "class_func_ratio", "books_per_student",
    "utility_score", "infra_score",
]

def detect_anomalies(df: pd.DataFrame, contamination: float = 0.15) -> np.ndarray:
    """
    Returns boolean array: True = anomalous (outlier school).
    contamination=0.15 → expects ~15% of schools to be anomalous.
    """
    if len(df) < 5:
        return np.zeros(len(df), dtype=bool)

    X = df[ML_FEATURE_COLS].fillna(0).values
    scaler = MinMaxScaler()
    X_scaled = scaler.fit_transform(X)

    iso = IsolationForest(
        n_estimators=200,
        contamination=contamination,
        random_state=42,
        n_jobs=-1,
    )
    preds = iso.fit_predict(X_scaled)          # -1 = anomaly, 1 = normal
    return preds == -1


# ──────────────────────────────────────────────────────────────────────────────
# Summary text
# ──────────────────────────────────────────────────────────────────────────────

def build_summary(school: SchoolInput, tier: str, anomaly: bool, factors: List[FactorDetail]) -> str:
    critical = [f for f in factors if f.severity == "critical"]
    high     = [f for f in factors if f.severity == "high"]
    anomaly_note = " (statistically anomalous vs peer schools)" if anomaly else ""

    if tier == "Critical":
        top = critical[0].factor if critical else (high[0].factor if high else "multiple deficits")
        return (
            f"{school.name} is at CRITICAL risk{anomaly_note}. "
            f"Immediate intervention required — primary concern: {top}. "
            f"{len(critical)} critical and {len(high)} high-priority issues identified."
        )
    elif tier == "High":
        top = high[0].factor if high else "facility shortages"
        return (
            f"{school.name} has HIGH facility risk{anomaly_note}. "
            f"Urgent attention needed for {top} and {max(len(factors)-1,0)} other area(s)."
        )
    elif tier == "Medium":
        return (
            f"{school.name} has moderate shortfalls{anomaly_note}. "
            f"Plan remediation within the current academic term."
        )
    else:
        return f"{school.name} meets baseline facility standards."


# ──────────────────────────────────────────────────────────────────────────────
# FastAPI app
# ──────────────────────────────────────────────────────────────────────────────

app = FastAPI(
    title="e-Dossier Facility Recommender",
    description="ML-powered school facility shortage detector for Taraba State MoE",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],        # restrict to your Go API origin in production
    allow_methods=["POST", "GET"],
    allow_headers=["*"],
)


@app.get("/")
def root():
    return {
        "service": "e-Dossier Facility Recommender",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health",
        "endpoints": {
            "recommend": "POST /recommend",
        },
    }


@app.get("/favicon.ico")
def favicon():
    return Response(status_code=204)


@app.get("/health")
def health():
    return {"status": "ok", "service": "e-dossier-recommender"}


@app.post("/recommend", response_model=RecommendResponse)
def recommend(schools: List[SchoolInput]):
    if not schools:
        return RecommendResponse(
            total_schools=0, flagged_schools=0,
            critical_count=0, high_count=0,
            recommendations=[],
        )

    df = engineer_features(schools)
    anomalies = detect_anomalies(df)

    # Compute per-school risk scores
    scored = []
    for i, school in enumerate(schools):
        row = df.iloc[i]
        score, factors = compute_risk_score(row, school)
        scored.append((school, score, factors, bool(anomalies[i])))

    # Assign tiers via KMeans on all scores
    all_scores = np.array([s for _, s, _, _ in scored])
    tiers = assign_tiers(all_scores)

    recommendations: List[SchoolRecommendation] = []
    for i, (school, score, factors, anomaly) in enumerate(scored):
        tier = tiers[i]
        summary = build_summary(school, tier, anomaly, factors)
        recommendations.append(SchoolRecommendation(
            school_id=school.id,
            school_name=school.name,
            zone=school.zone,
            lga=school.lga,
            level_type=school.level_type,
            risk_score=score,
            risk_tier=tier,
            anomaly=anomaly,
            factors=factors,
            summary=summary,
        ))

    # Sort by risk score descending
    recommendations.sort(key=lambda r: r.risk_score, reverse=True)

    critical_count = sum(1 for r in recommendations if r.risk_tier == "Critical")
    high_count     = sum(1 for r in recommendations if r.risk_tier == "High")
    flagged        = sum(1 for r in recommendations if r.risk_tier in ("Critical", "High"))

    return RecommendResponse(
        total_schools=len(schools),
        flagged_schools=flagged,
        critical_count=critical_count,
        high_count=high_count,
        recommendations=recommendations,
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("recommender:app", host="0.0.0.0", port=8001, reload=True)