package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
)

func getAllVecExprPRs(begin, end time.Time) []*github.PullRequest {
	client := github.NewClient(nil)
	opt := &github.PullRequestListOptions{
		State: "closed",
		Head:  "implement vectorized evaluation",
	}
	opt.PerPage = 300
	_, resp, err := client.PullRequests.List(context.Background(), "pingcap", "tidb", opt)
	if err != nil {
		panic(err)
	}

	resp.LastPage = 5

	results := make([]*github.PullRequest, 0, 128)
	for i := 0; i <= resp.LastPage; i++ {
		fmt.Println("Total Page:", resp.LastPage, ", Page Now:", i)
		opt.ListOptions.Page = i
		prs, _, err := client.PullRequests.List(context.Background(), "pingcap", "tidb", opt)
		if err != nil {
			panic(err)
		}
		for _, pr := range prs {
			if !strings.Contains(*pr.Title, "implement vectorized evaluation") {
				continue
			}

			closedAt := *pr.ClosedAt
			if closedAt.Before(begin) || closedAt.After(end) {
				continue
			}

			results = append(results, pr)
		}
		time.Sleep(time.Second * 5) // avoid rate limit
	}

	return results
}

func main() {
	layout := "2006-01-02 15:04:05"
	str := "2019-09-19 00:00:00"
	end, err := time.Parse(layout, str)
	if err != nil {
		panic(err)
	}
	begin := end.Add(-time.Hour * 24 * 7)
	fmt.Println("Begin:", begin, ", End:", end)

	prs := getAllVecExprPRs(begin, end)
	countMap := make(map[string][]*github.PullRequest)
	for _, pr := range prs {
		user := *pr.User.Name
		countMap[user] = append(countMap[user], pr)
	}

	type SortItem struct {
		Cnt  int
		User string
	}
	items := make([]*SortItem, 0, len(countMap))
	for k, v := range countMap {
		items = append(items, &SortItem{len(v), k})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Cnt > items[j].Cnt
	})

	for _, it := range items {
		fmt.Println(it.User, it.Cnt)
		prs := countMap[it.User]
		for _, pr := range prs {
			fmt.Println(">>", *pr.Title)
		}
	}
}
