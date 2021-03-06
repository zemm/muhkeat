// Copyright © 2015 Jussi Rajala <zemm@iki.fi>
//
// This work is free. You can redistribute it and/or modify it under the
// terms of the Do What The Fuck You Want To Public License, Version 2,
// as published by Sam Hocevar. See the COPYING file for more details.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Word string

type WordSet map[Word]struct{}

type RuneSet map[rune]struct{}

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

type WordSetMasks struct {
	words WordSet
	wordsByMasks map[WordMask][]Word
	wordMasksByWeight map[WordMaskWeight][]WordMask
}

func main() {
	filename := flag.String("f", "alastalon_salissa.txt", "source file")
	whitelistChars := flag.String("c", "abcdefghijklmnopqrstuvwzyxåäö", "handled characters")
	flag.Parse()

	words, err := readUniqWordsFromFile(*filename, *whitelistChars)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wsm := NewWordSetMasks(words)

	fmt.Printf("                      Input file: %s\n", *filename)
	fmt.Printf("              Characters handled: %s\n", *whitelistChars)
	fmt.Printf(" Unique (case insensitive) words: %d\n", len(wsm.words))
	fmt.Printf("            Unique sets of chars: %d\n", len(wsm.wordsByMasks))

	// calculate plz
	topMasks, topWeight := wsm.topPairsAndWeight()

	fmt.Println()
	fmt.Printf(" Top pairs found (weight %d)\n", topWeight)
	fmt.Printf("-----------------------------\n")
	for _, pair := range wsm.maskPairsToWordPairs(topMasks) {
		fmt.Printf("%s %s\n", pair.a, pair.b)
	}
}

// WordSetMasks
//

// Create a helper structure for calculating the weights
func NewWordSetMasks(words WordSet) *WordSetMasks {
	// create a map with which to create the word masks
	runeMaskMap := make(map[rune]WordMask)
	i := 0
	for rune := range wordListUniqRunes(words) {
		runeMaskMap[rune] = 1 << WordMask(i)
		i = i + 1
	}

	// group words by their masks
	wordsByMasks := make(map[WordMask][]Word)
	for word := range words {
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

	return &WordSetMasks{
		words: words,
		wordsByMasks: wordsByMasks,
		wordMasksByWeight: wordMasksByWeight,
	}
}

// Convert mask-pairs to words-pairs
func (wsm WordSetMasks) maskPairsToWordPairs(maskPairs []WordMaskPair) []WordPair {
	wordPairs := make([]WordPair, 0)
	for _, maskPair := range maskPairs {
		for _, wordA := range wsm.wordsByMasks[maskPair.a] {
			for _, wordB := range wsm.wordsByMasks[maskPair.b] {
				wordPairs = append(wordPairs, WordPair{wordA, wordB})
			}
		}
	}
	return wordPairs
}

// Find mask pairs that have the most weight (most uniq chars)
func (wsm WordSetMasks) topPairsAndWeight() ([]WordMaskPair, WordMaskWeight) {
	topWeight := WordMaskWeight(0)
	var topPairs []WordMaskPair
	checkedMasks := make(map[WordMask]struct{}, len(wsm.wordsByMasks))
	for iWeight, iMasks := range wsm.wordMasksByWeight {
		for _, iMask := range iMasks {
			checkedMasks[iMask] = struct{}{}
			for jWeight, jMasks := range wsm.wordMasksByWeight {
				if iWeight + jWeight < topWeight {
					continue // this weight combination cannot win
				}
				for _, jMask := range jMasks {
					if _, ok := checkedMasks[jMask]; ok {
						continue // mask already checked
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
	return topPairs, topWeight
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
func wordListUniqRunes(words WordSet) RuneSet {
	allRunesMap := RuneSet{}
	for word := range words {
		for _, r := range word {
			allRunesMap[r] = struct{}{}
		}
	}
	return allRunesMap
}

// Read unique words from a file given a set of accepted characters
func readUniqWordsFromFile(path, whitelist string) (WordSet, error) {
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

	wordSet := WordSet{}
	for scanner.Scan() {
		str := scanner.Text()
		if str != "" {
			str = strings.ToLower(str)
			runes := make([]rune, 0, len(str))
			for _, r := range str {
				if _, ok := whitemap[r]; ok {
					runes = append(runes, r)
				}
			}
			wordSet[Word(runes)] = struct{}{}
		}
	}

	return wordSet, scanner.Err()
}
