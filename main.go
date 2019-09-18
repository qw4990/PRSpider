package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
)

func getAllVecExprPRs() []*github.PullRequest {
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

	results := make([]*github.PullRequest, 0, 128)
	for i := 0; i <= resp.LastPage; i++ {
		fmt.Println("Total Page:", resp.LastPage, ", Page Now:", i)
		opt.ListOptions.Page = i
		prs, _, err := client.PullRequests.List(context.Background(), "pingcap", "tidb", opt)
		if err != nil {
			panic(err)
		}
		for _, pr := range prs {
			if strings.Contains(*pr.Title, "implement vectorized evaluation") {
				results = append(results, pr)
			}
		}
		time.Sleep(time.Second)
	}

	return results
}

func main() {
	prs := getAllVecExprPRs()
	for _, pr := range prs {
		fmt.Println(*pr.Title)
	}
}
