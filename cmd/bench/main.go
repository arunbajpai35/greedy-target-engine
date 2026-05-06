package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android", "target url")
	conc := flag.Int("c", 50, "concurrent workers")
	dur := flag.Duration("d", 10*time.Second, "test duration")
	flag.Parse()

	client := &http.Client{Timeout: 5 * time.Second}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		samples  []time.Duration
		ok, fail int64
		stop     = make(chan struct{})
	)

	start := time.Now()
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make([]time.Duration, 0, 1024)
			for {
				select {
				case <-stop:
					mu.Lock()
					samples = append(samples, local...)
					mu.Unlock()
					return
				default:
				}
				t0 := time.Now()
				resp, err := client.Get(*url)
				lat := time.Since(t0)
				if err != nil {
					atomic.AddInt64(&fail, 1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode >= 500 {
					atomic.AddInt64(&fail, 1)
					continue
				}
				atomic.AddInt64(&ok, 1)
				local = append(local, lat)
			}
		}()
	}

	time.Sleep(*dur)
	close(stop)
	wg.Wait()
	elapsed := time.Since(start)

	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
	pct := func(p float64) time.Duration {
		if len(samples) == 0 {
			return 0
		}
		i := int(float64(len(samples)-1) * p)
		return samples[i]
	}

	fmt.Printf("url        %s\n", *url)
	fmt.Printf("workers    %d\n", *conc)
	fmt.Printf("duration   %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("requests   ok=%d fail=%d\n", ok, fail)
	fmt.Printf("rps        %.0f\n", float64(ok)/elapsed.Seconds())
	fmt.Printf("latency    p50=%s p95=%s p99=%s max=%s\n",
		pct(0.50).Round(time.Microsecond),
		pct(0.95).Round(time.Microsecond),
		pct(0.99).Round(time.Microsecond),
		pct(1.0).Round(time.Microsecond),
	)
}
