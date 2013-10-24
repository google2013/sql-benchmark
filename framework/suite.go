package framework

import (
	"database/sql"
	"fmt"
	"runtime"
	"time"
)

type driver struct {
	name string
	db   *sql.DB
}

type Result struct {
	Err      error
	Queries  int
	Duration time.Duration
}

func (res *Result) QueriesPerSecond() float64 {
	return float64(res.Queries) / res.Duration.Seconds()
}

type benchmark struct {
	name string
	n    int
	bm   func(*sql.DB, int) error
}

func (b *benchmark) run(db *sql.DB) Result {
	runtime.GC()

	start := time.Now()
	err := b.bm(db, b.n)
	end := time.Now()

	return Result{
		Err:      err,
		Queries:  b.n,
		Duration: end.Sub(start),
	}
}

type BenchmarkSuite struct {
	drivers    []driver
	benchmarks []benchmark
}

func (bs *BenchmarkSuite) AddDriver(name, drv, dsn string) error {
	db, err := sql.Open(drv, dsn)
	if err != nil {
		return fmt.Errorf("Error registering driver '%s': %s", name, err.Error())
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("Error on driver '%s': %s", name, err.Error())
	}

	bs.drivers = append(bs.drivers, driver{
		name: name,
		db:   db,
	})
	return nil
}

func (bs *BenchmarkSuite) AddBenchmark(name string, n int, bm func(*sql.DB, int) error) {
	bs.benchmarks = append(bs.benchmarks, benchmark{
		name: name,
		n:    n,
		bm:   bm,
	})
}

func (bs *BenchmarkSuite) Run() {
	if len(bs.drivers) < 1 {
		fmt.Println("No drivers registered to run benchmarks with!")
		return
	}

	if len(bs.benchmarks) < 1 {
		fmt.Println("No benchmark functions registered!")
		return
	}

	fmt.Println("Run..")

	for _, benchmark := range bs.benchmarks {
		fmt.Println(benchmark.name, benchmark.n, "iterations")
		for _, driver := range bs.drivers {
			fmt.Println(driver.name)
			res := benchmark.run(driver.db)
			if res.Err != nil {
				fmt.Println(res.Err.Error())
			} else {
				fmt.Println(res.Duration.String(), "   ", res.QueriesPerSecond())
			}
		}
		fmt.Println()
	}
}