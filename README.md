# season-rank

赛季排行榜：并列名次、积分衰减与稳定分页。只用 Go 标准库。

```
go test ./...
go run ./cmd/season-rank --sample ties
go run ./cmd/season-rank --sample roster --idle-days 10
```

## 口径（README 为准）

- **并列名次按标准竞赛排名**：分数相同的玩家拿同一个名次，下一个名次跳号（1,1,3,3）。
  `--sample ties`（两组同分 a/b=100、c/d=90）按口径是 `a:1,b:1,c:3,d:3`，现在给 `a:1,b:2,c:3,d:4`。
- **衰减覆盖所有不活跃玩家**：所有超过 `--idle-days` 未上线的玩家都要扣分，不只是前 10 名。
  `--sample roster --idle-days 10` 按口径是 `decayed=11`，现在只统计前 10 名，给 `decayed=5`。
- **分页要稳定**：先按 (分数降序, 名字升序) 全序排序再切片，翻页不会重复或漏项；
  口径要求 `page1 + page2` 无重复、无缺失。
- 不变量：名次条数与玩家数一致；名次单调不减；同分玩家名次相同；
  同一份数据重复计算给出同样的排行榜。
- 规模：100 万玩家 2 秒内算完排行榜，内存 O(n log n) 排序所需；不许为每个玩家重扫全表。

## 现在的行为

- `Rank` 直接 `Rank: index + 1`：同分也拿不同名次。
- `Decay` 在 `index >= 10` 处 break：只有前十名会被扣分。
- `Roster` 特意造了 11 个不活跃玩家，用来暴露第二个缺陷。

## 计划新增的能力

```
--sample ties|roster
--idle-days N
```

## 输出

```
ranks=a:1,b:1,c:3,d:3
players=20 decayed=11
```

## 目录

```
rank.go              排名与衰减
cmd/season-rank      命令行入口
rank_test.go         go test 用例
```
