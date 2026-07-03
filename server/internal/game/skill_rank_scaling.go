package game

import "math"

// RankScalingCurve controls how base + per-rank values grow with skill rank.
type RankScalingCurve struct {
	Type            string `json:"type"`
	PercentPerRank  int    `json:"percent_per_rank"`
	LinearPerRank   bool   `json:"linear_per_rank,omitempty"`
}

func defaultRankScalingCurve() RankScalingCurve {
	return RankScalingCurve{Type: "compound_percent", PercentPerRank: 8}
}

func defaultManaScalingCurve() RankScalingCurve {
	return RankScalingCurve{Type: "compound_percent", PercentPerRank: 10}
}

func rankScaledInt(curve RankScalingCurve, base, perRank, rank int) int {
	if rank < 1 {
		rank = 1
	}
	if curve.Type == "" {
		curve = defaultRankScalingCurve()
	}

	switch curve.Type {
	case "linear":
		value := base + perRank*(rank-1)
		if value < 0 {
			return 0
		}

		return value
	case "compound_percent":
		fallthrough
	default:
		pct := curve.PercentPerRank
		if pct < 0 {
			pct = 0
		}
		factor := math.Pow(1.0+float64(pct)/100.0, float64(rank-1))
		value := int(math.Round(float64(base)*factor + float64(perRank)*float64(rank-1)))
		if value < 0 {
			return 0
		}

		return value
	}
}

func rankScaledFloat(curve RankScalingCurve, base, perRank float64, rank int) float64 {
	if rank < 1 {
		rank = 1
	}
	if curve.Type == "" {
		curve = defaultRankScalingCurve()
	}

	switch curve.Type {
	case "linear":
		value := base + perRank*float64(rank-1)
		if value < 0 {
			return 0
		}

		return value
	case "compound_percent":
		fallthrough
	default:
		pct := curve.PercentPerRank
		if pct < 0 {
			pct = 0
		}
		factor := math.Pow(1.0+float64(pct)/100.0, float64(rank-1))
		value := base*factor + perRank*float64(rank-1)
		if value < 0 {
			return 0
		}

		return value
	}
}

func (r *Rules) skillRankScalingCurve() RankScalingCurve {
	if r == nil {
		return defaultRankScalingCurve()
	}
	curve := r.CharacterProgression.SkillRankScaling
	if curve.Type == "" && curve.PercentPerRank == 0 {
		return defaultRankScalingCurve()
	}

	return curve
}

func (r *Rules) skillManaScalingCurve() RankScalingCurve {
	if r == nil {
		return defaultManaScalingCurve()
	}
	curve := r.CharacterProgression.SkillManaScaling
	if curve.Type == "" && curve.PercentPerRank == 0 {
		return defaultManaScalingCurve()
	}

	return curve
}

func (r *Rules) rankScaledInt(base, perRank, rank int) int {
	return rankScaledInt(r.skillRankScalingCurve(), base, perRank, rank)
}

func (r *Rules) rankScaledFloat(base, perRank float64, rank int) float64 {
	return rankScaledFloat(r.skillRankScalingCurve(), base, perRank, rank)
}

func (r *Rules) rankScaledManaCost(base, perRank, rank int) int {
	return rankScaledInt(r.skillManaScalingCurve(), base, perRank, rank)
}
