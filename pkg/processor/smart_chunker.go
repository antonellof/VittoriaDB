package processor

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SmartChunker implements intelligent chunking with sentence boundary detection
// Inspired by Memvid's approach but adapted for VittoriaDB
type SmartChunker struct {
	sentencePattern     *regexp.Regexp
	paragraphPattern    *regexp.Regexp
	abbreviationPattern *regexp.Regexp
	numberPattern       *regexp.Regexp
}

// NewSmartChunker creates a new smart chunker with enhanced boundary detection
func NewSmartChunker() *SmartChunker {
	return &SmartChunker{
		// Enhanced sentence pattern that handles more cases
		sentencePattern: regexp.MustCompile(`[.!?]+\s+`),
		// Paragraph detection
		paragraphPattern: regexp.MustCompile(`\n\s*\n`),
		// Common abbreviations that shouldn't end sentences
		abbreviationPattern: regexp.MustCompile(`(?i)\b(?:dr|mr|mrs|ms|prof|vs|etc|inc|ltd|corp|co|st|ave|blvd|dept|univ|assn|bros|ph\.d|m\.d|b\.a|m\.a|d\.d\.s|j\.d|ll\.b|ll\.m|m\.b\.a|c\.p\.a|r\.n|p\.a|d\.o|d\.v\.m|pharm\.d|ed\.d|psy\.d|m\.s\.w|m\.f\.t|o\.d|au\.d|sc\.d|d\.n\.p|d\.p\.t|o\.t\.r|r\.d|c\.r\.n\.a|f\.n\.p|p\.t|o\.t|s\.l\.p|r\.t|m\.t|r\.r\.t|c\.v\.t|l\.v\.n|c\.n\.a|h\.h\.a|p\.c\.a|e\.m\.t|paramedic|r\.c\.p|r\.p\.t|c\.o\.t\.a|p\.t\.a|o\.t\.a|s\.l\.p\.a|a\.u\.d|ccc-slp|ccc-a|bcba|bcaba|lcsw|lmft|lpc|lmhc|lpcc|lcpc|lcat|lat|ladc|lcadc|cac|cadc|csac|mac|sac|ncac|i|ii|iii|iv|v|vi|vii|viii|ix|x|xi|xii|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec|mon|tue|wed|thu|fri|sat|sun|am|pm|ad|bc|ce|bce|est|pst|mst|cst|edt|pdt|mdt|cdt|gmt|utc|usa|uk|eu|nato|un|fbi|cia|nasa|fda|cdc|who|nato|eu|asap|rsvp|etc|ps|pps|cc|bcc|re|fwd|attn|dept|mgr|dir|pres|vp|ceo|cfo|cto|coo|hr|it|pr|qa|rd|ops|admin|info|sales|support|help|contact|about|home|news|blog|faq|terms|privacy|legal|copyright|trademark|patent|llc|inc|corp|ltd|co|plc|sa|ag|gmbh|kg|og|as|ab|oy|spa|srl|bv|nv|cv|vof|eurl|sarl|sas|sasu|snc|scs|sei|scop|gie|aie|geie|eeig|se|scic|epic|cic|cio|cil|rsl|iom|pcc|spc|llp|lp|gp|jv|upa|ulp|mlp|pllc|pc|pa|sc|psc|chartered|professional|association|partnership|corporation|company|limited|liability|public|private|holding|investment|trust|fund|group|international|global|worldwide|national|regional|local|municipal|federal|state|provincial|county|city|town|village|district|territory|commonwealth|republic|kingdom|empire|union|federation|confederation|alliance|league|organization|organisation|institute|institution|foundation|society|association|club|council|committee|commission|board|panel|tribunal|court|agency|bureau|department|ministry|office|service|authority|administration|government|parliament|congress|senate|house|assembly|legislature|cabinet|executive|judicial|legislative|military|army|navy|air|force|marines|coast|guard|police|fire|emergency|medical|hospital|clinic|center|centre|university|college|school|academy|institute|library|museum|gallery|theater|theatre|cinema|studio|laboratory|lab|factory|plant|mill|mine|farm|ranch|store|shop|market|mall|plaza|square|park|garden|zoo|aquarium|stadium|arena|field|court|track|pool|gym|spa|resort|hotel|motel|inn|lodge|cabin|cottage|apartment|condo|house|home|building|tower|bridge|tunnel|road|street|avenue|boulevard|lane|drive|way|path|trail|highway|freeway|expressway|interstate|route|circle|court|place|terrace|crescent|close|mews|gardens|park|square|green|common|heath|moor|hill|mount|mountain|valley|river|lake|sea|ocean|bay|gulf|strait|channel|island|peninsula|cape|point|head|rock|reef|beach|shore|coast|harbor|harbour|port|dock|pier|wharf|quay|marina|airport|station|terminal|depot|garage|parking|lot|meter|metre|yard|foot|feet|inch|inches|mile|miles|kilometer|kilometre|gram|grams|kilogram|kilograms|pound|pounds|ounce|ounces|ton|tons|tonne|tonnes|gallon|gallons|liter|litre|litres|quart|quarts|pint|pints|cup|cups|tablespoon|tablespoons|teaspoon|teaspoons|fluid|ounce|ounces|milliliter|millilitre|millilitres|celsius|fahrenheit|kelvin|degree|degrees|percent|percentage|dollar|dollars|cent|cents|euro|euros|pound|pounds|yen|yuan|rupee|rupees|peso|pesos|franc|francs|mark|marks|krona|kronor|krone|kroner|ruble|rubles|dinar|dinars|dirham|dirhams|riyal|riyals|shekel|shekels|lira|lire|rand|baht|won|ringgit|singapore|hong|kong|new|zealand|south|africa|united|states|america|canada|mexico|brazil|argentina|chile|colombia|venezuela|peru|ecuador|bolivia|uruguay|paraguay|guyana|suriname|french|guiana|falkland|islands|antarctica|australia|papua|guinea|fiji|samoa|tonga|vanuatu|solomon|marshall|micronesia|palau|nauru|kiribati|tuvalu|cook|niue|tokelau|pitcairn|norfolk|christmas|cocos|keeling|heard|mcdonald|macquarie|ross|dependency|british|antarctic|territory|south|georgia|sandwich|bouvet|prince|edward|marion|crozet|kerguelen|amsterdam|saint|paul|reunion|mauritius|seychelles|comoros|mayotte|madagascar|socotra|maldives|sri|lanka|india|pakistan|bangladesh|nepal|bhutan|myanmar|burma|thailand|laos|cambodia|vietnam|malaysia|brunei|indonesia|philippines|taiwan|china|mongolia|north|korea|japan|russia|kazakhstan|kyrgyzstan|tajikistan|turkmenistan|uzbekistan|afghanistan|iran|iraq|turkey|syria|lebanon|jordan|israel|palestine|saudi|arabia|yemen|oman|emirates|qatar|bahrain|kuwait|georgia|armenia|azerbaijan|cyprus|greece|albania|macedonia|montenegro|bosnia|herzegovina|serbia|croatia|slovenia|slovakia|czech|republic|poland|lithuania|latvia|estonia|finland|sweden|norway|denmark|iceland|ireland|united|kingdom|netherlands|belgium|luxembourg|germany|austria|switzerland|liechtenstein|france|monaco|andorra|spain|portugal|italy|san|marino|vatican|malta|tunisia|algeria|morocco|libya|egypt|sudan|ethiopia|eritrea|djibouti|somalia|kenya|uganda|tanzania|rwanda|burundi|democratic|republic|congo|central|african|cameroon|equatorial|guinea|gabon|sao|tome|principe|nigeria|benin|togo|ghana|burkina|faso|mali|senegal|gambia|guinea|bissau|cape|verde|sierra|leone|liberia|ivory|coast|mauritania|western|sahara|canary|madeira|azores|faroe|shetland|orkney|hebrides|channel|jersey|guernsey|isle|man|anglesey|skye|mull|islay|arran|bute|orkney|shetland|fair|isle|st|kilda|rockall|lundy|scilly|wight|portland|purbeck|thanet|sheppey|canvey|mersea|hayling|portsea|thorney|wallasea|foulness|two|tree|osea|northey|rat|mouse|badger|seal|puffin|gannet|cormorant|guillemot|razorbill|fulmar|petrel|shearwater|skua|gull|tern|plover|turnstone|sanderling|dunlin|knot|curlew|godwit|redshank|greenshank|whimbrel|oystercatcher|avocet|stilt|phalarope|stint|sandpiper|ruff|snipe|woodcock|lapwing|dotterel|golden|grey|ringed|little|kentish|killdeer|mountain|semipalmated|piping|wilson|black|bellied|american|pacific|common|arctic|sandwich|roseate|forster|royal|elegant|caspian|least|black|white|winged|sooty|bridled|brown|noddy|black|skimmer|dovekie|murre|thick|billed|razorbill|atlantic|puffin|horned|rhinoceros|tufted|crested|parakeet|marbled|kittlitz|xantus|craveri|ancient|cassin|whiskered|least|crested|japanese|rhinoceros|horned|tufted|atlantic|horned|parakeet|least|whiskered|crested|ancient|marbled|kittlitz|xantus|craveri|cassin)\.\s*`),
		// Numbers with decimals that shouldn't end sentences
		numberPattern: regexp.MustCompile(`\d+\.\d+`),
	}
}

// ChunkText implements intelligent chunking with multiple strategies
func (sc *SmartChunker) ChunkText(text string, config *ProcessingConfig) ([]DocumentChunk, error) {
	if text == "" {
		return []DocumentChunk{}, nil
	}

	// Clean and normalize text first
	text = sc.cleanText(text)

	// Try different chunking strategies based on text characteristics
	if sc.isParagraphStructured(text) {
		return sc.chunkByParagraphs(text, config)
	}

	// Default to smart sentence-based chunking
	return sc.chunkBySentences(text, config)
}

// chunkBySentences implements Memvid-style smart sentence chunking
func (sc *SmartChunker) chunkBySentences(text string, config *ProcessingConfig) ([]DocumentChunk, error) {
	sentences := sc.splitIntoSentences(text)
	if len(sentences) == 0 {
		return []DocumentChunk{}, nil
	}

	var chunks []DocumentChunk
	var currentChunk strings.Builder
	var currentSize int
	chunkIndex := 0

	for i, sentence := range sentences {
		sentenceLen := len(sentence)

		// Check if adding this sentence would exceed chunk size
		if currentSize > 0 && currentSize+sentenceLen+1 > config.ChunkSize {
			// Finalize current chunk if it meets minimum size
			if currentChunk.Len() >= config.MinChunkSize {
				chunk := sc.createChunk(
					strings.TrimSpace(currentChunk.String()),
					chunkIndex,
					"smart_sentence",
					map[string]string{
						"sentences":    fmt.Sprintf("%d", sc.countSentences(currentChunk.String())),
						"boundary_type": "sentence",
					},
				)
				chunks = append(chunks, chunk)
				chunkIndex++
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			currentSize = 0

			// Add overlap from previous sentences if configured
			if config.ChunkOverlap > 0 && len(chunks) > 0 {
				overlapText := sc.getOverlapText(sentences, i, config.ChunkOverlap)
				if overlapText != "" {
					currentChunk.WriteString(overlapText)
					currentSize = len(overlapText)
				}
			}
		}

		// Add sentence to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
			currentSize++
		}
		currentChunk.WriteString(sentence)
		currentSize += sentenceLen
	}

	// Add final chunk if it has content
	if currentChunk.Len() >= config.MinChunkSize {
		chunk := sc.createChunk(
			strings.TrimSpace(currentChunk.String()),
			chunkIndex,
			"smart_sentence",
			map[string]string{
				"sentences":    fmt.Sprintf("%d", sc.countSentences(currentChunk.String())),
				"boundary_type": "sentence",
				"final_chunk":  "true",
			},
		)
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// chunkByParagraphs chunks text by paragraph boundaries when appropriate
func (sc *SmartChunker) chunkByParagraphs(text string, config *ProcessingConfig) ([]DocumentChunk, error) {
	paragraphs := sc.paragraphPattern.Split(text, -1)
	
	var chunks []DocumentChunk
	var currentChunk strings.Builder
	chunkIndex := 0

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// If this paragraph alone exceeds chunk size, split it by sentences
		if len(paragraph) > config.ChunkSize {
			// Split large paragraph by sentences
			sentenceChunks, err := sc.chunkBySentences(paragraph, config)
			if err != nil {
				return nil, err
			}
			
			// Adjust chunk indices and add to result
			for _, sentenceChunk := range sentenceChunks {
				sentenceChunk.ID = fmt.Sprintf("chunk_%d", chunkIndex)
				sentenceChunk.Position = chunkIndex
				sentenceChunk.Metadata["boundary_type"] = "paragraph_sentence_hybrid"
				chunks = append(chunks, sentenceChunk)
				chunkIndex++
			}
			continue
		}

		// If adding this paragraph would exceed chunk size, finalize current chunk
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph)+2 > config.ChunkSize {
			if currentChunk.Len() >= config.MinChunkSize {
				chunk := sc.createChunk(
					strings.TrimSpace(currentChunk.String()),
					chunkIndex,
					"smart_paragraph",
					map[string]string{
						"paragraphs":   fmt.Sprintf("%d", strings.Count(currentChunk.String(), "\n\n")+1),
						"boundary_type": "paragraph",
					},
				)
				chunks = append(chunks, chunk)
				chunkIndex++
			}

			// Start new chunk
			currentChunk.Reset()
		}

		// Add paragraph to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
	}

	// Add final chunk if it has content
	if currentChunk.Len() >= config.MinChunkSize {
		chunk := sc.createChunk(
			strings.TrimSpace(currentChunk.String()),
			chunkIndex,
			"smart_paragraph",
			map[string]string{
				"paragraphs":   fmt.Sprintf("%d", strings.Count(currentChunk.String(), "\n\n")+1),
				"boundary_type": "paragraph",
				"final_chunk":  "true",
			},
		)
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// splitIntoSentences splits text into sentences with enhanced boundary detection
func (sc *SmartChunker) splitIntoSentences(text string) []string {
	// Simple but effective sentence splitting
	// First, protect abbreviations and decimals
	protectedText := text
	protectedText = sc.abbreviationPattern.ReplaceAllStringFunc(protectedText, func(match string) string {
		return strings.ReplaceAll(match, ".", "<!DOT!>")
	})
	protectedText = sc.numberPattern.ReplaceAllStringFunc(protectedText, func(match string) string {
		return strings.ReplaceAll(match, ".", "<!DECIMAL!>")
	})

	// Split on sentence endings followed by whitespace and capital letter or end of text
	sentenceEndPattern := regexp.MustCompile(`([.!?]+)\s+([A-Z]|$)`)
	
	var sentences []string
	lastEnd := 0
	
	matches := sentenceEndPattern.FindAllStringIndex(protectedText, -1)
	
	for _, match := range matches {
		// Extract sentence from lastEnd to the end of punctuation
		sentenceEnd := match[0] + len(protectedText[match[0]:match[1]]) - len(protectedText[match[1]-1:match[1]])
		if match[1] < len(protectedText) {
			sentenceEnd = match[0] + 1 // Include the punctuation
		} else {
			sentenceEnd = match[1] // End of text
		}
		
		sentence := strings.TrimSpace(protectedText[lastEnd:sentenceEnd])
		if sentence != "" {
			// Restore protected characters
			sentence = strings.ReplaceAll(sentence, "<!DOT!>", ".")
			sentence = strings.ReplaceAll(sentence, "<!DECIMAL!>", ".")
			sentences = append(sentences, sentence)
		}
		
		lastEnd = match[1] - 1 // Start next sentence from the capital letter
	}
	
	// Add remaining text as last sentence if any
	if lastEnd < len(protectedText) {
		remaining := strings.TrimSpace(protectedText[lastEnd:])
		if remaining != "" {
			remaining = strings.ReplaceAll(remaining, "<!DOT!>", ".")
			remaining = strings.ReplaceAll(remaining, "<!DECIMAL!>", ".")
			sentences = append(sentences, remaining)
		}
	}
	
	// Fallback: if no sentences found, return the whole text
	if len(sentences) == 0 && strings.TrimSpace(text) != "" {
		sentences = append(sentences, strings.TrimSpace(text))
	}

	return sentences
}

// handleAbbreviations prevents splitting on common abbreviations
func (sc *SmartChunker) handleAbbreviations(text string) string {
	// Replace abbreviations with temporary markers
	text = sc.abbreviationPattern.ReplaceAllStringFunc(text, func(match string) string {
		return strings.ReplaceAll(match, ".", "<!DOT!>")
	})
	
	// Handle numbers with decimals
	text = sc.numberPattern.ReplaceAllStringFunc(text, func(match string) string {
		return strings.ReplaceAll(match, ".", "<!DECIMAL!>")
	})
	
	return text
}

// restoreAbbreviations restores the original dots in abbreviations
func (sc *SmartChunker) restoreAbbreviations(text string) string {
	text = strings.ReplaceAll(text, "<!DOT!>", ".")
	text = strings.ReplaceAll(text, "<!DECIMAL!>", ".")
	return text
}

// getOverlapText gets overlap text from previous sentences (Memvid-style)
func (sc *SmartChunker) getOverlapText(sentences []string, currentIndex, overlapSize int) string {
	if currentIndex == 0 || overlapSize <= 0 {
		return ""
	}

	var overlap strings.Builder
	overlapChars := 0

	// Go backwards from current sentence to build overlap
	for i := currentIndex - 1; i >= 0 && overlapChars < overlapSize; i-- {
		sentence := sentences[i]
		if overlapChars+len(sentence) <= overlapSize {
			if overlap.Len() > 0 {
				overlap.WriteString(" ")
			}
			overlap.WriteString(sentence)
			overlapChars += len(sentence) + 1
		} else {
			break
		}
	}

	return overlap.String()
}

// isParagraphStructured determines if text has clear paragraph structure
func (sc *SmartChunker) isParagraphStructured(text string) bool {
	paragraphs := sc.paragraphPattern.Split(text, -1)
	
	// Consider it paragraph-structured if:
	// 1. Has multiple paragraphs
	// 2. Average paragraph length is reasonable
	// 3. Not too many very short paragraphs (likely not prose)
	
	if len(paragraphs) < 3 {
		return false
	}

	var totalLength int
	shortParagraphs := 0
	
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		
		totalLength += len(p)
		if len(p) < 50 { // Very short paragraph
			shortParagraphs++
		}
	}

	avgLength := float64(totalLength) / float64(len(paragraphs))
	shortRatio := float64(shortParagraphs) / float64(len(paragraphs))

	// Good paragraph structure: reasonable average length, not too many short paragraphs
	return avgLength > 100 && shortRatio < 0.5
}

// cleanText removes extra whitespace and normalizes text
func (sc *SmartChunker) cleanText(text string) string {
	// Remove extra whitespace but preserve paragraph breaks
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n[ \t]*\n`).ReplaceAllString(text, "\n\n")
	
	// Remove non-printable characters except newlines and tabs
	text = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return -1
	}, text)

	return strings.TrimSpace(text)
}

// countSentences counts the number of sentences in text
func (sc *SmartChunker) countSentences(text string) int {
	sentences := sc.splitIntoSentences(text)
	return len(sentences)
}

// createChunk creates a DocumentChunk with the given parameters
func (sc *SmartChunker) createChunk(content string, position int, chunkType string, metadata map[string]string) DocumentChunk {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	
	metadata["chunk_type"] = chunkType
	metadata["char_count"] = fmt.Sprintf("%d", len(content))
	metadata["word_count"] = fmt.Sprintf("%d", len(strings.Fields(content)))

	return DocumentChunk{
		ID:       fmt.Sprintf("chunk_%d", position),
		Content:  content,
		Position: position,
		Size:     len(content),
		Metadata: metadata,
	}
}
