package search

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type QueryValidator interface {
	ValidateQuery(query string) error
	SanitizeQuery(query string) string
	PreprocessQuery(query string) (*PreprocessedQuery, error)
}

type PreprocessedQuery struct {
	Original    string   `json:"original"`
	Sanitized   string   `json:"sanitized"`
	Tokens      []string `json:"tokens"`
	HasYear     bool     `json:"has_year"`
	HasMake     bool     `json:"has_make"`
	HasModel    bool     `json:"has_model"`
	IsValid     bool     `json:"is_valid"`
	Suggestions []string `json:"suggestions,omitempty"`
}

type queryValidator struct {
	minLength    int
	maxLength    int
	allowedChars *regexp.Regexp
	knownMakes   map[string]bool
	knownModels  map[string]bool
}

func NewQueryValidator() QueryValidator {
	return &queryValidator{
		minLength:    1,
		maxLength:    100,
		allowedChars: regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9\s\-\.]+$`),
		knownMakes:   initKnownMakes(),
		knownModels:  initKnownModels(),
	}
}

func (qv *queryValidator) ValidateQuery(query string) error {
	if len(strings.TrimSpace(query)) < qv.minLength {
		return fmt.Errorf("query too short: minimum %d characters required", qv.minLength)
	}
	
	if len(query) > qv.maxLength {
		return fmt.Errorf("query too long: maximum %d characters allowed", qv.maxLength)
	}
	
	if !qv.allowedChars.MatchString(query) {
		return fmt.Errorf("query contains invalid characters: only letters, numbers, spaces, hyphens and dots are allowed")
	}
	
	// Check for suspicious patterns
	if qv.containsSuspiciousPatterns(query) {
		return fmt.Errorf("query contains suspicious patterns")
	}
	
	return nil
}

func (qv *queryValidator) SanitizeQuery(query string) string {
	// Remove leading/trailing whitespace
	sanitized := strings.TrimSpace(query)
	
	// Remove multiple consecutive spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	sanitized = spaceRegex.ReplaceAllString(sanitized, " ")
	
	// Remove special characters except allowed ones
	charRegex := regexp.MustCompile(`[^a-zA-Z\p{Cyrillic}0-9\s\-\.]`)
	sanitized = charRegex.ReplaceAllString(sanitized, "")
	
	// Remove excessive punctuation
	punctRegex := regexp.MustCompile(`[\.]{2,}`)
	sanitized = punctRegex.ReplaceAllString(sanitized, ".")
	
	hyphenRegex := regexp.MustCompile(`[-]{2,}`)
	sanitized = hyphenRegex.ReplaceAllString(sanitized, "-")
	
	// Trim again after cleaning
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

func (qv *queryValidator) PreprocessQuery(query string) (*PreprocessedQuery, error) {
	preprocessed := &PreprocessedQuery{
		Original: query,
	}
	
	// Validate first
	if err := qv.ValidateQuery(query); err != nil {
		preprocessed.IsValid = false
		preprocessed.Suggestions = qv.generateSuggestions(query)
		return preprocessed, err
	}
	
	// Sanitize
	preprocessed.Sanitized = qv.SanitizeQuery(query)
	
	// Tokenize
	preprocessed.Tokens = qv.tokenize(preprocessed.Sanitized)
	
	// Analyze tokens
	preprocessed.HasYear = qv.hasYear(preprocessed.Tokens)
	preprocessed.HasMake = qv.hasMake(preprocessed.Tokens)
	preprocessed.HasModel = qv.hasModel(preprocessed.Tokens)
	
	preprocessed.IsValid = true
	
	// Generate suggestions for improvement if needed
	if !preprocessed.HasMake && !preprocessed.HasModel {
		preprocessed.Suggestions = qv.generateSuggestions(query)
	}
	
	return preprocessed, nil
}

func (qv *queryValidator) containsSuspiciousPatterns(query string) bool {
	suspiciousPatterns := []string{
		`<script`,
		`javascript:`,
		`SELECT.*FROM`,
		`DROP.*TABLE`,
		`INSERT.*INTO`,
		`UPDATE.*SET`,
		`DELETE.*FROM`,
	}
	
	queryLower := strings.ToLower(query)
	
	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, queryLower); matched {
			return true
		}
	}
	
	return false
}

func (qv *queryValidator) tokenize(query string) []string {
	// Split by spaces and clean each token
	rawTokens := strings.Fields(query)
	tokens := make([]string, 0, len(rawTokens))
	
	for _, token := range rawTokens {
		cleaned := strings.TrimSpace(token)
		if len(cleaned) > 0 {
			tokens = append(tokens, cleaned)
		}
	}
	
	return tokens
}

func (qv *queryValidator) hasYear(tokens []string) bool {
	yearRegex := regexp.MustCompile(`^(19|20)\d{2}$`)
	
	for _, token := range tokens {
		if yearRegex.MatchString(token) {
			return true
		}
	}
	
	return false
}

func (qv *queryValidator) hasMake(tokens []string) bool {
	for _, token := range tokens {
		tokenLower := strings.ToLower(token)
		if qv.knownMakes[tokenLower] {
			return true
		}
		
		// Check for common abbreviations
		if qv.isKnownMakeAbbreviation(tokenLower) {
			return true
		}
	}
	
	return false
}

func (qv *queryValidator) hasModel(tokens []string) bool {
	for _, token := range tokens {
		tokenLower := strings.ToLower(token)
		if qv.knownModels[tokenLower] {
			return true
		}
		
		// Check for model patterns (like X5, A4, etc.)
		if qv.isModelPattern(token) {
			return true
		}
	}
	
	return false
}

func (qv *queryValidator) isKnownMakeAbbreviation(token string) bool {
	abbreviations := map[string]bool{
		"vw":   true,
		"bmw":  true,
		"mb":   true,
		"benz": true,
	}
	
	return abbreviations[token]
}

func (qv *queryValidator) isModelPattern(token string) bool {
	// BMW X-series pattern
	if matched, _ := regexp.MatchString(`^[Xx][1-7]$`, token); matched {
		return true
	}
	
	// Audi A/Q series pattern
	if matched, _ := regexp.MatchString(`^[AaQq][1-8]$`, token); matched {
		return true
	}
	
	// BMW 3-digit series pattern
	if matched, _ := regexp.MatchString(`^\d{3}[a-zA-Z]*$`, token); matched {
		return true
	}
	
	return false
}

func (qv *queryValidator) generateSuggestions(query string) []string {
	suggestions := make([]string, 0)
	
	queryLower := strings.ToLower(query)
	
	// Suggest adding make if only model is present
	if qv.hasModel([]string{query}) && !qv.hasMake([]string{query}) {
		popularMakes := []string{"Toyota", "Honda", "Volkswagen", "BMW", "Mercedes"}
		for _, make := range popularMakes {
			suggestions = append(suggestions, fmt.Sprintf("%s %s", make, query))
		}
	}
	
	// Suggest corrections for common typos
	typoCorrections := map[string]string{
		"golf":    "Golf",
		"camry":   "Camry",
		"civic":   "Civic",
		"corolla": "Corolla",
		"focus":   "Focus",
		"bmw":     "BMW",
		"vw":      "Volkswagen",
	}
	
	for typo, correction := range typoCorrections {
		if strings.Contains(queryLower, typo) && typo != correction {
			corrected := strings.ReplaceAll(queryLower, typo, correction)
			suggestions = append(suggestions, corrected)
		}
	}
	
	// Suggest popular searches if query is too vague
	if len(strings.TrimSpace(query)) < 3 {
		popularSearches := []string{
			"Toyota Camry",
			"Honda Civic",
			"Volkswagen Golf",
			"BMW X5",
			"Mercedes C-Class",
		}
		suggestions = append(suggestions, popularSearches...)
	}
	
	return qv.deduplicateSuggestions(suggestions)
}

func (qv *queryValidator) deduplicateSuggestions(suggestions []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	
	for _, suggestion := range suggestions {
		if !seen[suggestion] {
			seen[suggestion] = true
			result = append(result, suggestion)
		}
	}
	
	return result
}

func initKnownMakes() map[string]bool {
	makes := []string{
		"toyota", "honda", "volkswagen", "vw", "bmw", "mercedes", "benz",
		"audi", "ford", "nissan", "hyundai", "kia", "mazda", "subaru",
		"chevrolet", "gmc", "cadillac", "lexus", "infiniti", "acura",
		"volvo", "saab", "peugeot", "renault", "fiat", "alfa", "romeo",
		"jaguar", "land", "rover", "mini", "smart", "tesla", "porsche",
		"тойота", "хонда", "фольксваген", "бмв", "мерседес", "ауди",
		"форд", "ниссан", "хендай", "киа", "мазда", "субару", "вольво",
	}
	
	makeMap := make(map[string]bool)
	for _, make := range makes {
		makeMap[strings.ToLower(make)] = true
	}
	
	return makeMap
}

func initKnownModels() map[string]bool {
	models := []string{
		"golf", "passat", "jetta", "tiguan", "touareg", "polo", "beetle",
		"camry", "corolla", "prius", "rav4", "highlander", "sienna", "tacoma",
		"civic", "accord", "cr-v", "pilot", "fit", "ridgeline", "odyssey",
		"focus", "fiesta", "mondeo", "kuga", "explorer", "f-150", "mustang",
		"sentra", "altima", "maxima", "rogue", "pathfinder", "titan", "leaf",
		"elantra", "sonata", "tucson", "santa", "genesis", "veloster",
		"optima", "sorento", "sportage", "soul", "forte", "cadenza",
		"cx-5", "cx-9", "mazda3", "mazda6", "miata", "mx-5",
		"outback", "forester", "impreza", "legacy", "ascent", "crosstrek",
		"x1", "x2", "x3", "x4", "x5", "x6", "x7", "i3", "i8", "z4",
		"a3", "a4", "a5", "a6", "a7", "a8", "q3", "q5", "q7", "q8", "tt",
		"c-class", "e-class", "s-class", "glc", "gle", "gls", "gla", "glb",
		"гольф", "пассат", "джетта", "камри", "королла", "приус", "сивик",
		"аккорд", "фокус", "фиеста", "альтима", "сентра", "элантра", "соната",
	}
	
	modelMap := make(map[string]bool)
	for _, model := range models {
		modelMap[strings.ToLower(model)] = true
	}
	
	return modelMap
}

// Additional utility functions for query preprocessing

func NormalizeSpacing(s string) string {
	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(spaceRegex.ReplaceAllString(s, " "))
}

func RemoveNonAlphanumeric(s string) string {
	// Keep only letters, numbers, spaces, hyphens, and dots
	regex := regexp.MustCompile(`[^a-zA-Zа-яА-Ё0-9\s\-\.]`)
	return regex.ReplaceAllString(s, "")
}

func IsValidCarQuery(query string) bool {
	// Basic validation for car-related queries
	if len(strings.TrimSpace(query)) < 2 {
		return false
	}
	
	// Check if query contains at least one letter
	hasLetter := false
	for _, r := range query {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	
	return hasLetter
}