package main

type matchInfo struct {
	other      string
	otherScene *Scene
	score      float64
}

type matchInfoMap map[string][]matchInfo

func (m *matchInfoMap) add(subject, match string, score float64) {
	existing := (*m)[subject]
	existing = append(existing, matchInfo{
		other: match,
		score: score,
	})

	(*m)[subject] = existing

	existing = (*m)[match]
	existing = append(existing, matchInfo{
		other: subject,
		score: score,
	})

	(*m)[match] = existing
}
