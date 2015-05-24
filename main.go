package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

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
	wordMasksPerWeight map[WordMaskWeight][]WordMask
}

type Opts struct {
	filename string
	whitelist string
}

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
	words, err := ReadWordSetFromFile(opts.filename, opts.whitelist)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wordSet := makeWordSet(words)

	fmt.Printf("             Reading file: %s\n", opts.filename)
	fmt.Printf("         Qualifying chars: %s\n", opts.whitelist)
	fmt.Printf("             Unique words: %d\n", len(words))
	fmt.Printf("Words unique set of chars: %d\n", len(wordSet.wordsByMasks))

	// calculate plz
	topWeight, topMasks := wordSet.findTopWeightAndPairs()
	topWords := wordSet.findWordPairsByMasks(topMasks)

	fmt.Println()
	fmt.Printf("          Top pair weight: %d\n", topWeight)
	fmt.Printf("                Top pairs:\n")
	for _, pair := range topWords {
		fmt.Printf("%s %s\n", string(pair.a), string(pair.b))
	}
}

func (ws WordSet) findWordPairsByMasks(maskPairs []WordMaskPair) []WordPair {
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

func (ws WordSet) findTopWeightAndPairs() (WordMaskWeight, []WordMaskPair) {
	topWeight := WordMaskWeight(0)
	var topPairs []WordMaskPair
	checkedMasks := make(map[WordMask]struct{}, len(ws.wordsByMasks))
	for iWeight, iMasks := range ws.wordMasksPerWeight {
		for _, iMask := range iMasks {
			checkedMasks[iMask] = struct{}{}
			for jWeight, jMasks := range ws.wordMasksPerWeight {
				if iWeight + jWeight < topWeight {
					continue // this pair cannot win
				}
				for _, jMask := range jMasks {
					if _, ok := checkedMasks[jMask]; ok {
						continue // this mask was already iMask
					}
					pairMask := iMask | jMask
					pairWeight := popcount(pairMask)
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

func makeWordSet(words []Word) WordSet {
	runeMaskMap := make(map[rune]WordMask)
	for i, r := range wordSetRunes(words) {
		runeMaskMap[r] = 1 << WordMask(i)
	}

	wordsByMasks := make(map[WordMask][]Word)
	for _, word := range words {
		mask := word.makeWordMask(runeMaskMap)
		if _, ok := wordsByMasks[mask]; !ok {
			wordsByMasks[mask] = make([]Word, 0)
		}
		wordsByMasks[mask] = append(wordsByMasks[mask], word)
	}

	wordMasksPerWeight := make(map[WordMaskWeight][]WordMask, 0)
	for mask := range wordsByMasks {
		weight := popcount(mask)
		if _, ok := wordMasksPerWeight[weight]; !ok {
			wordMasksPerWeight[weight] = make([]WordMask, 0)
		}
		wordMasksPerWeight[weight] = append(wordMasksPerWeight[weight], mask)
	}

	return WordSet{
		words: words,
		wordsByMasks: wordsByMasks,
		wordMasksPerWeight: wordMasksPerWeight,
	}
}

func (word Word) makeWordMask(runeMaskMap map[rune]WordMask) WordMask {
	mask := WordMask(0)
	for _, rune := range word {
		mask = mask | runeMaskMap[rune]
	}
	return mask
}

func wordSetRunes(words []Word) Word {
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

func ReadWordSetFromFile(path, whitelist string) ([]Word, error) {
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

// http://en.wikipedia.org/wiki/Hamming_weight
func popcount(wm WordMask) WordMaskWeight {
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
