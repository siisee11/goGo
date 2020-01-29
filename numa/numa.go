package main

import (
	"fmt"
	"log"
	"math/rand"
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

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func setAffinity(cpuID int) {
	runtime.LockOSThread()
	C.lock_thread(C.int(cpuID))
}

func putter(wg *sync.WaitGroup, id int, lock bool) {

	if lock {
		setAffinity(id)
	}

	defer timeTrack(time.Now(), "Putter")
	defer wg.Done()

	fmt.Printf("putter: %d, CPU: %d\n", id, C.sched_getcpu())

	b := make([]byte, 4096)
	rand_letters := randSeq(4096)

	for i := 0; i <= 100000; i++ {
		rand.Seed(time.Now().UnixNano())
		h := hashtable.CreateHashTable()
		h.Put(1, b)
		h.Put(2, rand_letters)
	}
}

func main() {
	var wg sync.WaitGroup

	fmt.Printf("# of CPUs: %d\n", runtime.NumCPU())
	//	lock := len(os.Getenv("LOCK")) > 0
	lock := true
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go putter(&wg, i, lock)
	}

	wg.Wait()
	//	time.Sleep(2 * time.Second)

}
