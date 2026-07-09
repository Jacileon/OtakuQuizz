package handlers

import (
	"encoding/json"
	"math/rand"
	"time"
)

// Fisher-Yates shuffle with a seeded random
func fisherYatesShuffle(arr []string, rng *rand.Rand) []string {
	n := len(arr)
	for i := n - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// buildQCMOptions composes exactly 4 options from correct_answers + distractors
// Returns the 4 options shuffled, without indicating which are correct
func buildQCMOptions(correctAnswers []string, distractors []string, rng *rand.Rand) []string {
	numCorrect := len(correctAnswers)
	if numCorrect > 4 {
		numCorrect = 4
	}
	needed := 4 - numCorrect

	pool := make([]string, len(distractors))
	copy(pool, distractors)
	fisherYatesShuffle(pool, rng)

	if len(pool) < needed {
		needed = len(pool)
	}

	options := make([]string, 0, 4)
	options = append(options, correctAnswers[:numCorrect]...)
	options = append(options, pool[:needed]...)

	fisherYatesShuffle(options, rng)
	return options
}

// randomizeQuizQuestions takes raw questions from DB and returns them with
// randomized options for QCM types, matching pairs shuffled for column Y,
// and stores session questions in DB
func (h *Handler) randomizeQuizQuestions(questions []map[string]interface{}, sessionID string) []map[string]interface{} {
	seed := int(time.Now().UnixNano())
	rng := rand.New(rand.NewSource(int64(seed)))

	for i, q := range questions {
		qType, _ := q["question_type"].(string)
		qID, _ := q["id"].(string)

		switch qType {
		case "text", "image", "gif", "audio":
			// QCM: build options from answers + distractors
			correctAnswers := []string{}
			answers, _ := q["answers"].([]interface{})
			for _, a := range answers {
				am, ok := a.(map[string]interface{})
				if !ok {
					continue
				}
				if ic, ok := am["is_correct"].(bool); ok && ic {
					if at, ok := am["answer_text"].(string); ok {
						correctAnswers = append(correctAnswers, at)
					}
				}
			}

			distractors := []string{}
			if d, ok := q["distractors"]; ok && d != nil {
				switch dv := d.(type) {
				case []interface{}:
					for _, di := range dv {
						if ds, ok := di.(string); ok {
							distractors = append(distractors, ds)
						}
					}
				case string:
					json.Unmarshal([]byte(dv), &distractors)
				}
			}

			if len(correctAnswers) > 0 {
				options := buildQCMOptions(correctAnswers, distractors, rng)
				q["presented_options"] = options
				h.storeSessionQuestion(sessionID, qID, seed, options)
			}

		case "matching":
			// Parse options - can be map (proper JSONB) or string (legacy)
			var opts map[string]interface{}
			switch v := q["options"].(type) {
			case map[string]interface{}:
				opts = v
			case string:
				json.Unmarshal([]byte(v), &opts)
			}
			if opts != nil {
				pairsRaw, _ := json.Marshal(opts["pairs"])
				var pairs []map[string]interface{}
				json.Unmarshal(pairsRaw, &pairs)

				// Extract Y values, shuffle them
				yValues := make([]string, len(pairs))
				for j, p := range pairs {
					if y, ok := p["y"].(string); ok {
						yValues[j] = y
					}
				}
				fisherYatesShuffle(yValues, rng)

				q["presented_y"] = yValues
				h.storeSessionQuestion(sessionID, qID, seed, yValues)
			}

		case "fill_in":
			// No randomization needed, just pass options through
			switch v := q["options"].(type) {
			case string:
				h.storeSessionQuestion(sessionID, qID, seed, v)
			default:
				b, _ := json.Marshal(v)
				h.storeSessionQuestion(sessionID, qID, seed, string(b))
			}
		}

		questions[i] = q
	}

	return questions
}

// storeSessionQuestion saves presented options for a session question
func (h *Handler) storeSessionQuestion(sessionID, questionID string, seed int, options interface{}) {
	data, _ := json.Marshal(map[string]interface{}{
		"session_id":        sessionID,
		"question_id":       questionID,
		"session_seed":      seed,
		"presented_options": options,
	})
	h.db.Insert("quiz_session_questions", data, true)
}

// validateMatchingAnswer checks a matching answer and returns partial score
// pairs is the original pairs array, matches is [{x_id, y_id}]
// Returns (correctCount, totalPairs)
func validateMatchingAnswer(pairs []map[string]interface{}, matches []map[string]interface{}) (int, int) {
	total := len(pairs)
	if total == 0 {
		return 0, 0
	}

	// Build correct mapping: x_id -> y_id
	correctMap := map[string]string{}
	for _, p := range pairs {
		xID, _ := p["id"].(string)
		y, _ := p["y"].(string)
		correctMap[xID] = y
	}

	correct := 0
	for _, m := range matches {
		xID, _ := m["x_id"].(string)
		yID, _ := m["y_id"].(string)
		if correctMap[xID] == yID {
			correct++
		}
	}

	return correct, total
}

// validateFillInAnswer checks fill_in answers
// blanks is the original blanks array, playerBlanks is [{id, value}]
// Returns (correctCount, totalBlanks)
func validateFillInAnswer(blanks []map[string]interface{}, playerBlanks []map[string]interface{}) (int, int) {
	total := len(blanks)
	if total == 0 {
		return 0, 0
	}

	correct := 0
	for _, pb := range playerBlanks {
		pbID, _ := pb["id"].(float64)
		pbValue, _ := pb["value"].(string)

		for _, b := range blanks {
			bID, _ := b["id"].(float64)
			if bID != pbID {
				continue
			}
			bAnswer, _ := b["answer"].(string)
			caseSensitive, _ := b["case_sensitive"].(bool)
			acceptNoAccents, _ := b["accept_without_accents"].(bool)

			playerVal := pbValue
			correctVal := bAnswer

			if !caseSensitive {
				playerVal = toLower(playerVal)
				correctVal = toLower(correctVal)
			}
			if acceptNoAccents {
				playerVal = removeAccents(playerVal)
				correctVal = removeAccents(correctVal)
			}

			if playerVal == correctVal {
				correct++
			}
			break
		}
	}

	return correct, total
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// Simple accent removal for common characters
func removeAccents(s string) string {
	replacer := map[byte]byte{
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'à': 'a', 'â': 'a', 'ä': 'a',
		'ù': 'u', 'û': 'u', 'ü': 'u',
		'ô': 'o', 'ö': 'o',
		'î': 'i', 'ï': 'i',
		'ç': 'c',
		'ñ': 'n',
	}
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if r, ok := replacer[s[i]]; ok {
			result[i] = r
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// For matching: build the pairs map from question options
func getMatchingPairs(q map[string]interface{}) []map[string]interface{} {
	var opts map[string]interface{}
	switch v := q["options"].(type) {
	case map[string]interface{}:
		opts = v
	case string:
		json.Unmarshal([]byte(v), &opts)
	}
	if opts == nil {
		return nil
	}
	pairsRaw, _ := json.Marshal(opts["pairs"])
	var pairs []map[string]interface{}
	json.Unmarshal(pairsRaw, &pairs)
	return pairs
}

// For fill_in: get the blanks array from question options
func getFillInBlanks(q map[string]interface{}) []map[string]interface{} {
	var opts map[string]interface{}
	switch v := q["options"].(type) {
	case map[string]interface{}:
		opts = v
	case string:
		json.Unmarshal([]byte(v), &opts)
	}
	if opts == nil {
		return nil
	}
	blanksRaw, _ := json.Marshal(opts["blanks"])
	var blanks []map[string]interface{}
	json.Unmarshal(blanksRaw, &blanks)
	return blanks
}

// For fill_in: get the template from question options
func getFillInTemplate(q map[string]interface{}) string {
	var opts map[string]interface{}
	switch v := q["options"].(type) {
	case map[string]interface{}:
		opts = v
	case string:
		json.Unmarshal([]byte(v), &opts)
	}
	if opts == nil {
		return ""
	}
	t, _ := opts["template"].(string)
	return t
}
