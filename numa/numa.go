package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/siisee11/goGo/hashtable"
)

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

func putter(wg *sync.WaitGroup, id int, lock bool) {
	defer wg.Done()
	defer TimeTrack(time.Now())

	if lock {
		setAffinity(id)
	}

	fmt.Printf("putter: %d, CPU: %d\n", id, C.sched_getcpu())

	h := hashtable.CreateHashTable()

	for i := 0; i <= 50000; i++ {
		rand.Seed(time.Now().UnixNano())
		rand_letters := randSeq(4096)
		h.Put(rand.Intn(1000), rand_letters)
	}
	//	h.Display()
}

func getter(wg *sync.WaitGroup, id int, lock bool) {
	defer wg.Done()
	defer TimeTrack(time.Now())

	if lock {
		setAffinity(id)
	}

	fmt.Printf("getter: %d, CPU: %d\n", id, C.sched_getcpu())

	h := hashtable.CreateHashTable()

	for i := 0; i <= 300000; i++ {
		rand.Seed(time.Now().UnixNano())
		h.Get(rand.Intn(1000))
	}
	//	h.Display()
}

func main() {
	var wg sync.WaitGroup

	runtime.GOMAXPROCS(40)
	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())
	lock := true
	wg.Add(1)
	go putter(&wg, 0, lock)
	wg.Wait()

	wg.Add(1)
	go getter(&wg, 15, lock)
	wg.Wait()
}
