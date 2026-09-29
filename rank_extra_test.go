package seasonrank

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

// 参考实现：按 (分数降序, 名字升序) 排序后，按分数分组打标准竞赛名次。
func referenceRank(players []Player) []Entry {
	sorted := append([]Player(nil), players...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score > sorted[j].Score
		}
		return sorted[i].Name < sorted[j].Name
	})
	entries := make([]Entry, len(sorted))
	for i := range sorted {
		if i > 0 && sorted[i].Score == sorted[i-1].Score {
			entries[i] = Entry{Name: sorted[i].Name, Score: sorted[i].Score, Rank: entries[i-1].Rank}
			continue
		}
		entries[i] = Entry{Name: sorted[i].Name, Score: sorted[i].Score, Rank: i + 1}
	}
	return entries
}

func assertRankInvariants(t *testing.T, players []Player, entries []Entry) {
	t.Helper()
	if len(entries) != len(players) {
		t.Fatalf("名次条数 %d 与玩家数 %d 不一致", len(entries), len(players))
	}
	for i := range entries {
		if i > 0 {
			if entries[i].Rank < entries[i-1].Rank {
				t.Fatalf("名次必须单调不减：第 %d 条 %d < 前一条 %d", i, entries[i].Rank, entries[i-1].Rank)
			}
			if entries[i].Score == entries[i-1].Score && entries[i].Rank != entries[i-1].Rank {
				t.Fatalf("同分玩家名次必须相同：%v 与 %v", entries[i-1], entries[i])
			}
			if entries[i].Score != entries[i-1].Score && entries[i].Score > entries[i-1].Score {
				t.Fatalf("分数必须降序：第 %d 条 %v", i, entries[i])
			}
			if entries[i].Score == entries[i-1].Score && entries[i].Name < entries[i-1].Name {
				t.Fatalf("同分时名字必须升序：%s 在 %s 前", entries[i].Name, entries[i-1].Name)
			}
		}
	}
}

func TestStandardCompetitionRanksForTies(t *testing.T) {
	entries := Rank(Ties())
	got := map[string]int{}
	for _, entry := range entries {
		got[entry.Name] = entry.Rank
	}
	want := map[string]int{"a": 1, "b": 1, "c": 3, "d": 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("并列名次口径应为 1,1,3,3，实际 %v（%v）", got, entries)
	}
	assertRankInvariants(t, Ties(), entries)
}

func TestRankDeterministicAcrossRuns(t *testing.T) {
	players := []Player{
		{Name: "a", Score: 50}, {Name: "b", Score: 50}, {Name: "c", Score: 90},
		{Name: "d", Score: 10}, {Name: "e", Score: 90}, {Name: "f", Score: 50},
	}
	first := Rank(players)
	for run := 0; run < 10; run++ {
		if got := Rank(players); !reflect.DeepEqual(got, first) {
			t.Fatalf("第 %d 次计算结果与首次不一致：%v vs %v", run, got, first)
		}
	}
	// Rank 不能修改入参。
	original := []Player{{Name: "a", Score: 1}, {Name: "b", Score: 2}}
	input := append([]Player(nil), original...)
	Rank(input)
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("Rank 修改了入参：%v", input)
	}
}

func TestDecayCoversAllPlayers(t *testing.T) {
	if count := Decay(Roster(), 10); count != 11 {
		t.Fatalf("roster 口径 decayed=11，实际 %d", count)
	}
	// 不活跃玩家全部排在前 10 名之后时也必须被统计到。
	players := make([]Player, 30)
	for i := range players {
		players[i] = Player{Name: string(rune('a' + i)), Score: 1000 - i, IdleDays: 2}
	}
	for i := 10; i < 30; i++ {
		players[i].IdleDays = 15
	}
	if count := Decay(players, 10); count != 20 {
		t.Fatalf("前 10 名之后的 20 个不活跃玩家也应衰减，实际 %d", count)
	}
	// 边界：刚好等于 idle-days 不衰减，超过才衰减。
	if count := Decay([]Player{{Name: "x", IdleDays: 10}}, 10); count != 0 {
		t.Fatalf("刚好 idle-days 天不该衰减，实际 %d", count)
	}
}

func TestPaginationStableAcrossPages(t *testing.T) {
	players := []Player{
		{Name: "a", Score: 100}, {Name: "b", Score: 100}, {Name: "c", Score: 90},
		{Name: "d", Score: 90}, {Name: "e", Score: 80},
	}
	full := Rank(players)
	var stitched []Entry
	stitched = append(stitched, Page(players, 1, 2)...)
	stitched = append(stitched, Page(players, 2, 2)...)
	stitched = append(stitched, Page(players, 3, 2)...)
	if len(stitched) != len(full) {
		t.Fatalf("page1+page2+page3 条数 %d 与完整榜单 %d 不一致", len(stitched), len(full))
	}
	for i := range full {
		if stitched[i] != full[i] {
			t.Fatalf("分页拼接后第 %d 条 %v 不等于完整榜单 %v", i, stitched[i], full[i])
		}
	}
	// 无重复：名字唯一。
	seen := map[string]bool{}
	for _, entry := range stitched {
		if seen[entry.Name] {
			t.Fatalf("分页出现重复玩家 %s", entry.Name)
		}
		seen[entry.Name] = true
	}
	// 越界页返回空切片而非脏数据。
	if got := Page(players, 99, 2); len(got) != 0 {
		t.Fatalf("超出范围的页应为空，实际 %v", got)
	}
}

func TestRankMatchesReferenceOn200RandomBoards(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for board := 0; board < 200; board++ {
		n := rng.Intn(50) + 1
		players := make([]Player, n)
		for i := range players {
			players[i] = Player{
				Name:  "p" + strconvItoa(i) + "_" + strconvItoa(board),
				Score: rng.Intn(5), // 大量同分，逼出并列与跳号
			}
		}
		rng.Shuffle(n, func(i, j int) { players[i], players[j] = players[j], players[i] })
		got := Rank(players)
		want := referenceRank(players)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("第 %d 份随机榜单与参考实现不一致：\n got=%v\nwant=%v", board, got, want)
		}
		assertRankInvariants(t, players, got)
	}
}

func strconvItoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

func TestRankOneMillionPlayersWithin2Seconds(t *testing.T) {
	if testing.Short() {
		t.Skip("性能用例在 -short 下跳过")
	}
	const n = 1_000_000
	rng := rand.New(rand.NewSource(7))
	players := make([]Player, n)
	for i := range players {
		players[i] = Player{Name: "player" + strconvItoa(i), Score: rng.Intn(1000), IdleDays: rng.Intn(40)}
	}
	rng.Shuffle(n, func(i, j int) { players[i], players[j] = players[j], players[i] })

	start := time.Now()
	entries := Rank(players)
	elapsed := time.Since(start)

	if len(entries) != n {
		t.Fatalf("名次条数 %d 与玩家数 %d 不一致", len(entries), n)
	}
	assertRankInvariants(t, players, entries)
	if elapsed > 2*time.Second {
		t.Fatalf("100 万玩家排行榜耗时 %v，超过 2 秒门槛", elapsed)
	}
	t.Logf("100 万玩家排行榜耗时 %v", elapsed)
}
