// Package seasonrank 计算赛季排行榜与积分衰减。
package seasonrank

import (
	"sort"
)

// Player 是赛季数据。
type Player struct {
	Name     string
	Score    int
	IdleDays int
}

// Entry 是排行榜条目。
type Entry struct {
	Name  string
	Score int
	Rank  int
}

// Rank 按 (分数降序, 名字升序) 全序排序后打名次：
// 同分玩家拿同一个名次，下一个名次跳号（标准竞赛排名 1,1,3,3）。
func Rank(players []Player) []Entry {
	sorted := append([]Player(nil), players...)
	sort.Slice(sorted, func(left, right int) bool {
		if sorted[left].Score != sorted[right].Score {
			return sorted[left].Score > sorted[right].Score
		}
		return sorted[left].Name < sorted[right].Name
	})
	entries := make([]Entry, 0, len(sorted))
	for index, player := range sorted {
		rank := index + 1
		if index > 0 && player.Score == sorted[index-1].Score {
			rank = entries[index-1].Rank
		}
		entries = append(entries, Entry{Name: player.Name, Score: player.Score, Rank: rank})
	}
	return entries
}

// Decay 统计所有超过 idleDays 天未上线的玩家。
func Decay(players []Player, idleDays int) int {
	decayed := 0
	for _, player := range players {
		if player.IdleDays > idleDays {
			decayed++
		}
	}
	return decayed
}

// Page 返回排行榜的第 page 页（从 1 开始），每页 size 条。
// 排行榜本身按 (分数降序, 名字升序) 全序排序，切片不会重复或漏项。
func Page(entries []Entry, page, size int) []Entry {
	if page < 1 || size < 1 {
		return nil
	}
	start := (page - 1) * size
	if start >= len(entries) {
		return nil
	}
	end := start + size
	if end > len(entries) {
		end = len(entries)
	}
	return entries[start:end]
}

// Ties 样本：两组同分。
func Ties() []Player {
	return []Player{
		{Name: "a", Score: 100}, {Name: "b", Score: 100},
		{Name: "c", Score: 90}, {Name: "d", Score: 90},
	}
}

// Roster 样本：20 个玩家，其中 12 个超过 10 天没上线。
func Roster() []Player {
	players := []Player{}
	for index := 0; index < 20; index++ {
		idle := 2
		if index%5 != 0 && index%3 != 0 {
			idle = 15
		}
		players = append(players, Player{Name: string(rune('a' + index)), Score: 1000 - index, IdleDays: idle})
	}
	return players
}
