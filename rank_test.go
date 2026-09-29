package seasonrank

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestRankOrder(t *testing.T) {
	entries := Rank([]Player{{Name: "a", Score: 10}, {Name: "b", Score: 30}, {Name: "c", Score: 20}})
	if entries[0].Name != "b" || entries[2].Name != "a" {
		t.Fatalf("排序不对：%v", entries)
	}
}

func TestRankLength(t *testing.T) {
	if len(Rank(Ties())) != 4 {
		t.Fatal("名次条数应该与玩家数一致")
	}
}

func TestRankIsStableForSameName(t *testing.T) {
	first := Rank(Ties())
	second := Rank(Ties())
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("同分排序应该稳定：%v vs %v", first, second)
		}
	}
}

func TestDecayCountsOnlyActiveWindow(t *testing.T) {
	if count := Decay([]Player{{Name: "a", IdleDays: 1}}, 10); count != 0 {
		t.Fatalf("活跃玩家不该被扣分：%d", count)
	}
	if count := Decay([]Player{{Name: "a", IdleDays: 30}}, 10); count != 1 {
		t.Fatalf("不活跃玩家应该被扣分：%d", count)
	}
}

func TestRankTiesShareRank(t *testing.T) {
	entries := Rank(Ties())
	want := []int{1, 1, 3, 3}
	for index, entry := range entries {
		if entry.Rank != want[index] {
			t.Fatalf("并列名次不对：%v", entries)
		}
	}
}

func TestDecayCoversAllPlayers(t *testing.T) {
	if count := Decay(Roster(), 10); count != 11 {
		t.Fatalf("衰减应该覆盖所有不活跃玩家：%d", count)
	}
	players := make([]Player, 0, 30)
	for index := 0; index < 30; index++ {
		players = append(players, Player{Name: fmt.Sprintf("p%02d", index), IdleDays: 30})
	}
	if count := Decay(players, 10); count != 30 {
		t.Fatalf("衰减不该只覆盖前 10 名：%d", count)
	}
}

func TestPageStableNoDupNoMissing(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	players := make([]Player, 0, 25)
	for index := 0; index < 25; index++ {
		players = append(players, Player{Name: fmt.Sprintf("p%02d", index), Score: rng.Intn(10)})
	}
	entries := Rank(players)
	seen := map[string]bool{}
	var paged []Entry
	for page := 1; ; page++ {
		chunk := Page(entries, page, 10)
		if len(chunk) == 0 {
			break
		}
		for _, entry := range chunk {
			if seen[entry.Name] {
				t.Fatalf("翻页出现重复：%s", entry.Name)
			}
			seen[entry.Name] = true
		}
		paged = append(paged, chunk...)
	}
	if len(paged) != len(entries) {
		t.Fatalf("翻页漏项：%d != %d", len(paged), len(entries))
	}
	for index := range entries {
		if paged[index] != entries[index] {
			t.Fatalf("翻页顺序与全序排序不一致：%v vs %v", paged, entries)
		}
	}
}

func TestRankInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	players := make([]Player, 0, 200)
	for index := 0; index < 200; index++ {
		players = append(players, Player{Name: fmt.Sprintf("p%03d", index), Score: rng.Intn(20)})
	}
	first := Rank(players)
	second := Rank(players)
	if len(first) != len(players) {
		t.Fatalf("名次条数应与玩家数一致：%d != %d", len(first), len(players))
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatal("同一份数据重复计算结果应相同")
		}
		if index > 0 {
			if first[index].Rank < first[index-1].Rank {
				t.Fatal("名次应单调不减")
			}
			if first[index].Score == first[index-1].Score && first[index].Rank != first[index-1].Rank {
				t.Fatal("同分玩家名次应相同")
			}
		}
	}
}

// referenceRanks 参考实现：排序后按分数分组打名次。
func referenceRanks(players []Player) []int {
	sorted := append([]Player(nil), players...)
	sort.Slice(sorted, func(left, right int) bool {
		if sorted[left].Score != sorted[right].Score {
			return sorted[left].Score > sorted[right].Score
		}
		return sorted[left].Name < sorted[right].Name
	})
	ranks := make([]int, len(sorted))
	for start := 0; start < len(sorted); {
		end := start
		for end < len(sorted) && sorted[end].Score == sorted[start].Score {
			end++
		}
		for index := start; index < end; index++ {
			ranks[index] = start + 1
		}
		start = end
	}
	return ranks
}

func TestRankMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for trial := 0; trial < 200; trial++ {
		size := 1 + rng.Intn(100)
		players := make([]Player, 0, size)
		for index := 0; index < size; index++ {
			players = append(players, Player{Name: fmt.Sprintf("p%03d", index), Score: rng.Intn(15)})
		}
		entries := Rank(players)
		want := referenceRanks(players)
		diff := 0
		for index := range entries {
			if entries[index].Rank != want[index] {
				diff++
			}
		}
		if diff != 0 {
			t.Fatalf("第 %d 份榜单与参考实现差异 %d 处", trial, diff)
		}
	}
}

func TestRankMillionPlayers(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过百万规模测试")
	}
	rng := rand.New(rand.NewSource(1))
	players := make([]Player, 0, 1000000)
	for index := 0; index < 1000000; index++ {
		players = append(players, Player{Name: fmt.Sprintf("p%07d", index), Score: rng.Intn(100000)})
	}
	start := time.Now()
	entries := Rank(players)
	elapsed := time.Since(start)
	if len(entries) != len(players) {
		t.Fatalf("名次条数应与玩家数一致：%d != %d", len(entries), len(players))
	}
	if elapsed > 2*time.Second {
		t.Fatalf("100 万玩家应在 2 秒内算完：%v", elapsed)
	}
}
