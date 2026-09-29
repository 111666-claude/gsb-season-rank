// Package seasonrank 计算赛季排行榜与积分衰减。
// 名次采用标准竞赛排名（同分同名次，下一个名次跳号），衰减覆盖所有超过
// idle-days 天未上线的玩家，分页在 (分数降序, 名字升序) 的全序上切片。
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

// byScoreDescNameAsc 的排序键是全序：分数降序，分数相同时名字升序。
type byScoreDescNameAsc []Player

func (players byScoreDescNameAsc) Len() int { return len(players) }

func (players byScoreDescNameAsc) Less(left, right int) bool {
	if players[left].Score != players[right].Score {
		return players[left].Score > players[right].Score
	}
	return players[left].Name < players[right].Name
}

func (players byScoreDescNameAsc) Swap(left, right int) {
	players[left], players[right] = players[right], players[left]
}

// sortedPlayers 返回按 (分数降序, 名字升序) 排好序的玩家副本，不修改入参。
func sortedPlayers(players []Player) []Player {
	sorted := append([]Player(nil), players...)
	sort.Stable(byScoreDescNameAsc(sorted))
	return sorted
}

// Rank 按标准竞赛排名返回排行榜：同分玩家同名次，下一个不同分数的名次跳号
// （例如 1,1,3,3）。排序与打名次合计 O(n log n)，不为任何玩家重扫全表。
func Rank(players []Player) []Entry {
	sorted := sortedPlayers(players)
	entries := make([]Entry, len(sorted))
	for index, player := range sorted {
		rank := index + 1
		if index > 0 && player.Score == sorted[index-1].Score {
			rank = entries[index-1].Rank
		}
		entries[index] = Entry{Name: player.Name, Score: player.Score, Rank: rank}
	}
	return entries
}

// Decay 统计所有超过 idleDays 天未上线的玩家数量，与名次、分页无关。
func Decay(players []Player, idleDays int) int {
	decayed := 0
	for _, player := range players {
		if player.IdleDays > idleDays {
			decayed++
		}
	}
	return decayed
}

// Page 在完整排行榜上按 (分数降序, 名字升序) 的全序排序后再切片，
// 返回第 page 页（页号从 1 开始，pageSize 为每页条数）。
// 先全序排序再切片保证各页拼起来与完整榜单一致：无重复、无缺失。
func Page(players []Player, page, pageSize int) []Entry {
	if page < 1 || pageSize <= 0 {
		return nil
	}
	entries := Rank(players)
	start := (page - 1) * pageSize
	if start >= len(entries) {
		return []Entry{}
	}
	end := start + pageSize
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

// Roster 样本：20 个玩家，其中 11 个超过 10 天没上线（含第 10 名之后的玩家）。
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
