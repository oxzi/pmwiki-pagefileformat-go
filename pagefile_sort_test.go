// SPDX-FileCopyrightText: 2020 Alvar Penning
//
// SPDX-License-Identifier: GPL-3.0-or-later

package pmwiki

import (
	"sort"
	"testing"
	"time"
)

func TestPageFileSortByTime(t *testing.T) {
	pfs := []PageFile{
		{Time: time.Unix(128, 0)},
		{Time: time.Unix(42, 0)},
		{Time: time.Unix(1312, 0)},
		{Time: time.Unix(23, 0)},
		{Time: time.Unix(42, 0)},
		{Time: time.Unix(512, 0)},
		{Time: time.Unix(42, 0)},
	}

	sort.Sort(ByTime(pfs))

	for i := 0; i < len(pfs)-1; i++ {
		if !(pfs[i].Time.Before(pfs[i+1].Time) || pfs[i].Time.Equal(pfs[i+1].Time)) {
			t.Fatalf("assertion %v <= %v failed", pfs[i].Time, pfs[i+1].Time)
		}
	}
}
