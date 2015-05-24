package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// Types
//
type Word []rune

type WordPair struct {
	a Word
	b Word
}

type WordMask uint64

type WordMaskPair struct {
	a WordMask
	b WordMask
}

type WordMaskWeight uint8

type WordSet struct {
	words []Word
	wordsByMasks map[WordMask][]Word
	wordMasksByWeight map[WordMaskWeight][]WordMask
}

type Opts struct {
	filename string
	whitelist string
}

// Parse program options
func handleArgs([]string) Opts {
	opts := Opts{
		filename: "",
		whitelist: "abcdefghijklmnopqrstuvwzyxåäö",
	}
	args := make([]string, 0)
	flags := make([]string, 0)
	for _, arg := range os.Args[1:] {
		if arg[0] == '-' {
			flags = append(flags, arg)
		} else {
			args = append(args, arg)
		}
	}
	switch(len(args)) {
		case 1:
			opts.filename = args[0]
		case 2:
			opts.filename = args[0]
			opts.whitelist = args[1]
	}
	if len(flags) > 0 || opts.filename == "" {
		fmt.Printf("Usage: %s filename [whitelistChars]\n", path.Base(os.Args[0]))
		os.Exit(255)
	}
	return opts
}

func main() {
	opts := handleArgs(os.Args)
	words, err := readUniqWordsFromFile(opts.filename, opts.whitelist)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wordSet := NewWordSet(words)

	fmt.Printf("                    Reading file: %s\n", opts.filename)
	fmt.Printf("                Qualifying chars: %s\n", opts.whitelist)
	fmt.Printf("         Unique normalized words: %d\n", len(words))
	fmt.Printf("Words with a unique set of chars: %d\n", len(wordSet.wordsByMasks))

	// calculate plz
	topWeight, topMasks := wordSet.topWeightAndPairs()
	topWords := wordSet.maskPairsToWordPairs(topMasks)

	fmt.Println()
	fmt.Printf("Top pairs found (weight %d):\n", topWeight)
	fmt.Printf("----------------------------\n")
	for _, pair := range topWords {
		fmt.Printf("%s %s\n", string(pair.a), string(pair.b))
	}
}

// WordSet
//

// Create a helper structure for calculating the weights
func NewWordSet(words []Word) *WordSet {
	// create a map with which to create the word masks
	runeMaskMap := make(map[rune]WordMask)
	for i, r := range wordListUniqRunes(words) {
		runeMaskMap[r] = 1 << WordMask(i)
	}

	// group words by their masks
	wordsByMasks := make(map[WordMask][]Word)
	for _, word := range words {
		mask := word.mask(runeMaskMap)
		if _, ok := wordsByMasks[mask]; !ok {
			wordsByMasks[mask] = make([]Word, 0)
		}
		wordsByMasks[mask] = append(wordsByMasks[mask], word)
	}

	// group masks by their weights
	wordMasksByWeight := make(map[WordMaskWeight][]WordMask, 0)
	for mask := range wordsByMasks {
		weight := mask.weight()
		if _, ok := wordMasksByWeight[weight]; !ok {
			wordMasksByWeight[weight] = make([]WordMask, 0)
		}
		wordMasksByWeight[weight] = append(wordMasksByWeight[weight], mask)
	}

	return &WordSet{
		words: words,
		wordsByMasks: wordsByMasks,
		wordMasksByWeight: wordMasksByWeight,
	}
}

func (ws WordSet) maskPairsToWordPairs(maskPairs []WordMaskPair) []WordPair {
	wordPairs := make([]WordPair, 0)
	for _, maskPair := range maskPairs {
		for _, wordA := range ws.wordsByMasks[maskPair.a] {
			for _, wordB := range ws.wordsByMasks[maskPair.b] {
				wordPairs = append(wordPairs, WordPair{wordA, wordB})
			}
		}
	}
	return wordPairs
}

// Find mask pairs that have the most weight (most uniq chars)
func (ws WordSet) topWeightAndPairs() (WordMaskWeight, []WordMaskPair) {
	topWeight := WordMaskWeight(0)
	var topPairs []WordMaskPair
	checkedMasks := make(map[WordMask]struct{}, len(ws.wordsByMasks))
	for iWeight, iMasks := range ws.wordMasksByWeight {
		for _, iMask := range iMasks {
			checkedMasks[iMask] = struct{}{}
			for jWeight, jMasks := range ws.wordMasksByWeight {
				if iWeight + jWeight < topWeight {
					continue // this weight combination cannot win
				}
				for _, jMask := range jMasks {
					if _, ok := checkedMasks[jMask]; ok {
						continue // this mask was already iMask
					}
					pairMask := iMask.union(jMask)
					pairWeight := pairMask.weight()
					if pairWeight > topWeight {
						topWeight = pairWeight
						topPairs = make([]WordMaskPair, 0)
					}
					if pairWeight == topWeight {
						topPairs = append(topPairs, WordMaskPair{iMask,jMask})
					}
				}
			}
		}
	}
	return topWeight, topPairs
}

// Word
//

// Calculate a mask for word given a rune mask map
func (word Word) mask(runeMaskMap map[rune]WordMask) WordMask {
	mask := WordMask(0)
	for _, rune := range word {
		mask = mask | runeMaskMap[rune]
	}
	return mask
}

// WordMask
//

// Create an union wordmask from two masks
func (wm WordMask) union(other WordMask) WordMask {
	return wm | other
}

// Calculate the weight of the word mask
// Weight is the amount of unique characters
// http://en.wikipedia.org/wiki/Hamming_weight
func (wm WordMask) weight() WordMaskWeight {
	mask1 := uint64(6148914691236517205) // 01010101...
	mask2 := uint64(3689348814741910323) // 00110011...
	mask4 := uint64(1085102592571150095) // 00001111...
	x := uint64(wm)
	x = x - ((x >> 1) & mask1)
	x = (x & mask2) + ((x >> 2) & mask2)
	x = (x + (x >> 4)) & mask4
	x = x + (x >> 8)
	x = x + (x >> 16)
	x = x + (x >> 32)
	return WordMaskWeight(x)
}

// Helpers
//

// Find all unique runes from a list of words
func wordListUniqRunes(words []Word) Word {
	allRunesMap := make(map[rune]struct{})
	for _, word := range words {
		for _, r := range word {
			allRunesMap[r] = struct{}{}
		}
	}

	allRunes := make([]rune, 0, len(allRunesMap))
	for r := range allRunesMap {
		allRunes = append(allRunes, r)
	}

	return allRunes
}

// Read unique words from a file given a set of accepted characters
func readUniqWordsFromFile(path, whitelist string) ([]Word, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	whitemap := make(map[rune]struct{})
	for _, r := range whitelist {
		whitemap[r] = struct{}{}
	}

	wordMap := make(map[string]struct{})
	for scanner.Scan() {
		str := scanner.Text()
		if str != "" {
			str = strings.ToLower(str)
			word := make([]rune, 0, len(str))
			for _, r := range str {
				if _, ok := whitemap[r]; ok {
					word = append(word, r)
				}
			}
			wordMap[string(word)] = struct{}{}
		}
	}

	uniqWords := make([]Word, 0, len(wordMap))
	for word := range wordMap {
		uniqWords = append(uniqWords, Word(word))
	}

	return uniqWords, scanner.Err()
}
