package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	rand.Seed(time.Now().Unix())
	// Using WithTimeout here, change 10 to 1 to see the program die before the tasks are done.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tenantsTasks := []string{"johndoe_task", "marcus_task", "hamish_task"}

	result, err := makeAsyncParallelFunctionLookSync(ctx, tenantsTasks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("-----------------------------")
	fmt.Printf("%d Tasks was successful.\n", len(result))
	fmt.Println("-----------------------------")
	for i, r := range result {
		fmt.Printf("%d: %s\n", i, r)
	}

}

// makeAsyncParallelFunctionLookSync, what this function does is to showcase how you can make
// a function that looks like it's synchronous but it's actually doing a lot of things at the same time.
func makeAsyncParallelFunctionLookSync(ctx context.Context, tenantsTasks []string) ([]string, error) {
	eg, ctx := errgroup.WithContext(ctx)
	result := make([]string, 0, len(tenantsTasks))
	mtx := sync.Mutex{} // protects the result slice.

	for _, task := range tenantsTasks {
		task := task // https://golang.org/doc/faq#closures_and_goroutines

		// eg.Go is a super simple package provided by the Go team,
		// it helps working with Go routines when each goroutine can fail.
		// If one of the goroutines errors out it will give the error at the Wait call.
		eg.Go(func() error {
			t, err := taskThatCanError(ctx, task)
			if err != nil {
				return err
			}

			mtx.Lock()
			result = append(result, t)
			mtx.Unlock()

			fmt.Printf("%d/%d Tasks Completed.\n", len(result), cap(result))

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// taskThatCanError represents a long task like calling a database.
// The API when calling a SQL database is something like db.QueryWithContext(ctx, query).
// This just replicate that behavior.
func taskThatCanError(ctx context.Context, task string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	// The tasks takes 0-5 seconds to finish.
	case <-time.After(time.Duration(rand.Intn(5)) * time.Second):
		return fmt.Sprintf("%s_done", task), nil
	}
}
