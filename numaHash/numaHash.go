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
	"golang.org/x/sys/unix"
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

	var cpuSet unix.CPUSet
	cpuSet.Zero()
	cpuSet.Set(cpuID)
	unix.SchedSetaffinity(0, &cpuSet)
}

func putter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, key chan int, value chan []byte, ch chan time.Duration) {
	defer wg.Done()
	defer timeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("putter: %d, CPU: %d\n", id, C.sched_getcpu())

	for k := range key {
		h.Put(k, <-value)
	}
}

func getter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, key chan int, value chan []byte, ch chan time.Duration) {
	defer wg.Done()
	defer timeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("getter: %d, CPU: %d\n", id, C.sched_getcpu())

	for k := range key {
		h.Get(k)
	}
}

func main() {

	/* plot creation */
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "NUMA HASH latency"
	p.X.Label.Text = "iter"
	p.Y.Label.Text = "latency"
	p.Add(plotter.NewGrid())

	/* Test n times */
	putterPts, getterPts := testNTime(5, false)

	err = plotutil.AddLinePoints(p,
		"putter", putterPts,
		"getter", getterPts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "NumaHash.png"); err != nil {
		panic(err)
	}
}

// run test N times and returns x, y points.
func testNTime(n int, numa bool) (plotter.XYs, plotter.XYs) {
	getterPts := make(plotter.XYs, n)
	putterPts := make(plotter.XYs, n)
	var wg sync.WaitGroup

	/* Cpu set up */
	var cpu1, cpu2 int
	if numa {
		cpu1 = 1
		cpu2 = 11
	} else {
		cpu1 = 1
		cpu2 = 2
	}

	/* Global state */
	runtime.GOMAXPROCS(40)
	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())

	/* channel */
	getterTime1 := make(chan time.Duration)
	getterTime2 := make(chan time.Duration)
	putterTime1 := make(chan time.Duration)
	putterTime2 := make(chan time.Duration)

	for i := range getterPts {
		key1 := make(chan int)
		key2 := make(chan int)
		value1 := make(chan []byte)
		value2 := make(chan []byte)
		h := hashtable.CreateHashTable()

		rand.Seed(time.Now().UnixNano())

		/* Create putters */
		go putter(&wg, cpu1, h, true, key1, value1, putterTime1)
		wg.Add(1)
		go putter(&wg, cpu2, h, true, key2, value2, putterTime2)
		wg.Add(1)

		/* Put n times into hash */
		for n := 0; n < 50000; n++ {
			/* create key and values */
			randValue := randSeq(4096)
			randKey := rand.Intn(1000000)

			/* odd key goes to putter1 and even key goes to putter2 */
			if randKey%2 == 0 {
				key2 <- randKey
				value2 <- randValue
			} else {
				key1 <- randKey
				value1 <- randValue
			}
		}
		/* Close channels */
		close(key1)
		close(key2)
		close(value1)
		close(value2)

		key1 = make(chan int)
		key2 = make(chan int)
		value1 = make(chan []byte)
		value2 = make(chan []byte)

		/* Calculate elapsed time and plot it */
		putter1Elapsed := <-putterTime1
		putter2Elapsed := <-putterTime2
		putterPts[i].Y = ( float64(putter1Elapsed) + float64(putter2Elapsed) ) / float64(time.Second) / 2
		putterPts[i].X = float64(i)
		wg.Wait()

		/* Create getters */
		go getter(&wg, cpu1, h, true, key1, value1, getterTime1)
		wg.Add(1)
		go getter(&wg, cpu2, h, true, key2, value2, getterTime2)
		wg.Add(1)

		/* Get n times from hash */
		for n := 0; n < 5000000; n++ {
			/* create search key */
			randKey := rand.Intn(1000000)

			/* odd key goes to putter1 and even key goes to putter2 */
			if randKey%2 == 0 {
				key2 <- randKey
			} else {
				key1 <- randKey
			}
		}
		/* Close channels */
		close(key1)
		close(key2)
		close(value1)
		close(value2)

		/* Calculate elapsed time and plot it */
		getter1Elapsed := <-getterTime1
		getter2Elapsed := <-getterTime2
		getterPts[i].Y = ( float64(getter1Elapsed) + float64(getter2Elapsed) ) / float64(time.Second) / 2
		getterPts[i].X = float64(i)
		wg.Wait()
	}
	return putterPts, getterPts
}
