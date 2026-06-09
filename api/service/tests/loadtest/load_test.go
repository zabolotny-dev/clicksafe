package loadtest_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func TestLoadScenarios(t *testing.T) {
	test := apitest.New(t, "TestLoad")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	srv := httptest.NewServer(test.Mux)
	defer srv.Close()

	base   := srv.URL
	cookie := sd.Admins[0].RawToken
	csrf   := sd.Admins[0].CSRFToken
	tokens := sd.PhishingTokens

	t.Logf("server: %s  targets: %d", base, len(tokens))

	fmt.Printf("\n%-28s %5s %9s %10s %10s %10s %10s\n",
		"Сценарий", "RPS", "Запросов", "Среднее", "p95", "p99", "Успешность")
	fmt.Println(strings.Repeat("-", 87))

	runScenario(t, "Панель управления", adminTargets(base, cookie, csrf), 300, 30*time.Second)
	runScenario(t, "Фишинговая атака",  phishingTargets(base, tokens),    500, 30*time.Second)
	runScenario(t, "Трекинг событий",   eventTargets(base, tokens),       500, 30*time.Second)
}

func runScenario(t *testing.T, name string, targets []vegeta.Target, rps uint64, dur time.Duration) {
	t.Helper()

	attacker := vegeta.NewAttacker(vegeta.TLSConfig(&tls.Config{InsecureSkipVerify: true}))
	rate     := vegeta.Rate{Freq: int(rps), Per: time.Second}

	var metrics vegeta.Metrics
	for res := range attacker.Attack(roundRobinTargeter(targets), rate, dur, name) {
		metrics.Add(res)
	}
	metrics.Close()

	avg     := float64(metrics.Latencies.Mean) / float64(time.Millisecond)
	p95     := float64(metrics.Latencies.P95) / float64(time.Millisecond)
	p99     := float64(metrics.Latencies.P99) / float64(time.Millisecond)
	success := metrics.Success * 100

	fmt.Printf("%-28s %5d %9d %9.2f мс %9.2f мс %9.2f мс %9.0f%%\n",
		name, rps, metrics.Requests, avg, p95, p99, success)

	if success < 100 {
		t.Errorf("%s: успешность %.1f%% < 100%%", name, success)
	}
}

// ─── целевые наборы для каждого сценария ────────────────────────────────────

func adminTargets(base, cookie, csrf string) []vegeta.Target {
	endpoints := []string{
		"/me", "/campaign", "/employee", "/message", "/landing", "/attachment", "/vtarget",
	}
	targets := make([]vegeta.Target, len(endpoints))
	for i, ep := range endpoints {
		targets[i] = vegeta.Target{
			Method: "GET",
			URL:    base + ep,
			Header: http.Header{
				"Cookie":       []string{"__Host-session=" + cookie},
				"X-Csrf-Token": []string{csrf},
			},
		}
	}
	return targets
}

// phishingTargets создаёт по одному GET-запросу на каждый токен.
// Vegeta ротирует по ним round-robin — каждый сотрудник проходит свой цикл.
func phishingTargets(base string, tokens []string) []vegeta.Target {
	targets := make([]vegeta.Target, len(tokens))
	for i, tok := range tokens {
		targets[i] = vegeta.Target{
			Method: "GET",
			URL:    base + "/" + tok,
			Header: http.Header{"User-Agent": []string{"Mozilla/5.0"}},
		}
	}
	return targets
}

// eventTargets чередует три типа событий по каждому токену:
// pixel (EMAIL_OPENED), visit (LINK_OPENED), submit (DATA_SENT).
func eventTargets(base string, tokens []string) []vegeta.Target {
	targets := make([]vegeta.Target, 0, len(tokens)*3)
	for _, tok := range tokens {
		targets = append(targets,
			vegeta.Target{
				Method: "GET",
				URL:    base + "/" + tok + ".gif",
				Header: http.Header{"User-Agent": []string{"Mozilla/5.0"}},
			},
			vegeta.Target{
				Method: "GET",
				URL:    base + "/" + tok,
				Header: http.Header{"User-Agent": []string{"Mozilla/5.0"}},
			},
			vegeta.Target{
				Method: "POST",
				URL:    base + "/" + tok,
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
					"User-Agent":   []string{"Mozilla/5.0"},
				},
				Body: []byte("username=test&password=test"),
			},
		)
	}
	return targets
}

func roundRobinTargeter(targets []vegeta.Target) vegeta.Targeter {
	i := 0
	return func(t *vegeta.Target) error {
		*t = targets[i%len(targets)]
		i++
		return nil
	}
}
