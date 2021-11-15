package main

import (
	"context"
	"crypto/sha512"
	"flag"
	"fmt"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"math/rand"
	"sync"
	"time"
)

func hash(text string) string {
	s := text
	h1 := sha512.Sum512([]byte(s)) // 문자열의 SHA512 해시 값 추출
	//fmt.Printf("%x\n", h1)
	return fmt.Sprintf("%x", h1)
}

// pci 호출하는 부분
func work(bar *mpb.Bar, total int, number int, cancel context.CancelFunc) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	max := 100 * time.Millisecond
	// FPGA card
	fmt.Printf("FPGA #%d checking...\n", number)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("FPGA #%d ready!!!\n", number)
	//bar.IncrBy(number * total)
	//bar.SetRefill(int64(number * total))
	for i := 0; i < total; i++ {
		if i > 1200 {
			cancel()
		}
		if bar.Completed() {
			break
		}
		// start variable is solely for EWMA calculation
		start := time.Now()
		bar.Increment()
		time.Sleep(time.Duration(rng.Intn(10)+1) * max / 100)
		// we need to call DecoratorEwmaUpdate to fulfill ewma decorator's contract
		bar.DecoratorEwmaUpdate(time.Since(start))
	}
}

func consoleRun(maxEnumerate int, numFPGA int, text string) {
	var wg sync.WaitGroup
	//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := mpb.NewWithContext(ctx, mpb.WithWaitGroup(&wg))
	total, numBars := maxEnumerate-1, numFPGA
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
		//bar.SetPriority(i * maxEnumerate)
		//bar.SetTotal(int64(((i + 1) * maxEnumerate) - 1), false)
		// simulating some work
		i := i
		go func() {
			defer wg.Done()
			work(bar, total, i, cancel)
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()

	result := hash(text)

	fmt.Println("------------------------------------[ Decrypt Success ]------------------------------------")
	fmt.Println("------------------------------------[ result ]---------------------------------------------")
	fmt.Println(result)
}

func main() {
	// set variables
	maxFPGA := 4

	// Scan input arguments
	text := flag.String("password", "", "암호 문자열")
	numFPGA := flag.Int("numFPGA", 3, "FPGA 분산처리 숫자")
	maxEnumerate := flag.Int("maxTry", 1000, "Number of FPGA Max Decrypt try")
	//success := flag.Bool("success", false, "decrypt success is true, fail is false")

	flag.Parse()

	// If input arguments empty print help
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if *numFPGA > maxFPGA {
		println("Max FPGA Number is", maxFPGA, "please retry.")
		return
	}
	fmt.Println("Input text : ", *text)
	fmt.Println("Number of FPGA : ", *numFPGA)
	fmt.Printf("Input max try value: %d\n", *maxEnumerate)
	consoleRun(*maxEnumerate, *numFPGA, *text)
}
