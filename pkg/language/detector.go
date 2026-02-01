package language

import (
	"strings"
)

// Language represents a detected language.
type Language string

const (
	English  Language = "English"
	German   Language = "German"
	French   Language = "French"
	Spanish  Language = "Spanish"
	Italian  Language = "Italian"
	Dutch    Language = "Dutch"
	Russian  Language = "Russian"
	Japanese Language = "Japanese"
	Chinese  Language = "Chinese"
	Korean   Language = "Korean"
	Unknown  Language = "Unknown"
)

// commonWords maps language to common words for detection.
var commonWords = map[Language][]string{
	English:  {"the", "is", "and", "to", "a", "of", "in", "that", "it", "for", "with", "you", "this", "be"},
	German:   {"der", "die", "und", "in", "den", "von", "zu", "das", "mit", "sich", "des", "auf", "für", "ist"},
	French:   {"le", "de", "un", "et", "à", "être", "en", "que", "pour", "dans", "ce", "il", "qui", "ne"},
	Spanish:  {"de", "la", "que", "el", "en", "y", "a", "los", "se", "del", "las", "un", "por", "con"},
	Italian:  {"il", "di", "da", "un", "è", "per", "e", "la", "che", "a", "in", "con", "si", "lo"},
	Dutch:    {"de", "en", "van", "het", "een", "die", "in", "te", "aan", "op", "dat", "er", "voor", "met"},
	Russian:  {"и", "в", "то", "что", "он", "на", "я", "с", "со", "а", "то", "все", "она", "так"},
	Japanese: {"の", "に", "は", "を", "た", "が", "で", "て", "と", "し", "れ", "さ", "ある", "いる"},
	Chinese:  {"的", "一", "是", "在", "不", "了", "有", "和", "人", "这", "中", "大", "为", "上"},
	Korean:   {"이", "그", "저", "것", "수", "등", "나", "우리", "저희", "따라", "의해", "에", "과", "또"},
}

// Detect detects the language of the given text.
// Returns the detected language and confidence (0-100).
func Detect(text string) (Language, int) {
	if text == "" {
		return Unknown, 0
	}

	text = strings.ToLower(strings.TrimSpace(text))

	// Check for CJK characters (Japanese, Chinese, Korean)
	if containsCJK(text) {
		if containsJapanese(text) {
			return Japanese, 85
		}
		if containsChinese(text) {
			return Chinese, 85
		}
		if containsKorean(text) {
			return Korean, 85
		}
	}

	// Check for Cyrillic (Russian)
	if containsCyrillic(text) {
		return Russian, 80
	}

	// Match common words for other languages
	scores := make(map[Language]int)
	words := strings.Fields(text)

	for lang, commonWordList := range commonWords {
		count := 0
		for _, word := range words {
			// Remove punctuation
			word = strings.TrimFunc(word, func(r rune) bool {
				return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
			})

			for _, common := range commonWordList {
				if word == common {
					count++
					break
				}
			}
		}
		if count > 0 {
			scores[lang] = (count * 100) / len(commonWordList)
		}
	}

	// Find best match
	bestLang := Unknown
	bestScore := 0
	for lang, score := range scores {
		if score > bestScore {
			bestScore = score
			bestLang = lang
		}
	}

	if bestScore == 0 {
		return English, 50 // Default to English if no match
	}

	return bestLang, bestScore
}

// containsCJK checks if text contains CJK characters.
func containsCJK(text string) bool {
	for _, r := range text {
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0xAC00 && r <= 0xD7AF) { // Hangul
			return true
		}
	}
	return false
}

// containsJapanese checks if text contains Japanese hiragana/katakana.
func containsJapanese(text string) bool {
	for _, r := range text {
		if (r >= 0x3040 && r <= 0x309F) || (r >= 0x30A0 && r <= 0x30FF) {
			return true
		}
	}
	return false
}

// containsChinese checks if text contains Chinese characters.
func containsChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// containsKorean checks if text contains Korean Hangul.
func containsKorean(text string) bool {
	for _, r := range text {
		if r >= 0xAC00 && r <= 0xD7AF {
			return true
		}
	}
	return false
}

// containsCyrillic checks if text contains Cyrillic characters.
func containsCyrillic(text string) bool {
	for _, r := range text {
		if (r >= 0x0400 && r <= 0x04FF) || // Cyrillic
			(r >= 0x0500 && r <= 0x052F) { // Cyrillic Supplement
			return true
		}
	}
	return false
}

// String returns the string representation of the language.
func (l Language) String() string {
	return string(l)
}

// IsNonEnglish returns true if the language is not English.
func (l Language) IsNonEnglish() bool {
	return l != English && l != Unknown
}
