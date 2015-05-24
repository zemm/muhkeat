package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Word []rune

type WordSet struct {
	words []Word
	runes []rune
}

type WordSetMasks struct {
	wordSet WordSet
	runeMaskMap map[rune]uint64
	wordsByMasks map[uint64][]Word
	wordMasksPerWeight map[uint8][]uint64
}

func main() {
	srcFile := "alastalon_salissa.txt"
	whitelist := "abcdefghijklmnopqrstuvwzyxåäö"
	wordStrings, err := ReadUniqWordStrings(srcFile, whitelist)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wordSet := makeWordSet(wordStrings)
	wordSetMasks := makeWordSetMasks(wordSet)

	fmt.Printf("unique words: %d\n", len(wordStrings))
	fmt.Printf("words with unique set of characters: %d\n", len(wordSetMasks.wordsByMasks))

	// calculate plz
	topWeight, topMasks := findTopWeightAndMasks(wordSetMasks)

	fmt.Printf("top pair weight: %d\n", topWeight)
	fmt.Println("top pairs:")
	for _, masks := range topMasks {
		fmt.Printf(" ")
		for _, word := range wordSetMasks.wordsByMasks[masks[0]] {
			fmt.Printf(" %s", string(word))
		}
		fmt.Printf(" +")
		for _, word := range wordSetMasks.wordsByMasks[masks[1]] {
			fmt.Printf(" %s", string(word))
		}
		fmt.Println()
	}
}

func findTopWeightAndMasks(wordSetMasks WordSetMasks) (uint8, [][]uint64) {
	topWeight := uint8(0)
	var topMasks [][]uint64
	checkedMasks := make(map[uint64]struct{}, len(wordSetMasks.wordsByMasks))
	for iWeight, iMasks := range wordSetMasks.wordMasksPerWeight {
		for _, iMask := range iMasks {
			checkedMasks[iMask] = struct{}{}
			for jWeight, jMasks := range wordSetMasks.wordMasksPerWeight {
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
						topMasks = make([][]uint64, 0)
					}
					if pairWeight == topWeight {
						topMasks = append(topMasks, []uint64{iMask,jMask})
					}
				}
			}
		}
	}
	return topWeight, topMasks
}

func makeWordSetMasks(wordSet WordSet) WordSetMasks {
	runeMaskMap := make(map[rune]uint64)
	for i, r := range wordSet.runes {
		runeMaskMap[r] = 1 << uint64(i)
	}

	wordsByMasks := make(map[uint64][]Word)
	for _, word := range wordSet.words {
		mask := makeWordMask(word, runeMaskMap)
		if _, ok := wordsByMasks[mask]; !ok {
			wordsByMasks[mask] = make([]Word, 0)
		}
		wordsByMasks[mask] = append(wordsByMasks[mask], word)
	}

	wordMasksPerWeight := make(map[uint8][]uint64, 0)
	for mask := range wordsByMasks {
		weight := popcount(mask)
		if _, ok := wordMasksPerWeight[weight]; !ok {
			wordMasksPerWeight[weight] = make([]uint64, 0)
		}
		wordMasksPerWeight[weight] = append(wordMasksPerWeight[weight], mask)
	}

	return WordSetMasks{
		wordSet: wordSet,
		runeMaskMap: runeMaskMap,
		wordsByMasks: wordsByMasks,
		wordMasksPerWeight: wordMasksPerWeight,
	}
}

func makeWordMask(word Word, runeMaskMap map[rune]uint64) uint64 {
	mask := uint64(0)
	for _, rune := range word {
		mask = mask | runeMaskMap[rune]
	}
	return mask
}

func makeWordSet(strWords []string) WordSet {
	words := make([]Word, 0, len(strWords))
	allRunesMap := make(map[rune]struct{})
	for _, strWord := range strWords {
		word := Word(strWord)
		words = append(words, word)
		for _, r := range word {
			allRunesMap[r] = struct{}{}
		}
	}
	runes := make([]rune, 0, len(allRunesMap))
	for r := range allRunesMap {
		runes = append(runes, r)
	}
	return WordSet{
		words: words,
		runes: runes,
	}
}

func ReadUniqWordStrings(path, whitelist string) ([]string, error) {
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

	uniqWords := make([]string, 0, len(wordMap))
	for word := range wordMap {
		uniqWords = append(uniqWords, word)
	}

	return uniqWords, scanner.Err()
}

// http://en.wikipedia.org/wiki/Hamming_weight
func popcount(x uint64) uint8 {
	mask1 := uint64(6148914691236517205) // 01010101...
	mask2 := uint64(3689348814741910323) // 00110011...
	mask4 := uint64(1085102592571150095) // 00001111...
	x = x - ((x >> 1) & mask1)
	x = (x & mask2) + ((x >> 2) & mask2)
	x = (x + (x >> 4)) & mask4
	x = x + (x >> 8)
	x = x + (x >> 16)
	x = x + (x >> 32)
	return uint8(x)
}
