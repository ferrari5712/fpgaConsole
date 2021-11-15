package main

import (
	"context"
	"fmt"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"sync"
)

func consoleFail(maxEnumerate int, numFPGA int) {
	var wg sync.WaitGroup

	p := mpb.New(mpb.WithWaitGroup(&wg))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	total, numBars := maxEnumerate, numFPGA
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("FPGA#%d:", i)
		bar := p.AddBar(int64(total),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(name),
				decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
				// decor.DSyncWidth bit enables column width synchronization
				//decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					// ETA decorator with ewma age of 60
					decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "Failed",
				),
			),
		)
		// simulating some work
		i := i
		go func() {
			defer wg.Done()
			work(bar, total, i, cancel)
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()

	fmt.Println("------------------------------------[ Decrypt Failed ]------------------------------------")
	//fmt.Println("------------------------------[ result ]------------------------------")
	//fmt.Println(result)
}
