package main

import (
	"courses/internal/core"
	"courses/internal/database"
	"courses/internal/xkcd"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkDiffMethToSearch(b *testing.B) {
	conf, _ := getConfig("../../config.yaml")

	myDB := database.NewDB(conf.DBFile)

	comics, _ := myDB.Read()
	if comics == nil {
		comics = make(map[int]core.ComicsDescript, 3000)
	}
	downloader := xkcd.NewComicsDownloader(conf.SourceUrl)

	comics, _ = fillMissedComics(5, comics, myDB, downloader)

	index := make(map[string][]int)
	var doc []string
	for k, v := range comics {
		doc = slices.Concat(doc, v.Keywords)
		for i, token := range v.Keywords {
			if !slices.Contains(v.Keywords[:i], token) {
				index[token] = append(index[token], k)
			}
		}
	}
	testString := []string{"my favorite comics is about unknown mystery person", "idk what comics to search",
		"cool banana man", "orange box sits under that orange table and takes orange to make orange juice",
		"funny comics about math"}
	for _, str := range testString {
		comicsName := "findByIndex-" + strconv.Itoa(len(str))
		b.Run(comicsName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				findByIndex(index, strings.Split(str, " "))
			}
		})
		comicsName = "findByComics-" + strconv.Itoa(len(str))
		b.Run(comicsName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				findByComics(comics, strings.Split(str, " "))
			}
		})
	}
}
