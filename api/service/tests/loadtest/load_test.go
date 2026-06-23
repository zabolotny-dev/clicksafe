package loadtest_test

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

// SLO задаёт порог приемлемой работы: p99 ≤ 100 мс при успешности ≥ 99 %.
// Наибольшая ступень нагрузки, прошедшая SLO, считается «потолком» сценария.
const (
	sloP99     = 100 * time.Millisecond
	sloSuccess = 0.99
)

// rpsSteps — ступени нагрузки в запросах/сек. Грубый разгон до 2500, затем
// мелкая сетка с шагом 250 в зоне 3000–5000, где обычно происходит слом —
// чтобы локализовать точку насыщения точнее. Финальные 6000 — контрольная
// точка для подтверждения коллапса.
var rpsSteps = []uint64{
	500, 1500, 2500,
	3000, 3250, 3500, 3750,
	4000, 4250, 4500, 4750,
	5000, 6000,
}

const stepDuration = 15 * time.Second

// iterations — сколько раз прогоняется каждый сценарий. Шум одного прогона
// (фоновые процессы, тротлинг, GC, разогрев pgxpool) даёт разброс ±2× по
// потолку RPS. Медиана по 3 прогонам устраняет эту вариативность и даёт
// устойчивые цифры. Можно временно поставить 1 для быстрой итерации.
const iterations = 3

func TestLoadScenarios(t *testing.T) {
	fmt.Printf("\nSLO: p99 ≤ %d мс, успешность ≥ %.0f %%\n",
		sloP99/time.Millisecond, sloSuccess*100)

	// Каждый сценарий получает полностью свежий стек: новую БД (в том же
	// контейнере), новый HTTP-сервер, повторный сидинг. Это исключает
	// кросс-контаминацию: катастрофическая просадка одного сценария
	// (TIME_WAIT, bloat, истощённый пул) не отравляет измерения следующего.
	scenarios := []struct {
		name          string
		buildTargeter func(base string, sd loadSeedData) vegeta.Targeter
	}{
		{"Панель управления", func(base string, sd loadSeedData) vegeta.Targeter {
			return adminTargeter(base, sd.Admins[0].RawToken, sd.Admins[0].CSRFToken, sd)
		}},
		{"Фишинговая атака", func(base string, sd loadSeedData) vegeta.Targeter {
			return roundRobinTargeter(phishingTargets(base, sd.PhishingTokens))
		}},
		{"Трекинг событий", func(base string, sd loadSeedData) vegeta.Targeter {
			return roundRobinTargeter(eventTargets(base, sd.PhishingTokens))
		}},
	}

	perScenarioRuns := make(map[string][]scenarioResult, len(scenarios))
	for iter := 1; iter <= iterations; iter++ {
		fmt.Printf("\n══════════════ Итерация %d/%d ══════════════\n", iter, iterations)
		for _, sc := range scenarios {
			// Каждый сценарий — отдельный subtest, чтобы t.Cleanup() от dbtest.New
			// (DROP DATABASE, закрытие пула) отрабатывал сразу после прогона, а не
			// копился до конца всего TestLoadScenarios. Без этого PG-контейнер
			// захлёбывается через несколько итераций — все БД живут одновременно.
			var res scenarioResult
			subName := fmt.Sprintf("iter%d/%s", iter, sc.name)
			t.Run(subName, func(subT *testing.T) {
				res = runScenarioIsolated(subT, sc.name, sc.buildTargeter)
			})
			perScenarioRuns[sc.name] = append(perScenarioRuns[sc.name], res)
		}
	}

	results := make([]scenarioResult, 0, len(scenarios))
	for _, sc := range scenarios {
		res := aggregateMedian(sc.name, perScenarioRuns[sc.name])
		if res.capacity == 0 {
			t.Errorf("%s: медиана по %d прогонам не уложилась в SLO ни на одной ступени", sc.name, iterations)
		}
		results = append(results, res)
	}

	printMedianTables(results)
	printSummary(results)
	writeCSV(t, results)
}

// aggregateMedian берёт N прогонов одного сценария и для каждой ступени RPS
// считает медиану по mean/p95/p99/throughput/success. SLO проверяется по
// медианным значениям. Если ступень была достигнута не во всех прогонах
// (из-за раннего завершения), медиана берётся по тем, что есть.
func aggregateMedian(name string, runs []scenarioResult) scenarioResult {
	byRPS := make(map[uint64][]stepResult)
	for _, run := range runs {
		for _, s := range run.steps {
			byRPS[s.rps] = append(byRPS[s.rps], s)
		}
	}

	res := scenarioResult{name: name}
	for _, rps := range rpsSteps {
		steps := byRPS[rps]
		if len(steps) == 0 {
			continue
		}
		med := medianStep(rps, steps)
		res.steps = append(res.steps, med)
		if med.passedSLO {
			res.capacity = rps
		}
	}
	return res
}

func medianStep(rps uint64, steps []stepResult) stepResult {
	means := make([]time.Duration, len(steps))
	p95s := make([]time.Duration, len(steps))
	p99s := make([]time.Duration, len(steps))
	thrs := make([]float64, len(steps))
	successes := make([]float64, len(steps))
	var reqs uint64
	for i, s := range steps {
		means[i], p95s[i], p99s[i] = s.mean, s.p95, s.p99
		thrs[i], successes[i] = s.throughput, s.success
		reqs += s.requests
	}
	slices.Sort(means)
	slices.Sort(p95s)
	slices.Sort(p99s)
	slices.Sort(thrs)
	slices.Sort(successes)

	mid := len(steps) / 2
	p99 := p99s[mid]
	success := successes[mid]

	return stepResult{
		rps:        rps,
		requests:   reqs / uint64(len(steps)),
		mean:       means[mid],
		p95:        p95s[mid],
		p99:        p99,
		throughput: thrs[mid],
		success:    success,
		passedSLO:  p99 <= sloP99 && success >= sloSuccess,
	}
}

func printMedianTables(results []scenarioResult) {
	fmt.Printf("\n══════════════════ Медианные значения по %d прогонам ══════════════════\n", iterations)
	for _, r := range results {
		fmt.Printf("\n══ %s ══\n", r.name)
		fmt.Printf("%-6s %10s %10s %10s %10s %12s %10s %4s\n",
			"RPS", "Запросов", "Среднее", "p95", "p99", "Throughput", "Успешн.", "SLO")
		fmt.Println(strings.Repeat("─", 84))
		for _, s := range r.steps {
			mark := "✗"
			if s.passedSLO {
				mark = "✓"
			}
			fmt.Printf("%-6d %10d %8.2f мс %8.2f мс %8.2f мс %9.0f RPS %9.1f %% %4s\n",
				s.rps, s.requests,
				float64(s.mean)/float64(time.Millisecond),
				float64(s.p95)/float64(time.Millisecond),
				float64(s.p99)/float64(time.Millisecond),
				s.throughput, s.success*100, mark)
		}
		fmt.Printf("→ медианный потолок при SLO: %d RPS\n", r.capacity)
	}
}

// writeCSV сохраняет все измерения в loadtest_results.csv, чтобы построить
// «hockey stick»-график (p99 vs RPS в логарифмической шкале) в Excel/Sheets/
// gnuplot. Файл попадает в каталог теста — путь выводится в лог.
func writeCSV(t *testing.T, results []scenarioResult) {
	t.Helper()

	const fileName = "loadtest_results.csv"
	f, err := os.Create(fileName)
	if err != nil {
		t.Logf("csv create: %v", err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{
		"scenario", "rps", "requests",
		"mean_ms", "p95_ms", "p99_ms",
		"throughput", "success_pct", "passed_slo",
	})
	for _, r := range results {
		for _, s := range r.steps {
			_ = w.Write([]string{
				r.name,
				fmt.Sprintf("%d", s.rps),
				fmt.Sprintf("%d", s.requests),
				fmt.Sprintf("%.3f", float64(s.mean)/float64(time.Millisecond)),
				fmt.Sprintf("%.3f", float64(s.p95)/float64(time.Millisecond)),
				fmt.Sprintf("%.3f", float64(s.p99)/float64(time.Millisecond)),
				fmt.Sprintf("%.1f", s.throughput),
				fmt.Sprintf("%.2f", s.success*100),
				fmt.Sprintf("%v", s.passedSLO),
			})
		}
	}

	abs, _ := filepath.Abs(fileName)
	fmt.Printf("\nCSV: %s\n", abs)
}

func runScenarioIsolated(
	t *testing.T,
	name string,
	buildTargeter func(base string, sd loadSeedData) vegeta.Targeter,
) scenarioResult {
	// Форсированно удаляем контейнер от прошлого сценария, чтобы dbtest.New
	// поднял свежий PG-процесс. Без этого WAL/autovacuum/checkpoint от
	// предыдущего write-heavy сценария тащатся в следующий, и латентность
	// деградирует даже на низких RPS. Если контейнера нет — команда тихо
	// проваливается, что нормально для первого прогона.
	_ = exec.Command("docker", "rm", "-f", "clicksafetest").Run()

	test := apitest.New(t, "TestLoad")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("seed: %v", err)
		return scenarioResult{}
	}

	srv := httptest.NewServer(test.Mux)
	defer srv.Close()

	return runStepped(t, name, buildTargeter(srv.URL, sd))
}

type scenarioResult struct {
	name     string
	capacity uint64 // максимальный RPS, прошедший SLO
	steps    []stepResult
}

type stepResult struct {
	rps        uint64
	requests   uint64
	mean       time.Duration
	p95        time.Duration
	p99        time.Duration
	throughput float64
	success    float64
	passedSLO  bool
}

func runStepped(t *testing.T, name string, targeter vegeta.Targeter) scenarioResult {
	t.Helper()

	fmt.Printf("\n══ %s ══\n", name)
	fmt.Printf("%-6s %10s %10s %10s %10s %12s %10s %4s\n",
		"RPS", "Запросов", "Среднее", "p95", "p99", "Throughput", "Успешн.", "SLO")
	fmt.Println(strings.Repeat("─", 84))

	res := scenarioResult{name: name}

	for _, rps := range rpsSteps {
		step := runStep(targeter, rps, stepDuration)
		res.steps = append(res.steps, step)

		if step.passedSLO {
			res.capacity = rps
		}

		mark := "✗"
		if step.passedSLO {
			mark = "✓"
		}

		fmt.Printf("%-6d %10d %8.2f мс %8.2f мс %8.2f мс %9.0f RPS %9.1f %% %4s\n",
			rps, step.requests,
			float64(step.mean)/float64(time.Millisecond),
			float64(step.p95)/float64(time.Millisecond),
			float64(step.p99)/float64(time.Millisecond),
			step.throughput, step.success*100, mark)

		// Если просадка сильнее предела — дальше пушить бессмысленно.
		if step.success < 0.90 || step.p99 > 5*sloP99 {
			fmt.Println("  → деградация сильная, дальнейшие ступени пропускаются")
			break
		}

		time.Sleep(2 * time.Second)
	}

	if res.capacity == 0 {
		// В одиночной итерации SLO мог сорваться из-за шума прогона.
		// Финальный вердикт принимаем по медиане после агрегации.
		t.Logf("%s (одиночный прогон): ни одна ступень не уложилась в SLO", name)
	}
	fmt.Printf("→ потолок при SLO: %d RPS\n", res.capacity)
	return res
}

func runStep(targeter vegeta.Targeter, rps uint64, dur time.Duration) stepResult {
	att := vegeta.NewAttacker(
		vegeta.TLSConfig(&tls.Config{InsecureSkipVerify: true}),
		vegeta.Workers(rps/20+10),
		vegeta.Timeout(5*time.Second),
	)
	rate := vegeta.Rate{Freq: int(rps), Per: time.Second}

	var m vegeta.Metrics
	for r := range att.Attack(targeter, rate, dur, "step") {
		m.Add(r)
	}
	m.Close()

	return stepResult{
		rps:        rps,
		requests:   m.Requests,
		mean:       m.Latencies.Mean,
		p95:        m.Latencies.P95,
		p99:        m.Latencies.P99,
		throughput: m.Throughput,
		success:    m.Success,
		passedSLO:  m.Latencies.P99 <= sloP99 && m.Success >= sloSuccess,
	}
}

func printSummary(results []scenarioResult) {
	fmt.Printf("\n══════════════════ Итог ══════════════════\n")
	fmt.Printf("%-28s %12s\n", "Сценарий", "Потолок RPS")
	fmt.Println(strings.Repeat("─", 42))
	for _, r := range results {
		capStr := fmt.Sprintf("%d", r.capacity)
		if r.capacity == 0 {
			capStr = "—"
		}
		fmt.Printf("%-28s %12s\n", r.name, capStr)
	}
}

// ─── целевые наборы для каждого сценария ────────────────────────────────────

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (compatible; Outlook/16.0)",
	"Mozilla/5.0 (X11; Linux x86_64; rv:115.0) Gecko/20100101 Thunderbird/115.0",
}

var credPairs = []string{
	"username=admin&password=admin",
	"username=user&password=password123",
	"login=ivanov&password=Ivanov2024!",
	"email=petrov%40company.com&pass=qwerty123",
	"username=sidorov&password=Welcome1",
	"login=kozlov.a&password=P%40ssw0rd",
}

// adminTargeter имитирует типовую работу администратора: 65% чтений
// (листинги + per-entity запросы), 20% обновлений (PUT на существующих
// сотрудников, ротация по строкам — без hot-row contention), 15% созданий
// (POST с уникальным email через атомарный счётчик — обходим unique-констрейнт).
func adminTargeter(base, cookie, csrf string, sd loadSeedData) vegeta.Targeter {
	readHdr := http.Header{
		"Cookie":       []string{"__Host-session=" + cookie},
		"X-Csrf-Token": []string{csrf},
	}
	writeHdr := http.Header{
		"Cookie":       []string{"__Host-session=" + cookie},
		"X-Csrf-Token": []string{csrf},
		"Content-Type": []string{"application/json"},
	}

	// Pool чтений: листинги + per-entity запросы (кампании и первые 20 сотрудников).
	readURLs := []string{
		base + "/me",
		base + "/campaign",
		base + "/employee",
		base + "/message",
		base + "/landing",
		base + "/attachment",
		base + "/vtarget",
	}
	if len(sd.Campaigns) > 0 {
		readURLs = append(readURLs, base+"/campaign/"+sd.Campaigns[0].ID.String())
	}
	for _, id := range sd.EmployeeIDs[:min(20, len(sd.EmployeeIDs))] {
		readURLs = append(readURLs, base+"/employee/"+id.String())
	}

	updBodies := [][]byte{
		[]byte(`{"first_name":"Обновление"}`),
		[]byte(`{"last_name":"Изменённый"}`),
		[]byte(`{"first_name":"Проверка"}`),
		[]byte(`{"last_name":"Корпус"}`),
	}

	empCount := uint64(len(sd.EmployeeIDs))
	readCount := uint64(len(readURLs))

	var n atomic.Uint64
	return func(t *vegeta.Target) error {
		i := n.Add(1) - 1
		// 20 слотов: 13 чтений + 4 PUT + 3 POST = 65/20/15.
		slot := i % 20

		switch {
		case slot < 13:
			t.Method = "GET"
			t.URL = readURLs[(slot+i/20)%readCount]
			t.Header = readHdr
			t.Body = nil
		case slot < 17:
			t.Method = "PUT"
			t.URL = base + "/employee/" + sd.EmployeeIDs[(i/20+(slot-13))%empCount].String()
			t.Header = writeHdr
			t.Body = updBodies[(slot-13)%uint64(len(updBodies))]
		default:
			t.Method = "POST"
			t.URL = base + "/employee"
			t.Header = writeHdr
			t.Body = fmt.Appendf(nil,
				`{"first_name":"Нагр","last_name":"Тест","email":"loadtest-%d@company.com"}`, i,
			)
		}
		return nil
	}
}

// phishingTargets создаёт по одному GET-запросу на каждый токен.
// Vegeta ротирует по ним round-robin — каждый сотрудник проходит свой цикл.
// User-Agent ротируется по всему пулу для имитации разных почтовых клиентов.
func phishingTargets(base string, tokens []string) []vegeta.Target {
	targets := make([]vegeta.Target, len(tokens))
	for i, tok := range tokens {
		targets[i] = vegeta.Target{
			Method: "GET",
			URL:    base + "/" + tok,
			Header: http.Header{"User-Agent": []string{userAgents[i%len(userAgents)]}},
		}
	}
	return targets
}

// eventTargets чередует три типа событий по каждому токену:
// pixel (EMAIL_OPENED), visit (LINK_OPENED), submit (DATA_SENT).
// User-Agent и тело запроса ротируются — имитируют разных сотрудников
// с разными браузерами и паролями.
func eventTargets(base string, tokens []string) []vegeta.Target {
	targets := make([]vegeta.Target, 0, len(tokens)*3)
	for i, tok := range tokens {
		targets = append(targets,
			vegeta.Target{
				Method: "GET",
				URL:    base + "/" + tok + ".gif",
				Header: http.Header{"User-Agent": []string{userAgents[(i*3)%len(userAgents)]}},
			},
			vegeta.Target{
				Method: "GET",
				URL:    base + "/" + tok,
				Header: http.Header{"User-Agent": []string{userAgents[(i*3+1)%len(userAgents)]}},
			},
			vegeta.Target{
				Method: "POST",
				URL:    base + "/" + tok,
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
					"User-Agent":   []string{userAgents[(i*3+2)%len(userAgents)]},
				},
				Body: []byte(credPairs[i%len(credPairs)]),
			},
		)
	}
	return targets
}

func roundRobinTargeter(targets []vegeta.Target) vegeta.Targeter {
	n := uint64(len(targets))
	var counter atomic.Uint64
	return func(t *vegeta.Target) error {
		*t = targets[counter.Add(1)%n]
		return nil
	}
}
