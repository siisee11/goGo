package main

/*
#define _GNU_SOURCE
#include <sched.h>
#include <pthread.h>

void lock_thread(int cpuid) {
        pthread_t tid;
        cpu_set_t cpuset;

        tid = pthread_self();
        CPU_ZERO(&cpuset);
        CPU_SET(cpuid, &cpuset);
    pthread_setaffinity_np(tid, sizeof(cpu_set_t), &cpuset);
}
*/
import "C"
import (
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/siisee11/goGo/hashtable"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func timeTrack(start time.Time, ch chan time.Duration) {
	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	fmt.Println(fmt.Sprintf("%s took %s", name, elapsed))
	ch <- elapsed
}

func setAffinity(cpuID int) {
	runtime.LockOSThread()
	C.lock_thread(C.int(cpuID))
}

func putter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, key, quit chan int, value chan []byte, ch chan time.Duration) {
	defer fmt.Println("Putter", id, " Done")
	defer wg.Done()
	defer timeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("putter: %d, CPU: %d\n", id, C.sched_getcpu())

	for {
		select {
		case key := <-key:
			h.Put(key, <-value)
		case <-quit:
			return
		}
	}
}

func getter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, ch chan time.Duration) {
	defer wg.Done()
	defer timeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("getter: %d, CPU: %d\n", id, C.sched_getcpu())

	for i := 0; i <= 500000; i++ {
		rand.Seed(time.Now().UnixNano())
		h.Get(rand.Intn(1000000))
	}
	//	h.Display()
}

func main() {

	/* plot creation */
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "NUMA latency"
	p.X.Label.Text = "iter"
	p.Y.Label.Text = "latency"
	p.Add(plotter.NewGrid())

	/* Test n times */
	putter1Pts, putter2Pts, getterPts := numaTestNTime(1)

	err = plotutil.AddLinePoints(p,
		"putter1", putter1Pts,
		"putter2", putter2Pts,
		"getter", getterPts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

// run test N times and returns x, y points.
func numaTestNTime(n int) (plotter.XYs, plotter.XYs, plotter.XYs) {
	getterPts := make(plotter.XYs, n)
	putter1Pts := make(plotter.XYs, n)
	putter2Pts := make(plotter.XYs, n)
	var wg sync.WaitGroup

	/* Global state */
	runtime.GOMAXPROCS(40)
	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())
	lock := true
	h := hashtable.CreateHashTable()

	/* channel */
	getterTime := make(chan time.Duration)
	putterTime1 := make(chan time.Duration)
	putterTime2 := make(chan time.Duration)
	key1 := make(chan int, 10)
	value1 := make(chan []byte, 10)
	quit1 := make(chan int, 1)
	key2 := make(chan int, 10)
	value2 := make(chan []byte, 10)
	quit2 := make(chan int, 1)

	for i := range getterPts {

		rand.Seed(time.Now().UnixNano())

		/* Create putters */
		go putter(&wg, 1, h, lock, key1, quit1, value1, putterTime1)
		wg.Add(1)
		go putter(&wg, 2, h, lock, key2, quit2, value2, putterTime2)
		wg.Add(1)

		for n := 0; n < 50000; n++ {
			/* create key and values */
			randValue := randSeq(4096)
			randKey := rand.Intn(1000000)

			/* odd key goes to putter1 and even key goes to putter2 */
			//			fmt.Printf("[PUT %d]\n", randKey)
			if randKey%2 == 0 {
				key2 <- randKey
				value2 <- randValue
			} else {
				key1 <- randKey
				value1 <- randValue
			}
		}
		quit1 <- 1
		quit2 <- 1
		putter1Elapsed := <-putterTime1
		putter2Elapsed := <-putterTime2
		putter1Pts[i].Y = float64(putter1Elapsed) / float64(time.Second)
		putter1Pts[i].X = float64(i)
		putter2Pts[i].Y = float64(putter2Elapsed) / float64(time.Second)
		putter2Pts[i].X = float64(i)
		wg.Wait()

		/* getter start */
		wg.Add(1)
		go getter(&wg, 1, h, lock, getterTime)
		getterElapsed := <-getterTime
		getterPts[i].Y = float64(getterElapsed) / float64(time.Second)
		getterPts[i].X = float64(i)
		wg.Wait()
	}
	return putter1Pts, putter2Pts, getterPts
}
