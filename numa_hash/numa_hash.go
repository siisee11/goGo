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

func TimeTrack(start time.Time, ch chan time.Duration) {
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
	defer wg.Done()
	defer TimeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	for {
		h.Put(<-key, <-value)
	}
}

func getter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, ch chan time.Duration) {
	defer wg.Done()
	defer TimeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("getter: %d, CPU: %d\n", id, C.sched_getcpu())

	for i := 0; i <= 30000; i++ {
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
	putter_pts, getter_pts := testNTime(10)

	err = plotutil.AddLinePoints(p,
		"putter", putter_pts,
		"getter", getter_pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

// run test N times and returns x, y points.
func testNTime(n int) (plotter.XYs, plotter.XYs) {
	getter_pts := make(plotter.XYs, n)
	putter_pts := make(plotter.XYs, n)
	var wg sync.WaitGroup

	/* Global state */
	runtime.GOMAXPROCS(40)
	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())
	lock := true
	h := hashtable.CreateHashTable()

	/* channel */
	getter_time := make(chan time.Duration)
	putter_time := make(chan time.Duration)
	key := make(chan int)
	value := make(chan []byte)

	for i := range getter_pts {

		/* create key and values */
		rand.Seed(time.Now().UnixNano())
		rand_value := randSeq(4096)
		rand_key := rand.Intn(1000000)

		/* putter start */
		wg.Add(1)
		go putter(&wg, 0, h, lock, putter_time)
		putter_elapsed := <-putter_time
		putter_pts[i].Y = float64(putter_elapsed) / float64(time.Second)
		putter_pts[i].X = float64(i)
		wg.Wait()

		/* getter start */
		wg.Add(1)
		go getter(&wg, 1, h, lock, getter_time)
		getter_elapsed := <-getter_time
		getter_pts[i].Y = float64(getter_elapsed) / float64(time.Second)
		getter_pts[i].X = float64(i)
		wg.Wait()
	}
	return putter_pts, getter_pts
}
