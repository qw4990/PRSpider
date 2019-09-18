package main

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

func getAllVecExprPRs(begin, end time.Time) []*github.PullRequest {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "ee7850684b533ffe22095caf0fe0a1bba9c4113a"},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	opt := &github.PullRequestListOptions{
		State: "closed",
		Head:  "implement+vectorized+evaluation+for",
	}
	opt.PerPage = 300
	_, resp, err := client.PullRequests.List(context.Background(), "pingcap", "tidb", opt)
	if err != nil {
		panic(err)
	}

	prNames := make(map[string]struct{})
	results := make([]*github.PullRequest, 0, 128)
	var lock sync.Mutex
	var wg sync.WaitGroup
	con := 5
	for i := 0; i < con; i++ {
		wg.Add(1)
		go func(id int, opt github.PullRequestListOptions) {
			defer wg.Done()
			for i := id; i < resp.LastPage; i += con {
				fmt.Println("Total Page:", resp.LastPage, ", Page Now:", i)
				opt.ListOptions.Page = i
				prs, _, err := client.PullRequests.List(context.Background(), "pingcap", "tidb", &opt)
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

					lock.Lock()
					if _, ok := prNames[*pr.Title]; ok {
						lock.Unlock()
						continue
					}

					results = append(results, pr)
					prNames[*pr.Title] = struct{}{}
					lock.Unlock()
				}
				time.Sleep(time.Duration(rand.Intn(20)+1) * time.Second) // avoid rate limit
			}
		}(i, *opt)
	}

	wg.Wait()
	return results
}

func builtinFuncName(title string) string {
	begin := strings.Index(title, "builtin")
	if begin < 0 {
		panic(title)
	}
	end := begin + 1
	for end < len(title) {
		c := title[end]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			end++
			continue
		}
		break
	}
	return title[begin:end]
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
		user := *pr.User.Login
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
			fmt.Println(builtinFuncName(*pr.Title), *pr.URL)
		}
	}
}
