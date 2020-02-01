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

func TimeTrack(start time.Time) {
	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	fmt.Println(fmt.Sprintf("%s took %s", name, elapsed))
}

func setAffinity(cpuID int) {
	runtime.LockOSThread()
	C.lock_thread(C.int(cpuID))
}

func putter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool) {
	defer wg.Done()
	defer TimeTrack(time.Now())

	if lock {
		setAffinity(id)
	}

	fmt.Printf("putter: %d, CPU: %d\n", id, C.sched_getcpu())

	for i := 0; i <= 25000; i++ {
		rand.Seed(time.Now().UnixNano())
		rand_letters := randSeq(4096)
		h.Put(rand.Intn(1000000), rand_letters)
	}
}

func getter(wg *sync.WaitGroup, id int, h hashtable.HashTable, lock bool) {
	defer wg.Done()
	defer TimeTrack(time.Now())

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
	var wg sync.WaitGroup

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	h := hashtable.CreateHashTable()

	runtime.GOMAXPROCS(40)
	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())
	lock := true
	wg.Add(1)
	go putter(&wg, 0, h, lock)
	wg.Wait()

	wg.Add(1)
	go getter(&wg, 15, h, lock)
	wg.Wait()

	p.Title.Text = "NUMA latency"
	p.X.Label.Text = "iter"
	p.Y.Label.Text = "latency"

	err = plotutil.AddLinePoints(p,
		"First", randomPoints(15),
		"Second", randomPoints(15),
		"Third", randomPoints(15))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

// randomPoints returns some random x, y points.
func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}