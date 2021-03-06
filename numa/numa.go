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

func putter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, ch chan time.Duration) {
	defer wg.Done()
	defer TimeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("putter: %d, CPU: %d\n", id, C.sched_getcpu())

	for i := 0; i <= 3000000; i++ {
		rand.Seed(time.Now().UnixNano())
		rand_letters := randSeq(4096)
		h.Put(rand.Intn(1000000), rand_letters)
	}
}

func getter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool, ch chan time.Duration) {
	defer wg.Done()
	defer TimeTrack(time.Now(), ch)

	if lock {
		setAffinity(id)
	}

	fmt.Printf("getter: %d, CPU: %d\n", id, C.sched_getcpu())

	for i := 0; i <= 3000000; i++ {
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
//	uma_putter_pts, uma_getter_pts := testNTime(10, 0, 1)
//	numa_putter_pts, numa_getter_pts := testNTime(10, 3, 13)

	uma_getter_pts, numa_getter_pts := testNTime(5, 0, 1, 13)
//	_, numa_getter_pts := testNTime(5, 3, 13)

	err = plotutil.AddLinePoints(p,
//		"UMA-putter", uma_putter_pts,
		"UMA-getter", uma_getter_pts,
//		"NUMA-putter", numa_putter_pts,
		"NUMA-getter", numa_getter_pts)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

// run test N times and returns x, y points.
func testNTime(n int, cpu1 int, cpu2 int, cpu3 int) (plotter.XYs, plotter.XYs) {
	numa_getter_pts := make(plotter.XYs, n)
	uma_getter_pts := make(plotter.XYs, n)
	var wg sync.WaitGroup

	/* Global state */
	runtime.GOMAXPROCS(40)
	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())
	lock := true

	/* channel */
	numa_getter_time := make(chan time.Duration)
	uma_getter_time := make(chan time.Duration)
	putter_time := make(chan time.Duration)

	for i := range numa_getter_pts {
		h := hashtable.CreateHashTable()
		/* putter start */
		wg.Add(1)
		go putter(&wg, cpu1, h, lock, putter_time)
		putter_elapsed := <-putter_time
		fmt.Println("unused: ", putter_elapsed)
		wg.Wait()

		/* getter start */
		wg.Add(1)
		go getter(&wg, cpu2, h, lock, uma_getter_time)
		uma_getter_elapsed := <-uma_getter_time
		uma_getter_pts[i].Y = float64(uma_getter_elapsed) / float64(time.Second)
		uma_getter_pts[i].X = float64(i)
		wg.Wait()

		/* numa_getter start */
		wg.Add(1)
		go getter(&wg, cpu3, h, lock, numa_getter_time)
		numa_getter_elapsed := <-numa_getter_time
		numa_getter_pts[i].Y = float64(numa_getter_elapsed) / float64(time.Second)
		numa_getter_pts[i].X = float64(i)
		wg.Wait()
	}
	return  uma_getter_pts, numa_getter_pts
}
