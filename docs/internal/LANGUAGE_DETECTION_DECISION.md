# Language Detection Decision Log

**Date:** February 1, 2026  
**Status:** Decision Made - Not Implemented  
**Reason:** Project Focus on English Language Only

## Summary

An automatic language detection feature was analyzed and partially implemented to enable the bot to respond in the language the user addressed it in (e.g., German question → German response). However, based on project requirements, this feature was **not finalized** as the bot will remain **English-only** for all responses.

## What Was Attempted

### 1. Language Detection Package (`pkg/language/detector.go`)
- Created a lightweight language detector using keyword and character pattern matching
- Supported 20+ languages including:
  - German, French, Spanish, Portuguese, Italian, Dutch, Polish
  - Russian, Ukrainian, Czech, Swedish, Norwegian, Danish, Finnish
  - Greek, Turkish, Japanese, Chinese, Korean, and more

### 2. Implementation Approach
- **Character Pattern Detection:** Identified language-specific accented characters
  - German: `ä`, `ö`, `ü`, `ß`
  - French: `ç`
  - Spanish: `ñ`
  - Portuguese: `ã`, `õ`
  - Polish/Czech: `ł`, `ą`, `ć`, `č`, `š`, `ž`, etc.
  - Russian/Ukrainian: Cyrillic characters
  - CJK Languages: Unicode ranges for Japanese, Chinese, Korean

- **Keyword Matching:** Used common words in each language for disambiguation
  - Weighted keyword scores to find best language match
  - Fallback to English if no strong match found
  - Threshold-based detection (30% of words must match)

### 3. LLM Service Enhancement
- Updated `GenerateResponse()` to accept optional language parameter
- Added `buildSystemPrompt()` method to inject language instructions
- System prompt modification: "Respond in {language}" appended when non-English language detected

### 4. Bot Command Integration
- Modified `/ask` command handler to detect user's language
- Language detection logged for debugging and analytics

### 5. Testing
- Created comprehensive test suite (`detector_test.go`)
- Achieved 18/20 test cases passing
- Remaining failures were edge cases with ambiguous Romance language detection

## Why Not Implemented

1. **Project Scope:** Living Lands Discord Bot is designed as an **English-only** companion for the Living Lands Reloaded Hytale mod.

2. **Complexity vs. Benefit Trade-off:**
   - Language detection is inherently imperfect without ML models
   - Keyword-based approach had ~90% accuracy on test cases
   - Remaining edge cases required more sophisticated approaches (like `lingua-go` library)

3. **Engineering Decision:**
   - Adding a new library dependency (`pemistahl/lingua-go`) would add ~2MB to binary
   - Simple heuristic detector works but requires fine-tuning for production quality
   - ML-based detection (Google Cloud Language API) requires API keys and costs

4. **User Experience:**
   - Most Discord communities default to English
   - Users can switch to English if bot doesn't understand their language
   - Support for 20+ languages adds significant maintenance burden

## Alternative Solutions (Not Implemented)

### 1. `pemistahl/lingua-go` Library
- **Pros:** 98% accuracy, supports 75 languages, trained on real data
- **Cons:** ~2MB binary size increase, external dependency
- **Recommendation:** Use if multi-language support becomes core requirement

### 2. Google Cloud Language API
- **Pros:** Industry-standard, highly accurate
- **Cons:** Requires API key, adds latency, potential cost
- **Recommendation:** Use for enterprise-scale deployments

### 3. Simple Heuristic (Implemented)
- **Pros:** Zero dependencies, fast (<50ms), lightweight
- **Cons:** ~90% accuracy, requires manual keyword tuning
- **Recommendation:** Good for MVP, needs ML upgrade for production

## Code Artifacts (Removed)

The following code was created but removed:
- `pkg/language/detector.go` (274 lines)
- `pkg/language/detector_test.go` (205 lines)
- Modified `internal/services/llm.go` (added language support)
- Modified `internal/bot/commands.go` (added language detection)

## Future Recommendations

If multi-language support is needed in the future:

1. **Phase 1 (Quick):** Use `pemistahl/lingua-go` library for 98% accuracy
2. **Phase 2 (Production):** Integrate Google Cloud Language API with caching
3. **Phase 3 (Enterprise):** Fine-tune custom ML model on Living Lands documentation

## Lessons Learned

- Language detection is a non-trivial NLP problem
- Keyword-based heuristics work well for high-signal languages (with unique characters)
- Romance languages (Spanish, French, Portuguese) share many keywords, causing ambiguity
- Slavic languages (Russian, Ukrainian, Polish) similarly overlap in character sets
- Production-quality detection requires ML models or specialized libraries

## Decision Made

**The bot will remain English-only.** All responses will be in English regardless of the language the user's question is asked in. This simplifies the codebase, reduces dependencies, and aligns with the project's current scope and target audience (English-speaking Hytale community).

---

**Related Documents:**
- `AGENTS.md` - Project development guidelines
- `QUICK_REFERENCE.md` - Quick development reference
- `docs/IMPLEMENTATION_PLAN.md` - Overall implementation plan
