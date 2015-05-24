package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Word []rune
type WordMask uint64
type WordMaskWeight uint8

type WordSetMasks struct {
	words []Word
	runeMaskMap map[rune]WordMask
	wordsByMasks map[WordMask][]Word
	wordMasksPerWeight map[WordMaskWeight][]WordMask
}

func main() {
	srcFile := "alastalon_salissa.txt"
	whitelist := "abcdefghijklmnopqrstuvwzyxåäö"
	words, err := ReadWordSetFromFile(srcFile, whitelist)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wordSetMasks := makeWordSetMasks(words)

	fmt.Printf("unique words: %d\n", len(words))
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

func findTopWeightAndMasks(wordSetMasks WordSetMasks) (WordMaskWeight, [][]WordMask) {
	topWeight := WordMaskWeight(0)
	var topMasks [][]WordMask
	checkedMasks := make(map[WordMask]struct{}, len(wordSetMasks.wordsByMasks))
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
						topMasks = make([][]WordMask, 0)
					}
					if pairWeight == topWeight {
						topMasks = append(topMasks, []WordMask{iMask,jMask})
					}
				}
			}
		}
	}
	return topWeight, topMasks
}

func makeWordSetMasks(words []Word) WordSetMasks {
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

	return WordSetMasks{
		words: words,
		runeMaskMap: runeMaskMap,
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
