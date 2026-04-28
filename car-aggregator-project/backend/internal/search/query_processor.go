package search

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type QueryProcessor interface {
	ParseQuery(query string) (*ParsedQuery, error)
	NormalizeQuery(query string) (*NormalizedQuery, error)
	ExtractMakeModel(query string) (make, model string, confidence float64)
}

type ParsedQuery struct {
	Original     string        `json:"original"`
	Make         string        `json:"make"`
	Model        string        `json:"model"`
	Year         *int          `json:"year,omitempty"`
	Confidence   float64       `json:"confidence"`
	Language     string        `json:"language"` // "ru", "en"
	Alternatives []Alternative `json:"alternatives"`
}

type Alternative struct {
	Make       string  `json:"make"`
	Model      string  `json:"model"`
	Confidence float64 `json:"confidence"`
}

type NormalizedQuery struct {
	Original   string `json:"original"`
	Normalized string `json:"normalized"`
	Language   string `json:"language"`
}

type queryProcessor struct {
	makeModelMap map[string]string // Russian to English mapping
	modelMap     map[string]string // Model name variations
}

func NewQueryProcessor() QueryProcessor {
	return &queryProcessor{
		makeModelMap: initMakeModelMap(),
		modelMap:     initModelMap(),
	}
}

func (qp *queryProcessor) ParseQuery(query string) (*ParsedQuery, error) {
	normalized, err := qp.NormalizeQuery(query)
	if err != nil {
		return nil, err
	}
	
	parsed := &ParsedQuery{
		Original:     query,
		Language:     normalized.Language,
		Alternatives: make([]Alternative, 0),
	}
	
	// Extract year if present
	if year := qp.extractYear(normalized.Normalized); year != nil {
		parsed.Year = year
	}
	
	// Extract make and model
	make, model, confidence := qp.ExtractMakeModel(normalized.Normalized)
	parsed.Make = make
	parsed.Model = model
	parsed.Confidence = confidence
	
	// Generate alternatives
	parsed.Alternatives = qp.generateAlternatives(normalized.Normalized)
	
	return parsed, nil
}

func (qp *queryProcessor) NormalizeQuery(query string) (*NormalizedQuery, error) {
	normalized := &NormalizedQuery{
		Original: query,
	}
	
	// Detect language
	normalized.Language = qp.detectLanguage(query)
	
	// Clean and normalize
	clean := strings.TrimSpace(query)
	clean = qp.removeExtraSpaces(clean)
	clean = qp.normalizeCase(clean)
	
	// Translate Russian to English if needed
	if normalized.Language == "ru" {
		clean = qp.translateRussianTerms(clean)
	}
	
	normalized.Normalized = clean
	return normalized, nil
}

func (qp *queryProcessor) ExtractMakeModel(query string) (make, model string, confidence float64) {
	query = strings.ToLower(strings.TrimSpace(query))
	
	// Remove year if present
	yearRegex := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	query = yearRegex.ReplaceAllString(query, "")
	query = strings.TrimSpace(query)
	
	// Try exact make+model patterns
	if make, model, conf := qp.tryExactPatterns(query); conf > 0.8 {
		return make, model, conf
	}
	
	// Try known make names
	if make, model, conf := qp.tryKnownMakes(query); conf > 0.6 {
		return make, model, conf
	}
	
	// Try model-only search
	if model, conf := qp.tryModelOnly(query); conf > 0.5 {
		return "", model, conf
	}
	
	// Fallback: treat entire query as model
	return "", query, 0.3
}

func (qp *queryProcessor) detectLanguage(query string) string {
	russianChars := 0
	totalChars := 0
	
	for _, r := range query {
		if unicode.IsLetter(r) {
			totalChars++
			if r >= 'а' && r <= 'я' || r >= 'А' && r <= 'Я' {
				russianChars++
			}
		}
	}
	
	if totalChars > 0 && float64(russianChars)/float64(totalChars) > 0.3 {
		return "ru"
	}
	return "en"
}

func (qp *queryProcessor) removeExtraSpaces(s string) string {
	spaceRegex := regexp.MustCompile(`\s+`)
	return spaceRegex.ReplaceAllString(s, " ")
}

func (qp *queryProcessor) normalizeCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func (qp *queryProcessor) translateRussianTerms(query string) string {
	words := strings.Fields(strings.ToLower(query))
	for i, word := range words {
		if translation, exists := qp.makeModelMap[word]; exists {
			words[i] = translation
		}
	}
	return strings.Join(words, " ")
}

func (qp *queryProcessor) extractYear(query string) *int {
	yearRegex := regexp.MustCompile(`\b(19|20)(\d{2})\b`)
	matches := yearRegex.FindStringSubmatch(query)
	if len(matches) >= 3 {
		if year, err := strconv.Atoi(matches[0]); err == nil {
			if year >= 1990 && year <= 2030 {
				return &year
			}
		}
	}
	return nil
}

func (qp *queryProcessor) tryExactPatterns(query string) (string, string, float64) {
	// Pattern: "Make Model"
	patterns := []string{
		`^(volkswagen|vw)\s+(golf|passat|jetta|tiguan|touareg)`,
		`^(toyota)\s+(camry|corolla|prius|rav4|highlander)`,
		`^(honda)\s+(civic|accord|cr-v|pilot|fit)`,
		`^(bmw)\s+(x[1-7]|[1-8]\s*series|\d{3}[a-z]*)`,
		`^(mercedes|benz)\s+(c-class|e-class|s-class|glc|gle)`,
		`^(audi)\s+(a[3-8]|q[3-8]|tt)`,
		`^(ford)\s+(focus|fiesta|mondeo|kuga|explorer)`,
		`^(nissan)\s+(sentra|altima|maxima|rogue|pathfinder)`,
	}
	
	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(query); len(matches) >= 3 {
			make := qp.normalizeMake(matches[1])
			model := qp.normalizeModel(matches[2])
			return make, model, 0.9
		}
	}
	
	return "", "", 0
}

func (qp *queryProcessor) tryKnownMakes(query string) (string, string, float64) {
	knownMakes := []string{
		"volkswagen", "vw", "toyota", "honda", "bmw", "mercedes", "benz",
		"audi", "ford", "nissan", "hyundai", "kia", "mazda", "subaru",
		"chevrolet", "gmc", "cadillac", "lexus", "infiniti", "acura",
	}
	
	words := strings.Fields(strings.ToLower(query))
	
	for i, word := range words {
		for _, make := range knownMakes {
			if word == make || qp.fuzzyMatch(word, make) > 0.8 {
				normalizedMake := qp.normalizeMake(word)
				
				// Get remaining words as model
				var modelWords []string
				for j, w := range words {
					if j != i {
						modelWords = append(modelWords, w)
					}
				}
				
				if len(modelWords) > 0 {
					model := strings.Join(modelWords, " ")
					return normalizedMake, qp.normalizeModel(model), 0.7
				}
			}
		}
	}
	
	return "", "", 0
}

func (qp *queryProcessor) tryModelOnly(query string) (string, float64) {
	popularModels := []string{
		"golf", "camry", "civic", "corolla", "focus", "accord",
		"passat", "jetta", "altima", "sentra", "prius", "rav4",
	}
	
	queryLower := strings.ToLower(query)
	
	for _, model := range popularModels {
		if strings.Contains(queryLower, model) || qp.fuzzyMatch(queryLower, model) > 0.7 {
			return qp.normalizeModel(model), 0.6
		}
	}
	
	return query, 0.4
}

func (qp *queryProcessor) generateAlternatives(query string) []Alternative {
	alternatives := make([]Alternative, 0)
	
	// Try different word combinations
	words := strings.Fields(strings.ToLower(query))
	if len(words) >= 2 {
		// Try reversing make/model order
		reversed := strings.Join([]string{words[1], words[0]}, " ")
		if make, model, conf := qp.ExtractMakeModel(reversed); conf > 0.5 {
			alternatives = append(alternatives, Alternative{
				Make:       make,
				Model:      model,
				Confidence: conf * 0.8, // Lower confidence for alternatives
			})
		}
	}
	
	return alternatives
}

func (qp *queryProcessor) fuzzyMatch(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	// Simple Levenshtein-based similarity
	return qp.levenshteinSimilarity(s1, s2)
}

func (qp *queryProcessor) levenshteinSimilarity(s1, s2 string) float64 {
	if len(s1) == 0 {
		return float64(len(s2))
	}
	if len(s2) == 0 {
		return float64(len(s1))
	}
	
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	distance := matrix[len(s1)][len(s2)]
	maxLen := max(len(s1), len(s2))
	
	return 1.0 - float64(distance)/float64(maxLen)
}

func (qp *queryProcessor) normalizeMake(make string) string {
	makeMap := map[string]string{
		"vw":       "Volkswagen",
		"benz":     "Mercedes-Benz",
		"mercedes": "Mercedes-Benz",
		"bmw":      "BMW",
	}
	
	makeLower := strings.ToLower(make)
	if normalized, exists := makeMap[makeLower]; exists {
		return normalized
	}
	
	return strings.Title(strings.ToLower(make))
}

func (qp *queryProcessor) normalizeModel(model string) string {
	modelMap := map[string]string{
		"golf":    "Golf",
		"camry":   "Camry",
		"civic":   "Civic",
		"corolla": "Corolla",
		"focus":   "Focus",
		"accord":  "Accord",
	}
	
	modelLower := strings.ToLower(model)
	if normalized, exists := modelMap[modelLower]; exists {
		return normalized
	}
	
	return strings.Title(strings.ToLower(model))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func initMakeModelMap() map[string]string {
	return map[string]string{
		"гольф":      "golf",
		"камри":      "camry",
		"сивик":      "civic",
		"королла":    "corolla",
		"фокус":      "focus",
		"аккорд":     "accord",
		"пассат":     "passat",
		"джетта":     "jetta",
		"альтима":    "altima",
		"сентра":     "sentra",
		"фольксваген": "volkswagen",
		"тойота":     "toyota",
		"хонда":      "honda",
		"бмв":        "bmw",
		"мерседес":   "mercedes",
		"ауди":       "audi",
		"форд":       "ford",
		"ниссан":     "nissan",
	}
}

func initModelMap() map[string]string {
	return map[string]string{
		"golf":    "Golf",
		"camry":   "Camry", 
		"civic":   "Civic",
		"corolla": "Corolla",
		"focus":   "Focus",
		"accord":  "Accord",
		"passat":  "Passat",
		"jetta":   "Jetta",
		"altima":  "Altima",
		"sentra":  "Sentra",
	}
}