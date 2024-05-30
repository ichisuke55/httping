package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"
	"time"
)

type Stat struct {
	counter      int
	successCount uint64
	failedCount  uint64
	rtts         []uint64
	destination  string
}

func NewStat(count int, dest string) *Stat {
	return &Stat{
		counter:      1,
		successCount: 0,
		failedCount:  0,
		rtts:         make([]uint64, count),
		destination:  dest,
	}
}

func (s *Stat) ping(ctx context.Context, wg *sync.WaitGroup, destination, method string, count int, interval float64, isRedirect, skipVerify bool) {
	limit := time.Duration(count) * time.Duration(int(interval*1000)) * time.Millisecond
	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}

	// custom redirect function handle
	if isRedirect {
		client.CheckRedirect = DisableRedirect
	}

	fmt.Printf("Destination is %s\n", destination)

	for begin := time.Now(); time.Since(begin) < limit; {
		select {
		case <-ctx.Done():
			defer wg.Done()
			return
		default:
			tr := &CustomTransport{
				dialer: &net.Dialer{
					Timeout: 10 * time.Second,
				},
			}
			tr.Transport = &http.Transport{
				DisableKeepAlives: true,
				Dial:              tr.dial,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipVerify,
				},
			}
			client.Transport = tr

			request, err := http.NewRequestWithContext(ctx, method, destination, nil)
			if err != nil {
				fmt.Printf("From %s: Sequence: %d, ErrReason: %v\n", destination, s.counter, err)
				s.counter++
				s.failedCount++
				time.Sleep(time.Duration(int(interval*1000)) * time.Millisecond)
				continue
			}
			resp, err := client.Do(request)
			if err != nil {
				fmt.Printf("From %s: Sequence: %d, ErrReason: %v\n", destination, s.counter, err)
				s.counter++
				s.failedCount++
				time.Sleep(time.Duration(int(interval*1000)) * time.Millisecond)
				continue
			}
			defer func() {
				if resp != nil {
					resp.Body.Close()
				}
			}()
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("From %s: Sequence: %d, ErrReason: %v\n", destination, s.counter, err)
				s.counter++
				s.failedCount++
				time.Sleep(time.Duration(int(interval*1000)) * time.Millisecond)
				continue
			}

			s.successCount++
			s.rtts[s.counter-1] = uint64(tr.Duration().Milliseconds())

			fmt.Printf("%d bytes from %s: Sequence: %d, StatusCode: %d, RTT: %dms\n", len(b), destination, s.counter, resp.StatusCode, tr.Duration().Milliseconds())

			s.counter++
			time.Sleep(time.Duration(int(interval*1000)) * time.Millisecond)
		}
	}

	defer wg.Done()
}

var (
	stdOutResultTemplate = `--- %s httping statistics ---
%d packet transmitted, %d received, %d%% success rate
rtt: min/avg/max = %d/%g/%d ms
`
)

func avg(s []uint64) float64 {
	var sum float64 = 0
	for _, v := range s {
		sum += float64(v)
	}
	return float64(sum) / float64(len(s))
}

func (s *Stat) printStat() {
	fmt.Printf(stdOutResultTemplate,
		s.destination, s.counter-1, s.successCount,
		s.successCount*100/(s.successCount+s.failedCount),
		slices.Min(s.rtts), avg(s.rtts), slices.Max(s.rtts))
}

func main() {
	dest := flag.String("d", "", "destination URL. e.g. 'http://localhost'")
	method := flag.String("X", "GET", "HTTP method: GET, POST")
	count := flag.Int("c", 5, "number of times execute")
	isRedirect := flag.Bool("r", false, "disable redirect (`bool`: default is false)")
	interval := flag.Float64("i", 1.0, "seconds between sending each httping request")
	skipVerify := flag.Bool("k", false, "skip SSL/TLS insecure verity (`bool`: default is false)")
	flag.Parse()

	s := NewStat(*count, *dest)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go s.ping(ctx, &wg, *dest, *method, *count, *interval, *isRedirect, *skipVerify)
	wg.Wait()

	// output result statistics
	s.printStat()
}
