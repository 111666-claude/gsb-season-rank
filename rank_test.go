package seasonrank

import "testing"

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
