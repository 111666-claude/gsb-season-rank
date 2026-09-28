// Package seasonrank 计算赛季排行榜与积分衰减。
// 缺陷：名次用下标 +1（同分不同名次）、衰减只看前 10 名、分页排序不稳定。
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

// Rank 缺陷：名次直接取下标 +1，同分玩家拿到不同名次。
func Rank(players []Player) []Entry {
	sorted := append([]Player(nil), players...)
	sort.SliceStable(sorted, func(left, right int) bool {
		if sorted[left].Score != sorted[right].Score {
			return sorted[left].Score > sorted[right].Score
		}
		return sorted[left].Name < sorted[right].Name
	})
	entries := make([]Entry, 0, len(sorted))
	for index, player := range sorted {
		entries = append(entries, Entry{Name: player.Name, Score: player.Score, Rank: index + 1})
	}
	return entries
}

// Decay 缺陷：只对前 10 名做衰减。
func Decay(players []Player, idleDays int) int {
	decayed := 0
	for index, player := range players {
		if index >= 10 {
			break
		}
		if player.IdleDays > idleDays {
			decayed++
		}
	}
	return decayed
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
