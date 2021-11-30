package main

import (
	"context"
	"crypto/sha512"
	"flag"
	"fmt"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func hash(text string) string {
	s := text
	h1 := sha512.Sum512([]byte(s)) // 문자열의 SHA512 해시 값 추출
	return fmt.Sprintf("%x", h1)
}

// pci 호출하는 부분
func work(bar *mpb.Bar, total int, number int, cancel context.CancelFunc, password string, hashPassword string, endFlag *bool) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	max := 100 * time.Millisecond
	// check FPGA card
	fmt.Printf("FPGA #%d checking...\n", number)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("FPGA #%d ready!!!\n", number)
	bar.SetCurrent(int64(number * total))
	for i := 0; i < total; i++ {
		// 중간에 password 찾게 되면 cancel
		numPassword := password + strconv.Itoa((number*total)+i)

		// hash function 실행. 다른 작업을 하고 싶으면 function 정의하고 이 부분을 수정
		resultPassword := hash(numPassword)

		// hash512 결과값과 입력된 hash의 값 비교 (대소문자 구분없이 비교) 해쉬 값을 찾았을 경우 출력후 다른 process들도 종료 시킨다.
		if strings.EqualFold(resultPassword, hashPassword) {
			cancel()
			// endFlag를 true 해줘야 다른 multi process bar들도 멈추고 종료
			*endFlag = true
			time.Sleep(500 * time.Millisecond)
			fmt.Println("======================================================== [ find password ] ========================================================")
			fmt.Println(" [ Find FPGA ] : FPGA #", number)
			fmt.Println(" [ Result ] : ", numPassword)
			fmt.Println(" [ Hash ] : ", hashPassword)
		}

		// 만약 FPGA 보드 중 한개가 결과를 찾았다면 다른 보드들의 작업을 중지
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

//func CountersNoUnit(pairFmt string, wcc ...WC) Decorator {
//	return decor.Counters(0, pairFmt, wcc...)
//}

func consoleRun(maxEnumerate int, numFPGA int, text string, hashText string) {
	var wg sync.WaitGroup
	//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	var endFlag = false
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := mpb.NewWithContext(ctx, mpb.WithWaitGroup(&wg))
	total, numBars := maxEnumerate, numFPGA
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("FPGA#%d:", i)
		barTotal := total*(i+1) - 1
		//decorator1 := decor.CountersNoUnit("%d / %d", decor.WCSyncWidth)
		//decorator1 = decorator1.Decor()
		bar := p.AddBar(int64(barTotal),
			//bar := p.AddBar(int64(((total+1)*(i+1))-1),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(name),
				decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
				// decor.Counters(i, "%d / %d", decor.WCSyncWidth),
				// decor.DSyncWidth bit enables column width synchronization
				// decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					// ETA decorator with ewma age of 60
					decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "Done",
				),
			),
		)
		//bar.SetPriority(i * maxEnumerate)
		//bar.SetTotal(int64(((i + 1) * maxEnumerate) - 1), false)
		// simulating some work
		i := i
		go func() {
			defer wg.Done()
			work(bar, total, i, cancel, text, hashText, &endFlag)
		}()
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()

	if endFlag == true {
		//fmt.Println("context canceled")
	} else {
		fmt.Println("------------------------------------ [ Decrypt Failed ] ------------------------------------")
	}
}

func main() {
	// set variables
	maxFPGA := 4

	// Scan input arguments
	text := flag.String("password", "", "암호 문자열")
	hashText := flag.String("hash", "", "해쉬 문자열")
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
	consoleRun(*maxEnumerate, *numFPGA, *text, *hashText)
}
