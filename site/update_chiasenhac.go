package site

import (
	"dna"
	"dna/csn"
	"dna/sqlpg"
	"dna/utils"
	"time"
)

// updateMissingIds updates missing ids and return total songids
// (failed or done ids)
func updateMissingIds(db *sqlpg.DB, siteConf *SiteConfig, nLastIds dna.Int) (totalCount dna.Int) {
	missingIDs, err := utils.SelectLastMissingIds("csnsongs", nLastIds, db)
	if err != nil {
		return -1
	} else {
		state := NewStateHandlerWithExtSlice(new(csn.Song), missingIDs, siteConf, db)
		counter := Update(state)
		return counter.Total
	}
}

// updateEmptyTitles returns an error if there is no missing titles
// of songs or videos.
func updateEmptyTitles(db *sqlpg.DB, siteConf *SiteConfig, lastId dna.Int) bool {
	var queryPat dna.String = "select id from %v where id > %v and title = ''"

	songids := &[]dna.Int{}
	songQuery := dna.Sprintf(queryPat, "csnsongs", lastId)
	db.Select(songids, songQuery)

	videoQuery := dna.Sprintf(queryPat, "csnvideos", lastId)
	videoids := &[]dna.Int{}
	db.Select(videoids, videoQuery)

	ids := dna.IntArray(*songids).Concat(dna.IntArray(*videoids))
	if ids.Length() > 0 {
		dna.Log(ids)
		state := NewStateHandlerWithExtSlice(new(csn.SongVideoUpdater), &ids, siteConf, db)
		Update(state)
		RecoverErrorQueries(SqlErrorLogPath, db)
		return false
	} else {
		// dna.Log("No record needs to be updated.")
		return true
	}
	// return donec
}

// UpdateChiasenhac updates new songs, videos and albums.
// New albums are created from new songs fetched.
// After getting new songs or videos, it runs re-fetching
// procedure until there is no missing songs, videos or
// the procesure runs 10 times.
//
// 	fillMode : only update records whose titles are empty.
//
// The update process goes through 6 steps:
// 	Step 1: Initializing db connection, site config, state handler and finding max id
// 	Step 2: Getting new songs from max id found in the table. If videos are found, then it inserts the videos into a video table.
// 	Step 3: Getting missing ids from last 10000 songids.
// 	Step 4: Refetching errors of songs having empty titles.
// 	Step 5: Recovering fail sql statement.
// 	Step 6: Creating new albums from new songs found.
func UpdateChiasenhac(fillMode dna.Bool) {
	// note: songid 1172662 1172663 1172664 are not continuos
	var errCount = 0
	var missingCount = 0
	var lastTotalMissing dna.Int = 0
	var done = false

	TIMEOUT_SECS = 240
	dna.Log("TIMEOUT:", TIMEOUT_SECS)
	// Step 1
	db, err := sqlpg.Connect(sqlpg.NewSQLConfig(SqlConfigPath))
	dna.PanicError(err)
	siteConf, err := LoadSiteConfig("csn", SiteConfigPath)
	dna.PanicError(err)
	// Getting LastSongId for SaveNewAlbums func
	csn.LastSongId, err = utils.GetMaxId("csnsongs", db)
	dna.PanicError(err)
	dna.Log("Max ID:", csn.LastSongId)

	// Step 2: Getting both new songs and videos and
	// inserting into appropriate tables respectively.
	if fillMode == false {
		state := NewStateHandler(new(csn.Song), siteConf, db)
		Update(state)

		// Step 3: Fetching missing ids.
		// It stops when total loop is 3 or the last total ids equals new total ones
		dna.Log(dna.String("\nGetting missing ids from last 10000 songids").ToUpperCase())
		for missingCount < 3 && done == false {
			temp := updateMissingIds(db, siteConf, 10000)
			if temp == lastTotalMissing {
				done = true
			}
			lastTotalMissing = temp
			missingCount += 1
		}

	}

	// SET LAST SONGID
	// csn.LastSongId = 1203603

	// Step 4: Re-fetching err songs
	db.Ping()
	dna.Log(dna.String("\nRe-fetching err from last 10000 songs having EMPTY titles & ID > " + csn.LastSongId.ToString()).ToUpperCase())
	for false == updateEmptyTitles(db, siteConf, csn.LastSongId) && errCount < 10 {
		db.Ping()
		errCount += 1
		dna.Log("RE-FETCHING ROUND:", errCount)
	}
	dna.Log("Re-fetching error done!")

	// Step 5: Recovering failed sql statments
	RecoverErrorQueries(SqlErrorLogPath, db)

	if fillMode == false {
		// Step 6: Saving new abums
		dna.Log("Finding and saving new albums from last songid:", csn.LastSongId)
		nAlbums, err := csn.SaveNewAlbums(db)
		if err != nil {
			dna.Log(err.Error())
		} else {
			dna.Log("New albums inserted:", nAlbums)
		}
	}

	dna.Log("SET TIMEOUT TO DEFAULT: 8s")
	TIMEOUT_SECS = 8
	time.Sleep(2 * time.Second)
	db.Close()
}
